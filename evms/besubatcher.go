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
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

// BesuBatchVM is s Evm-interface wrapper around the `evmtool` binary, based on Besu.
// The BatchVM spins up one 'master' instance of the VM, and uses that to execute tests
type BesuBatchVM struct {
	BesuVM
	cmd    *exec.Cmd // the 'master' process
	stdout io.ReadCloser
	stdin  io.WriteCloser
	mu     sync.Mutex
}

func NewBesuBatchVM(path, name string) Evm {
	return &BesuBatchVM{
		BesuVM: BesuVM{
			path:  path,
			name:  name,
			stats: new(VMStat),
		},
	}
}

func (evm *BesuBatchVM) Instance(threadID int) Evm {
	return &BesuBatchVM{
		BesuVM: BesuVM{
			path:  evm.path,
			name:  fmt.Sprintf("%v-%d", evm.name, threadID),
			stats: evm.stats,
		},
	}
}

// RunStateTest implements the Evm interface
func (evm *BesuBatchVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		err    error
		cmd    *exec.Cmd
		stdout io.ReadCloser
		stdin  io.WriteCloser
	)
	if evm.cmd == nil {
		if speedTest {
			cmd = exec.Command(evm.path, "--nomemory", "--notime", "state-test")
		} else {
			cmd = exec.Command(evm.path, "--nomemory", "--notime", "--json", "state-test")
		}
		if stdout, err = cmd.StdoutPipe(); err != nil {
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
	_, _ = fmt.Fprintf(evm.stdin, "%v\n", path)

	// copy everything for the _current_ statetest to the given writer
	evm.copyUntilEnd(out, evm.stdout)
	duration, slow := evm.stats.TraceDone(t0)
	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      evm.cmd.String(),
	}, nil
}

func (evm *BesuBatchVM) Close() {
	if evm.stdin != nil {
		evm.stdin.Close()
	}
	if evm.cmd != nil {
		_ = evm.cmd.Wait()
	}
}

func (evm *BesuBatchVM) GetStateRoot(path string) (root, command string, err error) {
	if evm.cmd == nil {
		evm.cmd = exec.Command(evm.path, "--nomemory", "--notime", "state-test")
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
	_, _ = fmt.Fprintf(evm.stdin, "%v\n", path)
	sRoot := evm.copyUntilEnd(io.Discard, evm.stdout)
	return sRoot.StateRoot, evm.cmd.String(), nil
}
