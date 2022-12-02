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
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"
)

// BesuBatchVM is s Evm-interface wrapper around the `evmtool` binary, based on Besu.
// The BatchVM spins up one 'master' instance of the VM, and uses that to execute tests
type BesuBatchVM struct {
	path   string
	cmd    *exec.Cmd // the 'master' process
	stdout io.ReadCloser
	stdin  io.WriteCloser
	mu     sync.Mutex
	name   string // in case multiple instances are used
}

func NewBesuBatchVM(path, name string) *BesuBatchVM {
	return &BesuBatchVM{
		path: path,
		name: name,
	}
}

func (evm *BesuBatchVM) Name() string {
	return fmt.Sprintf("besubatch-%v", evm.name)
}

// RunStateTest implements the Evm interface
func (evm *BesuBatchVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		err    error
		cmd    *exec.Cmd
		stdout io.ReadCloser
		stdin  io.WriteCloser
	)
	if evm.cmd == nil {
		if speedTest {
			cmd = exec.Command(evm.path, "--nomemory", "state-test")
		} else {
			cmd = exec.Command(evm.path, "--nomemory", "--json", "state-test")
		}
		if stdout, err = cmd.StdoutPipe(); err != nil {
			return cmd.String(), err
		}
		if stdin, err = cmd.StdinPipe(); err != nil {
			return cmd.String(), err
		}
		if err = cmd.Start(); err != nil {
			return cmd.String(), err
		}
		evm.cmd = cmd
		evm.stdout = stdout
		evm.stdin = stdin
	}
	evm.mu.Lock()
	defer evm.mu.Unlock()
	_, _ = evm.stdin.Write([]byte(fmt.Sprintf("%v\n", path)))
	// copy everything for the _current_ statetest to the given writer
	evm.copyUntilEnd(out, evm.stdout)
	// release resources, handle error but ignore non-zero exit codes
	return evm.cmd.String(), nil
}

func (vm *BesuBatchVM) Close() {
	if vm.stdin != nil {
		vm.stdin.Close()
	}
	if vm.cmd != nil {
		_ = vm.cmd.Wait()
	}
}

func (evm *BesuBatchVM) GetStateRoot(path string) (root, command string, err error) {
	if evm.cmd == nil {
		evm.cmd = exec.Command(evm.path, "--nomemory", "state-test")
		// The stateroot is delivered on stdout
		if evm.stdout, err = evm.cmd.StdoutPipe(); err != nil {
			return "", evm.cmd.String(), err
		}
		if evm.stdin, err = evm.cmd.StdinPipe(); err != nil {
			return "", evm.cmd.String(), err
		}
		if err = evm.cmd.Start(); err != nil {
			return "", evm.cmd.String(), err
		}
	}
	evm.mu.Lock()
	defer evm.mu.Unlock()
	_, _ = evm.stdin.Write([]byte(fmt.Sprintf("%v\n", path)))
	sRoot := evm.copyUntilEnd(devNull{}, evm.stdout)
	return sRoot.StateRoot, evm.cmd.String(), nil

}

type devNull struct{}

func (d devNull) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Copy feed reads from the reader, does some client-specific filtering and
// outputs BesuBatchVM onto the channel
func (evm *BesuBatchVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input)
}

func (evm *BesuBatchVM) copyUntilEnd(out io.Writer, input io.Reader) stateRoot {

	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	// We use a larger scanner buffer for besu: it does not have a way to
	// disable 'returndata', which can become larger than fits into a default
	// scanner buffer
	buf := make([]byte, 16*1024*1024)
	scanner.Buffer(buf, cap(buf))
	for scanner.Scan() {
		data := scanner.Bytes()
		var elem logger.StructLog
		// Besu (like Nethermind) sometimes report a negative refund
		//if i := bytes.Index(data, []byte(`"refund":-`)); i > 0 {
		//	// we can just make it positive, it will be zeroed later
		//	data[i+9] = byte(' ')
		//}

		err := json.Unmarshal(data, &elem)
		if err != nil {
			fmt.Printf("besu err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			if stateRoot.StateRoot == "" {
				var tempRoot besuStateRoot
				if err := json.Unmarshal(data, &tempRoot); err == nil {
					// Besu calls state root "postHash"
					stateRoot.StateRoot = tempRoot.StateRoot
				}
			}
			// If we have a stateroot, we're done
			break
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
			return stateRoot
		}
	}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
	return stateRoot
}
