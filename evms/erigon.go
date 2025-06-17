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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

// ErigonVM is s Evm-interface wrapper around the eroigon `evm` binary
type ErigonVM struct {
	path string
	name string // in case multiple instances are used
	// Some metrics
	stats *VMStat
}

func NewErigonVM(path, name string) Evm {
	return &ErigonVM{
		path:  path,
		name:  name,
		stats: new(VMStat),
	}
}

func (evm *ErigonVM) Instance(int) Evm {
	return evm
}

func (evm *ErigonVM) Name() string {
	return evm.name
}

// GetStateRoot runs the test and returns the stateroot
// This currently only works for non-filled statetests. TODO: make it work even if the
// test is filled. Either by getting the whole trace, or adding stateroot to exec std output
// even in success-case
func (evm *ErigonVM) GetStateRoot(path string) (root, command string, err error) {
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

// ParseStateRoot reads the stateroot from the combined output.
func (evm *ErigonVM) ParseStateRoot(data []byte) (string, error) {
	start := bytes.Index(data, []byte(`"stateRoot": "`))
	end := start + 14 + 66
	if start == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start+14 : end]), nil
}

// RunStateTest implements the Evm interface
func (evm *ErigonVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		stderr io.ReadCloser
		err    error
		cmd    = exec.Command(evm.path, "--json", "--noreturndata", "--nomemory", "statetest", path)
	)
	if speedTest {
		cmd = exec.Command(evm.path, "--nomemory", "--noreturndata", "--nostack", "statetest", path)
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
			Cmd:      cmd.String()},
		err
}

func (evm *ErigonVM) Close() {
}

// Copy reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *ErigonVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input, false)
}

// copyUntilEnd reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *ErigonVM) copyUntilEnd(out io.Writer, input io.Reader, speedMode bool) stateRoot {
	if speedMode {
		// In speednode, there's no jsonl output, just the json stateroot
		var r []stateRoot
		if err := json.NewDecoder(input).Decode(&r); err != nil {
			log.Warn("Error parsing erigonbatch output", "error", err)
			return stateRoot{}
		}
		rootJSON, _ := json.Marshal(r[0])
		if _, err := out.Write(append(rootJSON, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
		return r[0]
	}
	scanner := NewJsonlScanner("erigon", input, os.Stderr)
	defer scanner.Release()
	var stateRoot stateRoot

	for {
		var elem opLog
		if err := scanner.Next(&elem); err != nil {
			break
		}
		// If we have a stateroot, we're done
		if len(elem.StateRoot1) != 0 {
			stateRoot.StateRoot = elem.StateRoot1
			break
		}
		if elem.Depth == 0 {
			continue
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
	}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
	return stateRoot
}

func (evm *ErigonVM) Stats() []any {
	return evm.stats.Stats()
}
