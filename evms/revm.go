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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

type RethVM struct {
	path string
	name string

	stats *VMStat
}

func NewRethVM(path string, name string) Evm {
	return &RethVM{
		path:  path,
		name:  name,
		stats: &VMStat{},
	}
}

func (evm *RethVM) Instance(int) Evm {
	return evm
}

func (evm *RethVM) Name() string {
	return evm.name
}

func (evm *RethVM) GetStateRoot(path string) (root, command string, err error) {
	cmd := exec.Command(evm.path, "statetest", "--json-outcome", path)
	data, err := StdErrOutput(cmd)

	// If revm exits with 1 on stateroot errors, uncomment to ignore:
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		err = nil
	}
	if err != nil {
		return "", cmd.String(), err
	}

	root, err = evm.ParseStateRoot(data)
	if err != nil {
		log.Error("Failed to find stateroot", "vm", evm.Name(), "cmd", cmd.String())
		return "", cmd.String(), err
	}
	return root, cmd.String(), nil
}

func (evm *RethVM) ParseStateRoot(data []byte) (root string, err error) {
	pattern := []byte(`"stateRoot":"`)
	idx := bytes.Index(data, pattern)
	start := idx + len(pattern)
	end := start + 32*2 + 2
	if idx == -1 || end >= len(data) {
		fmt.Printf("no stateroot found. Data: \n%q\n", string(data))
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start:end]), nil
}

func (evm *RethVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		stderr io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	cmd = exec.Command(evm.path, "statetest", "--json", path)
	if speedTest {
		cmd = exec.Command(evm.path, "statetest", "--json-outcome", path)
	}

	if stderr, err = cmd.StderrPipe(); err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	evm.Copy(out, stderr)
	// drain stderr
	_, _ = io.ReadAll(stderr)
	err = cmd.Wait()
	duration, slow := evm.stats.TraceDone(t0)

	// revm exits with 1 on test-errors (expected stateroot != observed stateroot)
	// so need to ignore it.
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		err = nil
	}

	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      cmd.String(),
	}, err
}

func (evm *RethVM) Close() {
}

func (evm *RethVM) Copy(out io.Writer, input io.Reader) {
	scanner := NewJsonlScanner("revm", input, os.Stderr)
	defer scanner.Release()
	var stateRoot stateRoot

	var elem opLog
	for scanner.Next(&elem) == nil {
		if len(elem.StateRoot1) != 0 {
			stateRoot.StateRoot = elem.StateRoot1
			break
		}
		// Drop all STOP opcodes as geth does
		if elem.Op == 0x0 {
			continue
		}
		jsondata := CustomMarshal(&elem)
		if _, err := out.Write(append(jsondata, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
	}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to output: %v\n", err)
		return
	}
}

func (evm *RethVM) Stats() []any {
	return evm.stats.Stats()
}
