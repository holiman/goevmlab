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

	"github.com/ethereum/go-ethereum/core/vm"
)

// BesuVM is s Evm-interface wrapper around the `evmtool` binary, based on Besu.
type BesuVM struct {
	path string
}

func NewBesuVM(path string) *BesuVM {
	return &BesuVM{
		path: path,
	}
}

// RunStateTest implements the Evm interface
func (evm *BesuVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		stdout io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	if speedTest {
		cmd = exec.Command(evm.path, "--nomemory",
			"--Xberlin-enabled", "true",
			"state-test", path)
	} else {
		cmd = exec.Command(evm.path, "--nomemory",
			"--Xberlin-enabled", "true",
			"--json",
			"state-test", path) // exclude memory
		// For running this via docker, this is the 'raw' base command
		//docker run --rm  -i -v ~/yolov2:/yolov2/ hyperledger/besu-evmtool:develop --Xberlin-enabled true state-test  /yolov2/tests/$f

	}
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return cmd.String(), err
	}
	if err = cmd.Start(); err != nil {
		return cmd.String(), err
	}
	// copy everything to the given writer
	evm.Copy(out, stdout)
	// release resources, handle error but ignore non-zero exit codes
	_ = cmd.Wait()
	return cmd.String(), nil
}

func (evm *BesuVM) Name() string {
	return "besu"
}

func (vm *BesuVM) Close() {
}

func (vm *BesuVM) GetStateRoot(path string) (string, error) {
	// Run without tracing
	cmd := exec.Command(vm.path, "--nomemory",
		"--Xberlin-enabled", "true",
		"state-test", path)

	/// {"output":"","gasUsed":"0x798765","time":342028058,"test":"00000000-storagefuzz-0","fork":"Berlin","d":0,"g":0,"v":0,"postHash":"0xbaeec41171e943575740bb99cdd8ffd45fa6207bf170f99c1a61425069ffae97","postLogsHash":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","pass":false}
	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	start := strings.Index(string(data), `"postHash:":"`)
	if start > 0 {
		start = start + len(`"postHash:":"`)
		root := string(data[start : start+64])
		return root, nil
	}
	return "", errors.New("no stateroot found")
}

type besuStateRoot struct {
	StateRoot string `json:"postHash"`
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *BesuVM) Copy(out io.Writer, input io.Reader) {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	// We use a larger scanner buffer for besu: it does not have a way to
	// disable 'returndata', which can become larger than fits into a default
	// scanner buffer
	scanner.Buffer(make([]byte, 1*1024*1024), 1*1024*1024)
	for scanner.Scan() {
		data := scanner.Bytes()
		var elem vm.StructLog
		// Besu (like Nethermind) sometimes report a negative refund
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
					// Besu calls state root "rootHash"
					stateRoot.StateRoot = tempRoot.StateRoot
				}
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
		return
	}
}
