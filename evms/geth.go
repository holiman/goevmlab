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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/log"
)

// GethEVM is s Evm-interface wrapper around the `evm` binary, based on go-ethereum.
type GethEVM struct {
	path string
	name string // in case multiple instances are used
}

func NewGethEVM(path string, name string) *GethEVM {
	return &GethEVM{
		path: path,
		name: name,
	}
}

func (evm *GethEVM) Name() string {
	return fmt.Sprintf("geth-%v", evm.name)
}

// GetStateRoot runs the test and returns the stateroot
// This currently only works for non-filled statetests. TODO: make it work even if the
// test is filled. Either by getting the whole trace, or adding stateroot to exec std output
// even in success-case
func (evm *GethEVM) GetStateRoot(path string) (root, command string, err error) {
	// In this mode, we can run it without tracing
	cmd := exec.Command(evm.path, "statetest", path)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", cmd.String(), err
	}
	root, err = evm.ParseStateRoot(data)
	if err != nil {
		log.Error("Failed to find stateroot", "vm", evm.Name(), "cmd", cmd.String())
		return "", cmd.String(), err
	}
	return root, cmd.String(), err
}

// ParseStateRoot reads geth's stateroot from the combined output.
func (evm *GethEVM) ParseStateRoot(data []byte) (string, error) {
	start := bytes.Index(data, []byte(`"stateRoot": "`))
	end := start + 14 + 66
	if start == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start+14 : end]), nil
}

// RunStateTest implements the Evm interface
func (evm *GethEVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
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

func (vm *GethEVM) Close() {
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *GethEVM) Copy(out io.Writer, input io.Reader) {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	// Start with 1MB buffer, allow up to 32 MB
	scanner.Buffer(make([]byte, 1024*1024), 32*1024*1024)

	// When geth encounters an error, it may already have spat out the info, prematurely.
	// We need to merge it back to one item
	// https://github.com/ethereum/go-ethereum/pull/23970#issuecomment-979851712
	var prev *logger.StructLog
	var yield = func(current *logger.StructLog) {
		if prev == nil {
			prev = current
			return
		}
		jsondata, _ := json.Marshal(prev)
		if _, err := out.Write(append(jsondata, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
		if current == nil { // final flush
			return
		}
		if prev.Pc == current.Pc && prev.Depth == current.Depth {
			// Yup, that happened here. Set the error and continue
			prev = nil
		} else {
			prev = current
		}
	}

	for scanner.Scan() {
		data := scanner.Bytes()
		if len(data) > 1 && data[0] == '#' {
			// Output preceded by # is ignored, but can be used for debugging, e.g.
			// to check that the generated tests cover the intended surface.
			fmt.Printf("%v: %v\n", evm.Name(), string(data))
			continue
		}
		var elem logger.StructLog
		if err := json.Unmarshal(data, &elem); err != nil {
			fmt.Printf("geth err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			/* It might be the stateroot
			{"output":"","gasUsed":"0x2d1cc4","error":"gas uint64 overflow"}
			{"stateRoot": "0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134"}
			*/
			if stateRoot.StateRoot == "" {
				_ = json.Unmarshal(data, &stateRoot)
			}
			continue
		}
		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}

		RemoveUnsupportedElems(&elem)
		yield(&elem)
	}
	yield(nil)
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		return
	}
}
