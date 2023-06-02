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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/log"
)

type EvmoneVM struct {
	path string
	name string

	stats *VmStat
}

func NewEvmoneVM(path string, name string) *EvmoneVM {
	return &EvmoneVM{
		path:  path,
		name:  name,
		stats: &VmStat{},
	}
}

func (evm *EvmoneVM) Instance(int) Evm {
	return evm
}

func (evm *EvmoneVM) Name() string {
	return fmt.Sprintf("evmone-%s", evm.name)
}

func (evm *EvmoneVM) GetStateRoot(path string) (root, command string, err error) {
	cmd := exec.Command(evm.path, "--ignore-state-root", "--ignore-logs", "--trace", path)
	data, err := cmd.CombinedOutput()
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

func (evm *EvmoneVM) ParseStateRoot(data []byte) (root string, err error) {
	start := bytes.Index(data, []byte(`"stateRoot": "`))
	end := start + 14 + 66
	if start == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start+14 : end]), nil
}

func (evm *EvmoneVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		stderr io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)

	cmd = exec.Command(evm.path, "--ignore-state-root", "--ignore-logs", "--trace", path)

	if stderr, err = cmd.StderrPipe(); err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	evm.Copy(out, stderr)
	err = cmd.Wait()
	duration, slow := evm.stats.TraceDone(t0)

	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      cmd.String(),
	}, err
}

func (vm *EvmoneVM) Close() {
}

func (evm *EvmoneVM) Copy(out io.Writer, input io.Reader) {
	buf := bufferPool.Get().([]byte)
	//lint:ignore SA6002: argument should be pointer-like to avoid allocations.
	defer bufferPool.Put(buf)
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	scanner.Buffer(buf, 32*1024*1024)

	for scanner.Scan() {
		data := scanner.Bytes()

		if strings.Contains(string(data), "depth") || strings.Contains(string(data), "\"error\":null") {
			continue
		}

		if strings.Contains(string(data), "Precompile") {
			fmt.Printf("evmone err: %v\n", string(data))
			continue
		}

		if strings.Contains(string(data), "stateRoot") {
			if stateRoot.StateRoot == "" {
				_ = json.Unmarshal(data, &stateRoot)
				continue
			}
		}

		var elem logger.StructLog
		if err := json.Unmarshal(data, &elem); err != nil {
			fmt.Printf("evmone err: %v, line\n\t%v\n", err, string(data))
			continue
		}

		// Drop all STOP opcodes as geth does
		if elem.Op == 0x0 {
			continue
		}
		jsondata := FastMarshal(&elem)
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

func (evm *EvmoneVM) Stats() []any {
	return []interface{}{"execSpeed", time.Duration(evm.stats.tracingSpeedWMA.Avg()).Round(100 * time.Microsecond), "longest", evm.stats.longestTracingTime}
}
