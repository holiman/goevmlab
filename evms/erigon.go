// Copyright 2019 Martin Holst Swende
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the goevmlab library. If not, see <http://www.gnu.org/licenses/>.

package evms

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"
)

// ErigonVM is s Evm-interface wrapper around the eroigon `evm` binary
type ErigonVM struct {
	path string
}

func NewErigonVM(path string) *ErigonVM {
	return &ErigonVM{
		path: path,
	}
}

// GetStateRoot runs the test and returns the stateroot
// This currently only works for non-filled statetests. TODO: make it work even if the
// test is filled. Either by getting the whole trace, or adding stateroot to exec std output
// even in success-case
func (evm *ErigonVM) GetStateRoot(path string) (string, error) {
	// In this mode, we can run it without tracing
	cmd := exec.Command(evm.path, "statetest", path)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	start := strings.Index(string(data), "mismatch: got ")
	end := strings.Index(string(data), ", want")
	if start > 0 && end > 0 {
		root := fmt.Sprintf("0x%v", string(data[start+len("mismatch: got "):end]))
		return root, nil
	}
	return "", errors.New("no stateroot found")
}

// RunStateTest implements the Evm interface
func (evm *ErigonVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		stderr io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	if speedTest {
		cmd = exec.Command(evm.path, "--nomemory", "--noreturndata", "--nostack", "statetest", path)
	} else {
		cmd = exec.Command(evm.path, "--json", "--noreturndata", "--nomemory", "statetest", path)
	}
	if stderr, err = cmd.StderrPipe(); err != nil {
		return cmd.String(), err
	}
	if err = cmd.Start(); err != nil {
		return cmd.String(), err
	}
	// copy everything to the given writer
	evm.Copy(out, stderr)
	// release resources
	return cmd.String(), cmd.Wait()
}

func (evm *ErigonVM) Name() string {
	return "erigon"
}

func (vm *ErigonVM) Close() {
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *ErigonVM) Copy(out io.Writer, input io.Reader) {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	buf := make([]byte, 4*1024*1024)
	scanner.Buffer(buf, cap(buf))
	for scanner.Scan() {
		data := scanner.Bytes()
		var elem logger.StructLog
		err := json.Unmarshal(data, &elem)
		if err != nil {
			fmt.Printf("erigon err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			if stateRoot.StateRoot == "" {
				if err := json.Unmarshal(data, &stateRoot); err == nil {
					// geth doesn't 0x-prefix stateroot
					if r := stateRoot.StateRoot; len(r) > 0 {
						stateRoot.StateRoot = fmt.Sprintf("0x%v", r)
					}
				}
			}

			/*  Most likely one of these:
			{"output":"","gasUsed":"0x2d1cc4","time":233624,"error":"gas uint64 overflow"}
			{"stateRoot": "a2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134"}
			*/
			// For now, just ignore these
			continue
		}
		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}
		RemoveUnsupportedElems(&elem)

		jsondata, _ := json.Marshal(elem)
		if _, err := out.Write(append(jsondata, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
			return
		}
	}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		return
	}
}
