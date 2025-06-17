// Copyright 2019 Martin Holst Swende
// Copyright 2024 Sam Wilson
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

// EelsEVM is s Evm-interface wrapper around the `evm` binary, based on go-ethereum.
type EelsEVM struct {
	path string
	name string // in case multiple instances are used

	// Some metrics
	stats *VMStat
}

func NewEelsEVM(path string, name string) Evm {
	return &EelsEVM{
		path:  path,
		name:  name,
		stats: &VMStat{},
	}
}

func (evm *EelsEVM) Instance(int) Evm {
	return evm
}

func (evm *EelsEVM) Name() string {
	return evm.name
}

// GetStateRoot runs the test and returns the stateroot
// This currently only works for non-filled statetests. TODO: make it work even if the
// test is filled. Either by getting the whole trace, or adding stateroot to exec std output
// even in success-case
func (evm *EelsEVM) GetStateRoot(path string) (root, command string, err error) {
	// In this mode, we can run it without tracing
	cmd := exec.Command(evm.path, "statetest", path)
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

// ParseStateRoot reads geth's stateroot from the combined output.
func (evm *EelsEVM) ParseStateRoot(data []byte) (string, error) {
	start := bytes.Index(data, []byte(`"stateRoot": "`))
	end := start + 14 + 66
	if start == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start+14 : end]), nil
}

// RunStateTest implements the Evm interface
func (evm *EelsEVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		stderr io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	if speedTest {
		cmd = exec.Command(evm.path, "statetest", "--nomemory", "--noreturndata", "--nostack", path)
	} else {
		cmd = exec.Command(evm.path, "statetest", "--json", "--noreturndata", "--nomemory", path)
	}
	if stderr, err = cmd.StderrPipe(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	if err = cmd.Start(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	// copy everything to the given writer
	evm.Copy(out, stderr)
	_, _ = io.ReadAll(stderr)
	err = cmd.Wait()
	// release resources
	duration, slow := evm.stats.TraceDone(t0)

	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      cmd.String(),
	}, err
}

func (evm *EelsEVM) Close() {
}

// Copy reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *EelsEVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input)
}

// copyUntilEnd reads from the reader, does some vm-specific filtering and
// outputs items onto the channel
func (evm *EelsEVM) copyUntilEnd(out io.Writer, input io.Reader) stateRoot {
	scanner := NewJsonlScanner("eels", input, os.Stderr)
	defer scanner.Release()
	var stateRoot stateRoot
	for {
		var elem opLog
		if scanner.Next(&elem) != nil {
			break
		}
		if len(elem.StateRoot1) != 0 {
			stateRoot.StateRoot = elem.StateRoot1
			break
		}
		// General parsing error, some jsonl-output with nothing meaningful
		if elem.Depth == 0 {
			continue
		}
		// Drop the stops
		if elem.Op == 0x0 {
			continue
		}
		outp := CustomMarshal(&elem)
		if _, err := out.Write(append(outp, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
	}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
	return stateRoot
}

func (evm *EelsEVM) Stats() []any {
	return evm.stats.Stats()
}
