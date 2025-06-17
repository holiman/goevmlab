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

// NethermindVM is s Evm-interface wrapper around the `nethtest` binary, based on Nethermind.
type NethermindVM struct {
	path string
	name string
	// Some metrics
	stats *VMStat
}

func NewNethermindVM(path, name string) Evm {
	return &NethermindVM{
		path:  path,
		name:  name,
		stats: &VMStat{},
	}
}

func (evm *NethermindVM) Instance(threadID int) Evm {
	return &NethermindVM{
		path:  evm.path,
		name:  fmt.Sprintf("%v-%d", evm.name, threadID),
		stats: evm.stats,
	}
}

func (evm *NethermindVM) Name() string {
	return evm.name
}

// GetStateRoot runs the test and returns the stateroot
func (evm *NethermindVM) GetStateRoot(path string) (root, command string, err error) {
	// In this mode, we can run it without tracing
	cmd := exec.Command(evm.path, "--neverTrace", "-m", "-s", "-i", path)
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
func (evm *NethermindVM) ParseStateRoot(data []byte) (string, error) {
	start := bytes.Index(data, []byte(`"stateRoot": "`))
	end := start + 14 + 66
	if start == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start+14 : end]), nil
}

// RunStateTest implements the Evm interface
func (evm *NethermindVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0      = time.Now()
		procOut io.ReadCloser
		err     error
		cmd     = exec.Command(evm.path, "--trace", "-m", "--input", path)
	)
	if !speedTest {
		// in normal execution, we read traces from standard error
		if procOut, err = cmd.StderrPipe(); err != nil {
			return &tracingResult{Cmd: cmd.String()}, err
		}
	} else {
		// In speedtest-mode, we don't want the actual traces, but we do
		// need to read the stateroot. The stateroot can be found on stdout
		cmd = exec.Command(evm.path, "-m", "--neverTrace", "--input", path)
		if procOut, err = cmd.StdoutPipe(); err != nil {
			return &tracingResult{Cmd: cmd.String()}, err
		}
	}
	if err = cmd.Start(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	// copy everything to the given writer
	evm.copyUntilEnd(out, procOut, speedTest)
	// release resources, handle error but ignore non-zero exit codes
	_, _ = io.ReadAll(procOut)
	_ = cmd.Wait()
	duration, slow := evm.stats.TraceDone(t0)
	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      cmd.String()}, nil
}

func (evm *NethermindVM) Close() {
}

// Copy feed reads from the reader, does some vm-specific filtering and
// outputs items onto the channel
func (evm *NethermindVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input, false)
}

func (evm *NethermindVM) copyUntilEnd(out io.Writer, input io.Reader, speedMode bool) stateRoot {
	if speedMode {
		// In speednode, there's no jsonl output, it instead looks like
		var r []stateRoot
		if err := json.NewDecoder(input).Decode(&r); err != nil {
			log.Warn("Error parsing nethermind output", "error", err)
			return stateRoot{}
		}
		rootJSON, _ := json.Marshal(r[0])
		if _, err := out.Write(append(rootJSON, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
		return r[0]
	}
	scanner := NewJsonlScanner("neth", input, os.Stderr)
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
		// Drop stops
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

func (evm *NethermindVM) Stats() []any {
	return evm.stats.Stats()
}
