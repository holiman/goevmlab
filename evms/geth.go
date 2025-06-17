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

// GethEVM is s Evm-interface wrapper around the `evm` binary, based on go-ethereum.
type GethEVM struct {
	path string
	name string // in case multiple instances are used

	// Some metrics
	stats *VMStat
}

func NewGethEVM(path string, name string) Evm {
	return &GethEVM{
		path:  path,
		name:  name,
		stats: &VMStat{},
	}
}

func (evm *GethEVM) Instance(int) Evm {
	return evm
}

func (evm *GethEVM) Name() string {
	return evm.name
}

// GetStateRoot runs the test and returns the stateroot
// This currently only works for non-filled statetests. TODO: make it work even if the
// test is filled. Either by getting the whole trace, or adding stateroot to exec std output
// even in success-case
func (evm *GethEVM) GetStateRoot(path string) (root, command string, err error) {
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
func (evm *GethEVM) ParseStateRoot(data []byte) (string, error) {
	start := bytes.Index(data, []byte(`"stateRoot": "`))
	end := start + 14 + 66
	if start == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start+14 : end]), nil
}

// RunStateTest implements the Evm interface
func (evm *GethEVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		stderr io.ReadCloser
		err    error
		//cmd    = exec.Command(evm.path, "--json", "--noreturndata", "--nomemory", "statetest", path)
		cmd = exec.Command(evm.path, "statetest", "--trace", "--trace.format=json",
			"--trace.nomemory=true", "--trace.noreturndata=true", path)
	)
	if speedTest {
		//cmd = exec.Command(evm.path, "--nomemory", "--noreturndata", "--nostack", "statetest", path)
		cmd = exec.Command(evm.path, "statetest", path)
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

func (evm *GethEVM) Close() {
}

// Copy reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *GethEVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input)
}

// copyUntilEnd reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *GethEVM) copyUntilEnd(out io.Writer, input io.Reader) stateRoot {
	scanner := NewJsonlScanner("geth", input, os.Stderr)
	defer scanner.Release()
	var stateRoot stateRoot
	// When geth encounters an error, it may already have spat out the info, prematurely.
	// We need to merge it back to one item
	// https://github.com/ethereum/go-ethereum/pull/23970#issuecomment-979851712
	var prev *opLog
	var yield = func(current *opLog) {
		if prev == nil {
			prev = current
			return
		}
		data := CustomMarshal(prev)
		if _, err := out.Write(append(data, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
		if current == nil { // final flush
			return
		}
		if prev.Pc == current.Pc && prev.Depth == current.Depth && prev.FunctionDepth == current.FunctionDepth {
			// Yup, that happened here. Set the error and continue
			prev = nil
		} else {
			prev = current
		}
	}
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
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			continue
		}
		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}
		yield(&elem)
	}
	yield(nil)
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
	return stateRoot
}

func (evm *GethEVM) Stats() []any {
	return evm.stats.Stats()
}
