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

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/uint256"
	"math/big"
	"strings"
)

// NimbusEVM is s Evm-interface wrapper around the `evmstate` binary, based on nimbus-eth1.
type NimbusEVM struct {
	path string
	name string
}

func NewNimbusEVM(path string, name string) *NimbusEVM {
	return &NimbusEVM{
		path: path,
		name: name,
	}
}

func (evm *NimbusEVM) Name() string {
	return fmt.Sprintf("nimb-%v", evm.name)
}

// GetStateRoot runs the test and returns the stateroot
// This currently only works for non-filled statetests. TODO: make it work even if the
// test is filled. Either by getting the whole trace, or adding stateroot to exec std output
// even in success-case
func (evm *NimbusEVM) GetStateRoot(path string) (root, command string, err error) {
	// In this mode, we can run it without tracing
	cmd := exec.Command(evm.path, path)
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

// ParseStateRoot reads geth's stateroot from the combined output.
func (evm *NimbusEVM) ParseStateRoot(data []byte) (string, error) {

	start := bytes.Index(data, []byte(`post state root mismatch: got `))
	end := start + 30 + 64
	if start == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	root := strings.ToLower("0x" + string(data[start+30:end]))
	return root, nil
}

// RunStateTest implements the Evm interface
func (evm *NimbusEVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
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
		return cmd.String(), err
	}
	if err = cmd.Start(); err != nil {
		return cmd.String(), err
	}
	// copy everything to the given writer
	evm.Copy(out, stderr)
	// release resources
	return cmd.String(), cmd.Wait()
}

func (vm *NimbusEVM) Close() {
}

type nimbusOp struct {
	Op      string
	Pc      uint64
	Depth   int
	Gas     uint64
	Stack   []string
	GasCost uint64
}

func (evm *NimbusEVM) Copy(out io.Writer, input io.Reader) {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	// We use a larger scanner buffer for besu: it does not have a way to
	// disable 'returndata', which can become larger than fits into a default
	// scanner buffer
	buf := make([]byte, 4*1024*1024)
	scanner.Buffer(buf, cap(buf))
	for scanner.Scan() {
		data := scanner.Bytes()
		var op nimbusOp
		err := json.Unmarshal(data, &op)
		if err != nil {
			fmt.Printf("nimb err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		var stack []uint256.Int
		for _, x := range op.Stack {
			a, _ := big.NewInt(0).SetString(x, 16)
			y := new(uint256.Int)
			y.SetFromBig(a)
			stack = append(stack, *y)
		}
		var elem = logger.StructLog{
			Pc:      op.Pc,
			Op:      vm.OpCode(ops.StringToOp(op.Op)),
			Gas:     op.Gas,
			GasCost: op.GasCost,
			Stack:   stack,
			Depth:   op.Depth,
			Err:     nil,
		}
		//err := json.Unmarshal(data, &elem)
		//if err != nil {
		//	fmt.Printf("nimb err: %v, line\n\t%v\n", err, string(data))
		//	continue
		//}
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
		RemoveUnsupportedElems(&elem)
		jsondata, _ := json.Marshal(elem)
		if _, err := out.Write(append(jsondata, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
			return
		}
	}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
	return
}
