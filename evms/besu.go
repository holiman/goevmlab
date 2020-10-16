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

	"github.com/ethereum/go-ethereum/core/vm"
)

// BesuVM is s Evm-interface wrapper around the `evmtool` binary, based on Besu.
type BesuVM struct {
	path string
}

func NewBesuVM(path string) *BesuVM {
	return &BesuVM{
		path: path,
	}
}

// RunStateTest implements the Evm interface
func (evm *BesuVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		stdout io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	if speedTest {
		cmd = exec.Command(evm.path, "--nomemory", "state-test", path)
	} else {
		// evm --nomemory --json state-test blaketest.json
		cmd = exec.Command(evm.path, "--nomemory", "--json", "state-test", path) // exclude memory
	}

	if stdout, err = cmd.StdoutPipe(); err != nil {
		return cmd.String(), err
	}
	if err = cmd.Start(); err != nil {
		return cmd.String(), err
	}
	// copy everything to the given writer
	evm.Copy(out, stdout)
	// release resources, handle error but ignore non-zero exit codes
	_ = cmd.Wait()
	return cmd.String(), nil
}

func (evm *BesuVM) Name() string {
	return "besu"
}

func (vm *BesuVM) Close() {
}

func (vm *BesuVM) GetStateRoot(path string) (string, error) {
	return "", nil
}

type besuStateRoot struct {
	StateRoot string `json:"postHash"`
}

// feed reads from the reader, does some besu-specific filtering and
// outputs items onto the channel
func (evm *BesuVM) Copy(out io.Writer, input io.Reader) {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		data := scanner.Bytes()
		var elem vm.StructLog
		err := json.Unmarshal(data, &elem)
		if err != nil {
			fmt.Printf("besu err: %v, line\n\t%v\n", err, string(data))
			continue
		}

		elem.Memory = make([]byte, 0)
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			if stateRoot.StateRoot == "" {
				var tempRoot besuStateRoot
				if err := json.Unmarshal(data, &tempRoot); err == nil {
					// Besu calls state root "rootHash"
					stateRoot.StateRoot = tempRoot.StateRoot
				}
			}
			//fmt.Printf("%v\n", string(data))
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

func (evm *BesuVM) RunStateTestBatch(paths []string) ([][]byte, error) {
	var (
		stdout io.ReadCloser
		err    error
		cmd    *exec.Cmd
		out    = make([][]byte, len(paths))
		buffer bytes.Buffer
		i      int
		args   = make([]string, 0, 3)
	)
	// Arguments have to be appended this way otherwise variadic doesn't work.
	args = append(args, "--nomemory")
	args = append(args, "--json")
	args = append(args, "state-test")
	args = append(args, paths...)

	cmd = exec.Command(evm.path, args...) // exclude memory

	if stdout, err = cmd.StdoutPipe(); err != nil {
		return out, err
	}
	if err = cmd.Start(); err != nil {
		return out, err
	}
	// Scan the stdout
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		data := scanner.Bytes()
		var elem vm.StructLog
		err := json.Unmarshal(data, &elem)
		if err != nil {
			fmt.Printf("besu err: %v, line\n\t%v\n", err, string(data))
			continue
		}

		RemoveUnsupportedElems(&elem)
		elem.Memory = make([]byte, 0)
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			var (
				stateRoot stateRoot
				tempRoot  besuStateRoot
			)
			if err := json.Unmarshal(data, &tempRoot); err == nil {
				// Besu calls state root "rootHash"
				stateRoot.StateRoot = tempRoot.StateRoot
			}
			root, _ := json.Marshal(stateRoot)
			if _, err := buffer.Write(append(root, '\n')); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
				return out, err
			}
			// The state root is the last we'll read of a test.
			// Now save the buffer and parse the next test output.
			out[i] = buffer.Bytes()
			buffer.Reset()
			i++
			// We read enough...
			if i >= len(paths) {
				break
			}
		}

		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}
		jsondata, _ := json.Marshal(elem)
		if _, err := buffer.Write(append(jsondata, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
			return out, err
		}
	}

	// release resources, handle error but ignore non-zero exit codes
	_ = cmd.Wait()
	return out, nil
}
