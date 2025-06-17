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

// Package evms contains various tooling for interacting with different evms.
package evms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

// BesuVM is s Evm-interface wrapper around the `evmtool` binary, based on Besu.
type BesuVM struct {
	path string
	name string // in case multiple instances are used
	// Some metrics
	stats *VMStat
}

func NewBesuVM(path, name string) Evm {
	return &BesuVM{
		path:  path,
		name:  name,
		stats: new(VMStat),
	}
}

func (evm *BesuVM) Instance(int) Evm {
	return evm
}

func (evm *BesuVM) Name() string {
	return evm.name
}

// RunStateTest implements the Evm interface
func (evm *BesuVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		stdout io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	if speedTest {
		cmd = exec.Command(evm.path, "--nomemory", "--notime", "state-test", path)
	} else {
		cmd = exec.Command(evm.path, "--nomemory", "--notime", "--json", "state-test", path) // exclude memory
	}
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	if err = cmd.Start(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	// copy everything to the given writer
	evm.Copy(out, stdout)
	_, _ = io.ReadAll(stdout)
	err = cmd.Wait()
	// release resources
	duration, slow := evm.stats.TraceDone(t0)

	return &tracingResult{
			Slow:     slow,
			ExecTime: duration,
			Cmd:      cmd.String()},
		err
}

func (evm *BesuVM) Close() {}

func (evm *BesuVM) GetStateRoot(path string) (root, command string, err error) {
	// Run without tracing
	cmd := exec.Command(evm.path, "--nomemory", "--notime", "state-test", path)

	data, err := cmd.Output()
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

// ParseStateRoot reads the stateroot from the combined output.
func (evm *BesuVM) ParseStateRoot(data []byte) (string, error) {
	start := strings.Index(string(data), `"postHash":"`)
	if start > 0 {
		start = start + len(`"postHash":"`)
		root := string(data[start : start+2+64])
		return root, nil
	}
	start = strings.Index(string(data), `"stateRoot":"`)
	if start > 0 {
		start = start + len(`"stateRoot":"`)
		root := string(data[start : start+2+64])
		return root, nil
	}
	return "", errors.New("besu: no stateroot/posthash found")
}

// Copy feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *BesuVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input)
}

func (evm *BesuVM) copyUntilEnd(out io.Writer, input io.Reader) stateRoot {
	scanner := NewJsonlScanner("besu", input, os.Stderr)
	defer scanner.Release()
	var stateRoot stateRoot
	var elem opLog
	for scanner.Next(&elem) == nil {
		// If we have a stateroot, we're done
		if len(elem.StateRoot1) != 0 {
			stateRoot.StateRoot = elem.StateRoot1
			break
		}
		if len(elem.StateRoot2) != 0 {
			stateRoot.StateRoot = elem.StateRoot2
			break
		}
		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}
		outp := CustomMarshal(&elem)
		if _, err := out.Write(append(outp, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
			return stateRoot
		}
		elem.FunctionDepth = 0 // function depth is optional and "gets dirty" if not set
		elem.Section = 0
	}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
	return stateRoot
}

func (evm *BesuVM) Stats() []any {
	return evm.stats.Stats()
}
