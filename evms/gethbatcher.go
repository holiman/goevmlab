// Copyright 2023 Martin Holst Swende
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

// The GethBatchVM spins up one 'master' instance of the VM, and uses that to execute tests
type GethBatchVM struct {
	GethEVM
	cmd    *exec.Cmd // the 'master' process
	stdout io.ReadCloser
	stdin  io.WriteCloser
	mu     sync.Mutex

	// The 'master' process for blocktests
	bt *batchProc
}

func NewGethBatchVM(path, name string) Evm {
	return &GethBatchVM{
		GethEVM: GethEVM{path, name, &VMStat{}},
	}
}

func (evm *GethBatchVM) Instance(threadID int) Evm {
	return &GethBatchVM{
		GethEVM: GethEVM{
			path:  evm.path,
			name:  fmt.Sprintf("%v-%d", evm.name, threadID),
			stats: evm.stats,
		},
	}
}

// RunStateTest implements the Evm interface
func (evm *GethBatchVM) RunStateTest(path string, out io.Writer, speedTest bool) (*TracingResult, error) {
	var (
		t0     = time.Now()
		err    error
		cmd    *exec.Cmd
		stdout io.ReadCloser
		stdin  io.WriteCloser
	)
	if evm.cmd == nil {
		if speedTest {
			//cmd = exec.Command(evm.path, "--nomemory", "--noreturndata", "--nostack", "statetest")
			cmd = exec.Command(evm.path, "statetest")
		} else {
			//cmd = exec.Command(evm.path, "--json", "--noreturndata", "--nomemory", "statetest")
			cmd = exec.Command(evm.path, "statetest", "--trace", "--trace.format=json",
				"--trace.nomemory=true", "--trace.noreturndata=true")

		}
		if stdout, err = cmd.StderrPipe(); err != nil {
			return &TracingResult{Cmd: cmd.String()}, err
		}
		if stdin, err = cmd.StdinPipe(); err != nil {
			return &TracingResult{Cmd: cmd.String()}, err
		}
		if err = cmd.Start(); err != nil {
			return &TracingResult{Cmd: cmd.String()}, err
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
	// release resources, handle error but ignore non-zero exit codes
	duration, slow := evm.stats.TraceDone(t0)
	return &TracingResult{
			Slow:     slow,
			ExecTime: duration,
			Cmd:      evm.cmd.String()},
		nil
}

func (evm *GethBatchVM) Close() {
	if evm.stdin != nil {
		evm.stdin.Close()
	}
	if evm.cmd != nil {
		_ = evm.cmd.Wait()
	}
	evm.bt.close()
}

// RunBlockTest implements the Evm interface. It uses a persistent process,
// which reads the test file paths from stdin. In traced mode, the trace stream
// must contain an end marker per test; otherwise the test boundary cannot be
// determined. Older evm binaries lack the marker, which is detected by a
// one-shot probe on the first test.
func (evm *GethBatchVM) RunBlockTest(path string, out io.Writer, mode BlockTestMode) (*TracingResult, error) {
	var (
		t0  = time.Now()
		err error
	)
	if mode.Engine && !evm.supportsEngineTests() {
		return &TracingResult{}, fmt.Errorf("%w: %v cannot run engine fixtures", ErrBlockTestUnsupported, evm.path)
	}
	evm.mu.Lock()
	defer evm.mu.Unlock()
	if !evm.bt.reusable(mode) {
		evm.bt.close()
		if mode.Trace {
			// Probe for the end marker
			probe := exec.Command(evm.path, evm.blockTestArgs(mode, path)...)
			stderr, err := probe.StderrPipe()
			if err != nil {
				return &TracingResult{Cmd: probe.String()}, err
			}
			if err := probe.Start(); err != nil {
				return &TracingResult{Cmd: probe.String()}, err
			}
			end := copyBlockTestTrace("geth", io.Discard, stderr, true)
			_, _ = io.ReadAll(stderr)
			_ = probe.Wait()
			if end == nil {
				return &TracingResult{Cmd: probe.String()},
					fmt.Errorf("%w: %v emits no blocktest end marker, use --geth instead of --gethbatch", ErrBlockTestUnsupported, evm.path)
			}
		}
		if evm.bt, err = startBatchProc(evm.path, evm.blockTestArgs(mode, ""), mode); err != nil {
			return &TracingResult{Cmd: evm.bt.String()}, err
		}
	}
	end, err := evm.bt.run(path, func(input io.Reader) *blockTestEnd {
		return copyBlockTestTrace("geth", out, input, true)
	})
	writeBlockTestEnd(out, end)
	duration, slow := evm.stats.TraceDone(t0)
	return &TracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      evm.bt.String(),
	}, err
}

func (evm *GethBatchVM) GetStateRoot(path string) (root, command string, err error) {
	if evm.cmd == nil {
		//evm.cmd = exec.Command(evm.path, "--nomemory", "--noreturndata", "--nostack", "statetest")
		evm.cmd = exec.Command(evm.path, "statetest")
		if evm.stdout, err = evm.cmd.StderrPipe(); err != nil {
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
