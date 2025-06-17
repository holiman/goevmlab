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

// NimbusEVM is s Evm-interface wrapper around the `evmstate` binary, based on nimbus-eth1.
type NimbusEVM struct {
	path string
	name string
	// Some metrics
	stats *VMStat
}

func NewNimbusEVM(path string, name string) Evm {
	return &NimbusEVM{
		path:  path,
		name:  name,
		stats: &VMStat{},
	}
}

func (evm *NimbusEVM) Instance(int) Evm {
	return evm
}

func (evm *NimbusEVM) Name() string {
	return evm.name
}

// GetStateRoot runs the test and returns the stateroot
// This currently only works for non-filled statetests. TODO: make it work even if the
// test is filled. Either by getting the whole trace, or adding stateroot to exec std output
// even in success-case
func (evm *NimbusEVM) GetStateRoot(path string) (root, command string, err error) {
	// In this mode, we can run it without tracing
	cmd := exec.Command(evm.path, path)
	data, _ := cmd.Output()

	root, err = evm.ParseStateRoot(data)
	if err != nil {
		log.Error("Failed to find stateroot", "vm", evm.Name(), "cmd", cmd.String())
		return "", cmd.String(), err
	}
	return root, cmd.String(), err
}

// ParseStateRoot reads geth's stateroot from the combined output.
func (evm *NimbusEVM) ParseStateRoot(data []byte) (string, error) {

	start := bytes.Index(data, []byte(`"stateRoot": "`))
	end := start + 14 + 66
	if start == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start+14 : end]), nil
}

// RunStateTest implements the Evm interface
func (evm *NimbusEVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		stderr io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	if speedTest {
		cmd = exec.Command(evm.path, "--noreturndata", "--nomemory", "--nostorage", path)
	} else {
		cmd = exec.Command(evm.path, "--json", "--noreturndata", "--nomemory", "--nostorage", path)
	}
	if stderr, err = cmd.StderrPipe(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	if err = cmd.Start(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	// copy everything to the given writer
	evm.Copy(out, stderr)
	// Nimbus returns a non-zero exit code for tests that do not pass. We just ignore that.
	_, _ = io.ReadAll(stderr)
	_ = cmd.Wait()
	// release resources
	duration, slow := evm.stats.TraceDone(t0)

	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      cmd.String(),
	}, nil
}

func (evm *NimbusEVM) Close() {
}

func (evm *NimbusEVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input, false)
}

// copyUntilEnd copies the input to output, and returns the stateroot on completion
func (evm *NimbusEVM) copyUntilEnd(out io.Writer, input io.Reader, speedMode bool) stateRoot {
	if speedMode {
		// In speednode, there's no jsonl output, it instead looks like
		var r []stateRoot
		if err := json.NewDecoder(input).Decode(&r); err != nil {
			log.Warn("Error parsing nimbus output", "error", err)
			return stateRoot{}
		}
		rootJSON, _ := json.Marshal(r[0])
		if _, err := out.Write(append(rootJSON, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
		return r[0]
	}
	scanner := NewJsonlScanner("nimb", input, os.Stderr)
	defer scanner.Release()
	var root stateRoot

	// When nimbus encounters an error, it may already have spat out the info prematurely.
	// We need to merge it back to one item, just like geth
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
		if prev.Pc == current.Pc && prev.Depth == current.Depth {
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
			root.StateRoot = elem.StateRoot1
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
	rootJSON, _ := json.Marshal(root)
	if _, err := out.Write(append(rootJSON, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
	return root
}

func (evm *NimbusEVM) Stats() []any {
	return evm.stats.Stats()
}
