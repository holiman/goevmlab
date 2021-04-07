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

// NethermindVM is s Evm-interface wrapper around the `nethtest` binary, based on Nethermind.
type NethermindVM struct {
	path   string
	buffer []byte // read buffer
}

func NewNethermindVM(path string) *NethermindVM {
	return &NethermindVM{
		path:   path,
		buffer: make([]byte, 4*1024*1024),
	}
}

// GetStateRoot runs the test and returns the stateroot
func (evm *NethermindVM) GetStateRoot(path string) (string, error) {
	// In this mode, we can run it without tracing
	cmd := exec.Command(evm.path, "-m", "-s", "-i", path)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	//fmt.Printf("cmd: '%v', output: %v\n", cmd.String(),string(data))
	marker := `{"stateRoot":"`
	start := strings.Index(string(data), marker)
	if start <= 0 {
		return "", errors.New("no stateroot found")
	}
	end := strings.Index(string(data)[start:], `"}`)
	if start > 0 && end > 0 {
		root := string(data[start+len(marker) : start+end])
		return root, nil
	}
	return "", errors.New("no stateroot found")
}

// RunStateTest implements the Evm interface
func (evm *NethermindVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		stderr io.ReadCloser
		err    error
	)
	if speedTest {
		return "", errors.New("nethermind does not support disabling json")
	}
	// nethtest  --input statetest1.json --trace 1> statetest1_nethermind_stdout.jsonl
	cmd := exec.Command(evm.path, "--input", path,
		"--trace",
		"-m") // -m excludes memory
	if stderr, err = cmd.StderrPipe(); err != nil {
		return cmd.String(), err
	}
	if err = cmd.Start(); err != nil {
		return cmd.String(), err
	}
	// copy everything to the given writer
	evm.Copy(out, stderr)
	// release resources, handle error but ignore non-zero exit codes
	_ = cmd.Wait()
	return cmd.String(), nil
}

func (evm *NethermindVM) Name() string {
	return "nethermind"
}

func (vm *NethermindVM) Close() {
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *NethermindVM) Copy(out io.Writer, input io.Reader) {
	var stateRoot stateRoot
	scanner := bufio.NewScanner(input)
	scanner.Buffer(evm.buffer, cap(evm.buffer))
	for scanner.Scan() {
		data := scanner.Bytes()
		var elem vm.StructLog

		// Nethermind sometimes report a negative refund
		// TODO(@holiman): they may have fixed this, if so, delete this code
		if i := bytes.Index(data, []byte(`"refund":-`)); i > 0 {
			// we can just make it positive, it will be zeroed later
			data[i+9] = byte(' ')
		}
		// Nethermind uses a hex-encoded memsize. Let's just nuke it, by remaning it
		if i := bytes.Index(data, []byte(`"memSize":"0x`)); i > 0 {
			// we can just make it positive, it will be zeroed later
			data[i+1] = byte('f')
			data[i+2] = byte('o')
			data[i+3] = byte('o')
		}

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
