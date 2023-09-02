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

	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/log"
	"time"
)

// NimbusEVM is s Evm-interface wrapper around the `evmstate` binary, based on nimbus-eth1.
type NimbusEVM struct {
	path string
	name string
	// Some metrics
	stats *VmStat
}

func NewNimbusEVM(path string, name string) *NimbusEVM {
	return &NimbusEVM{
		path:  path,
		name:  name,
		stats: &VmStat{},
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
	_ = cmd.Wait()
	// release resources
	duration, slow := evm.stats.TraceDone(t0)

	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      cmd.String(),
	}, nil
}

func (vm *NimbusEVM) Close() {
}

func (evm *NimbusEVM) Copy(out io.Writer, input io.Reader) {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	// Start with 1MB buffer, allow up to 32 MB
	scanner.Buffer(make([]byte, 1024*1024), 32*1024*1024)
	for scanner.Scan() {
		data := scanner.Bytes()

                // Nimbus sometimes report a negative refund
                if i := bytes.Index(data, []byte(`"refund":-`)); i > 0 {
                        // we can just make it positive, it will be zeroed later
                        data[i+9] = byte(' ')
                }

		var elem logger.StructLog
		if err := json.Unmarshal(data, &elem); err != nil {
			fmt.Printf("nimb err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			if stateRoot.StateRoot == "" {
				_ = json.Unmarshal(data, &stateRoot)
			}
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
			return
		}
	}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
}

func (evm *NimbusEVM) Stats() []any {
	return []interface{}{"execSpeed", time.Duration(evm.stats.tracingSpeedWMA.Avg()).Round(100 * time.Microsecond), "longest", evm.stats.longestTracingTime}
}
