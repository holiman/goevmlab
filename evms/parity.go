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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/ethereum/go-ethereum/core/vm"
)

type ParityVM struct {
	path string
}

func NewParityVM(path string) *ParityVM {
	return &ParityVM{
		path: path,
	}
}

func (evm *ParityVM) Name() string {
	return "parity"
}

// GetStateRoot runs the test and returns the stateroot
func (evm *ParityVM) GetStateRoot(path string) (string, error) {
	// In this mode, we can run it without tracing
	cmd := exec.Command(evm.path, "state-test", path)
	data, err := cmd.Output()
	if err != nil {
		//fmt.Printf("cmd: '%v', output: %v\n", cmd.String(),string(data))
		return "", err
	}
	//fmt.Printf("cmd: '%v', output: %v\n", cmd.String(),string(data))
	marker := ` State root mismatch (got: `
	start := strings.Index(string(data), marker)
	if start <= 0 {
		return "", errors.New("no stateroot found")
	}
	end := strings.Index(string(data)[start:], `, expected`)
	if start > 0 && end > 0 {
		root := string(data[start+len(marker) : start+end])
		return root, nil
	}
	return "", errors.New("no stateroot found")
}

// RunStateTest implements the Evm interface
func (evm *ParityVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		stderr io.ReadCloser
		stdout io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	if speedTest {
		cmd = exec.Command(evm.path, "state-test", path)
	} else {
		cmd = exec.Command(evm.path, "--std-json", "state-test", path)
	}
	if stderr, err = cmd.StderrPipe(); err != nil {
		return cmd.String(), err
	}
	// Parity, when there's an error on state root, spits out the error on
	// stdout. So we need to read that aswell
	if stdout, err = cmd.StdoutPipe(); err != nil {
		return cmd.String(), err
	}
	if err = cmd.Start(); err != nil {
		return cmd.String(), err
	}
	// copy everything to the given writer
	evm.Copy(out, stderr)
	// copy everything to the given writer -- this means that the
	// stdout output will come _after_ all stderr data. Which is good.
	evm.Copy(out, stdout)
	// release resources, handle error but ignore non-zero exit codes
	_ = cmd.Wait()
	return cmd.String(), nil
}

func (evm *ParityVM) Close() {
}

type parityErrorRoot struct {
	Error string
}

func (evm *ParityVM) Copy(out io.Writer, input io.Reader) {
	var sRoot stateRoot
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		// Calling bytes means that bytes in 'l' will be overwritten
		// in the next loop. Fine for now though, we immediately marshal it
		data := scanner.Bytes()
		var elem vm.StructLog
		_ = json.Unmarshal(data, &elem)
		// We ignore json errors, but need to see if the elem was parsed ok.
		// We can use depth for that
		if elem.Depth == 0 {
			/*  Most likely one of these:
			{"error":"State root mismatch (got: 0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134, expected: 0x0000000000000000000000000000000000000000000000000000000000000000)","gasUsed":"0x2dc6c0","time":146}
			*/
			if sRoot.StateRoot == "" {
				// Try to unmarshal an error root
				var p parityErrorRoot
				if err := json.Unmarshal(data, &p); err == nil {
					prefix := `State root mismatch (got: `
					if strings.HasPrefix(p.Error, prefix) {
						root := []byte(strings.TrimPrefix(p.Error, prefix))
						sRoot.StateRoot = string(root[:66])
						continue
					}
				}
				// Try to unmarshal the state root if no error occured
				var tmp stateRoot
				if err := json.Unmarshal(data, &tmp); err != nil {
					sRoot.StateRoot = tmp.StateRoot
				}
			}
			continue
		}
		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}
		RemoveUnsupportedElems(&elem)
		//fmt.Printf("parity: %v\n", string(data))
		jsondata, _ := json.Marshal(elem)
		if _, err := out.Write(append(jsondata, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
			return
		}
	}
	if sRoot.StateRoot != "" {
		root, _ := json.Marshal(sRoot)
		if _, err := out.Write(append(root, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
			return
		}
	}
}
