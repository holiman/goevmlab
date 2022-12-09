// Copyright 2022 Martin Holst Swende
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
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// The NethermindBatchVM spins up one 'master' instance of the VM, and uses that to execute tests
type NethermindBatchVM struct {
	NethermindVM
	cmd    *exec.Cmd // the 'master' process
	stdout io.ReadCloser
	stdin  io.WriteCloser
	mu     sync.Mutex
}

func NewNethermindBatchVM(path, name string) *NethermindBatchVM {
	return &NethermindBatchVM{
		NethermindVM: NethermindVM{
			path: path,
			name: name,
		},
	}
}

func (evm *NethermindBatchVM) Name() string {
	return fmt.Sprintf("nethbatch-%v", evm.name)
}

// RunStateTest implements the Evm interface
func (evm *NethermindBatchVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		err    error
		cmd    *exec.Cmd
		stdout io.ReadCloser
		stdin  io.WriteCloser
	)
	if evm.cmd == nil {
		if speedTest {
			cmd = exec.Command(evm.path, "-x", "--trace", "-m", "--neverTrace")
		} else {
			cmd = exec.Command(evm.path, "-x", "--trace", "-m")
		}
		if stdout, err = cmd.StderrPipe(); err != nil {
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

func (vm *NethermindBatchVM) Close() {
	if vm.stdin != nil {
		vm.stdin.Close()
	}
	if vm.cmd != nil {
		_ = vm.cmd.Wait()
	}
}

func (evm *NethermindBatchVM) GetStateRoot(path string) (root, command string, err error) {
	if evm.cmd == nil {
		evm.cmd = exec.Command(evm.path, "--neverTrace", "-m", "-s", "-x")
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
