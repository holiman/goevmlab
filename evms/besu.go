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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/log"
)

// BesuVM is s Evm-interface wrapper around the `evmtool` binary, based on Besu.
type BesuVM struct {
	path string
	name string // in case multiple instances are used
	// Some metrics
	stats *VmStat
}

func NewBesuVM(path, name string) *BesuVM {
	return &BesuVM{
		path: path,
		name: name,
	}
}

func (evm *BesuVM) Instance() Evm {
	return evm
}

func (evm *BesuVM) Name() string {
	return fmt.Sprintf("besu-%v", evm.name)
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
	err = cmd.Wait()
	// release resources
	duration, slow := evm.stats.TraceDone(t0)

	return &tracingResult{
			Slow:     slow,
			ExecTime: duration,
			Cmd:      cmd.String()},
		err
}

func (vm *BesuVM) Close() {}

func (vm *BesuVM) GetStateRoot(path string) (root, command string, err error) {
	// Run without tracing
	cmd := exec.Command(vm.path, "--nomemory", "--notime", "state-test", path)

	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", cmd.String(), err
	}
	root, err = vm.ParseStateRoot(data)
	if err != nil {
		log.Error("Failed to find stateroot", "vm", vm.Name(), "cmd", cmd.String())
		return "", cmd.String(), err
	}
	return root, cmd.String(), err
}

// ParseStateRoot reads the stateroot from the combined output.
func (vm *BesuVM) ParseStateRoot(data []byte) (string, error) {
	start := strings.Index(string(data), `"postHash":"`)
	if start > 0 {
		start = start + len(`"postHash":"`)
		root := string(data[start : start+2+64])
		return root, nil
	}
	return "", errors.New("besu: no stateroot found")
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *BesuVM) Copy(out io.Writer, input io.Reader) {
	evm.copyUntilEnd(out, input)
}

type besuStateRoot struct {
	StateRoot string `json:"postHash"`
}

func (evm *BesuVM) copyUntilEnd(out io.Writer, input io.Reader) stateRoot {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	// Start with 1MB buffer, allow up to 32 MB
	scanner.Buffer(make([]byte, 1024*1024), 32*1024*1024)
	for scanner.Scan() {
		data := scanner.Bytes()
		var elem logger.StructLog
		// Besu sometimes report a negative refund
		if i := bytes.Index(data, []byte(`"refund":-`)); i > 0 {
			// we can just make it positive, it will be zeroed later
			data[i+9] = byte(' ')
		}
		err := json.Unmarshal(data, &elem)
		if err != nil {
			fmt.Printf("besu err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			if stateRoot.StateRoot == "" {
				var tempRoot besuStateRoot
				if err := json.Unmarshal(data, &tempRoot); err == nil {
					// Besu calls state root "postHash"
					stateRoot.StateRoot = tempRoot.StateRoot
				}
			}
			// If we have a stateroot, we're done
			break
		}
		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}
		RemoveUnsupportedElems(&elem)
		jsondata, _ := json.Marshal(elem)
		if _, err := out.Write(append(jsondata, '\n')); err != nil {
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

func (evm *BesuVM) Stats() []any {
	return []interface{}{"execSpeed", time.Duration(evm.stats.tracingSpeedWMA.Avg()), "longest", evm.stats.longestTracingTime}
}
