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
	"time"
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
		NethermindVM: *NewNethermindVM(path, name),
	}
}

func (evm *NethermindBatchVM) Instance(threadId int) Evm {
	return &NethermindBatchVM{
		NethermindVM: NethermindVM{
			path:  evm.path,
			name:  fmt.Sprintf("%v-%d", evm.name, threadId),
			stats: evm.stats,
		},
	}
}

// RunStateTest implements the Evm interface
func (evm *NethermindBatchVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		err    error
		stdout io.ReadCloser
		stdin  io.WriteCloser
	)
	if evm.cmd == nil {
		cmd := exec.Command(evm.path, "-x", "--trace", "-m")
		if speedTest {
			cmd = exec.Command(evm.path, "-x", "--trace", "-m", "--neverTrace")
		}
		if stdout, err = cmd.StderrPipe(); err != nil {
			return &tracingResult{Cmd: cmd.String()}, err
		}
		if stdin, err = cmd.StdinPipe(); err != nil {
			return &tracingResult{Cmd: cmd.String()}, err
		}
		if err = cmd.Start(); err != nil {
			return &tracingResult{Cmd: cmd.String()}, err
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
	duration, slow := evm.stats.TraceDone(t0)
	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      evm.cmd.String(),
	}, nil
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
	sRoot := evm.copyUntilEnd(io.Discard, evm.stdout)
	return sRoot.StateRoot, evm.cmd.String(), nil
}
