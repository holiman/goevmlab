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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/log"
)

// NethermindVM is s Evm-interface wrapper around the `nethtest` binary, based on Nethermind.
type NethermindVM struct {
	path string
	name string
	// Some metrics
	stats *VmStat
}

func NewNethermindVM(path, name string) *NethermindVM {
	return &NethermindVM{
		path:  path,
		name:  name,
		stats: &VmStat{},
	}
}

func (evm *NethermindVM) Instance(threadId int) Evm {
	return &NethermindVM{
		path:  evm.path,
		name:  fmt.Sprintf("%v-%d", evm.name, threadId),
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

// getStateRoot reads the stateroot from the combined output.
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
		t0     = time.Now()
		stderr io.ReadCloser
		err    error
		cmd    = exec.Command(evm.path, "--trace", "-m", "--input", path)
	)
	if speedTest {
		cmd = exec.Command(evm.path, "-m", "--neverTrace", "--input", path)
	}
	if stderr, err = cmd.StderrPipe(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	if err = cmd.Start(); err != nil {
		return &tracingResult{Cmd: cmd.String()}, err
	}
	// copy everything to the given writer
	evm.Copy(out, stderr)
	// release resources, handle error but ignore non-zero exit codes
	_ = cmd.Wait()
	duration, slow := evm.stats.TraceDone(t0)
	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      cmd.String()}, nil
}

func (vm *NethermindVM) Close() {
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *NethermindVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input)
}

func (evm *NethermindVM) copyUntilEnd(out io.Writer, input io.Reader) stateRoot {
	buf := bufferPool.Get().([]byte)
	//lint:ignore SA6002: argument should be pointer-like to avoid allocations.
	defer bufferPool.Put(buf)
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	scanner.Buffer(buf, 32*1024*1024)

	for scanner.Scan() {
		data := scanner.Bytes()
		var elem logger.StructLog

		err := json.Unmarshal(data, &elem)
		if err != nil {
			fmt.Printf("nethermind err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			/*  Most likely one of these:
			{"output":"","gasUsed":"0x2d1cc4","time":233624,"error":"gas uint64 overflow"}
			{"stateRoot": "a2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134"}
			*/
			if stateRoot.StateRoot == "" {
				_ = json.Unmarshal(data, &stateRoot)
			}
			// If we have a stateroot, we're done
			if len(stateRoot.StateRoot) > 0 {
				break
			}

			//fmt.Printf("%v\n", string(data))
			// For now, just ignore these
			continue
		}
		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}
		outp := FastMarshal(&elem)
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
	return []interface{}{"execSpeed", time.Duration(evm.stats.tracingSpeedWMA.Avg()).Round(100 * time.Microsecond), "longest", evm.stats.longestTracingTime}
}
