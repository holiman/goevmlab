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

	"github.com/ethereum/go-ethereum/core/vm"
)

// AlethVM is s Evm-interface wrapper around the `testeth` binary, based on Aleth.
type AlethVM struct {
	path string
}

func NewAlethVM(path string) *AlethVM {
	return &AlethVM{
		path: path,
	}
}

// RunStateTest implements the Evm interface
func (evm *AlethVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		stderr io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	// ../../testeth -t GeneralStateTests --  --testfile ./statetest1.json --jsontrace {} 2> statetest1_testeth_stderr.jsonl
	if speedTest {
		cmd = exec.Command(evm.path, "-t", "GeneralStateTests",
			"--", "--testfile", path)
	} else {
		cmd = exec.Command(evm.path, "-t", "GeneralStateTests",
			"--", "--testfile", path, "--jsontrace", "{\"disableMemory\":true}")
	}
	if stderr, err = cmd.StdoutPipe(); err != nil {
		return cmd.String(), err
	}
	if err = cmd.Start(); err != nil {
		return cmd.String(), err
	}
	// copy everything to the given writer
	evm.Copy(out, stderr)
	// release resources
	_ = cmd.Wait()
	return cmd.String(), nil
}

func (evm *AlethVM) Name() string {
	return "alethvm"
}

// GetStateRoot runs the test and returns the stateroot
func (evm *AlethVM) GetStateRoot(path string) (string, error) {
	return "", errors.New("not implemented for testeth/aleth")
}

func (vm *AlethVM) Close() {
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *AlethVM) Copy(out io.Writer, input io.Reader) {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	buf := make([]byte, 4*1024*1024)
	scanner.Buffer(buf, cap(buf))
	for scanner.Scan() {
		// Calling bytes means that bytes in 'l' will be overwritten
		// in the next loop. Fine for now though, we immediately marshal it
		data := scanner.Bytes()
		var elem vm.StructLog
		if err := json.Unmarshal(data, &elem); err != nil {
			//fmt.Printf("aleth err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			/*  Most likely one of these:
			{"output":"","gasUsed":"0x2d1cc4","time":233624,"error":"gas uint64 overflow"}
			{"stateRoot": "a2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134"}
			*/
			// Don't overwrite stateroot if we already have it
			if stateRoot.StateRoot == "" {
				_ = json.Unmarshal(data, &stateRoot)
				if stateRoot.StateRoot != "" {
					// Aleth doesn't prefix stateroot
					stateRoot.StateRoot = fmt.Sprintf("0x%v", stateRoot.StateRoot)
				}
			}
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
