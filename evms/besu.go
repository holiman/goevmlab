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
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// BesuVM is s Evm-interface wrapper around the `evmtool` binary, based on Besu.
type BesuVM struct {
	path string
	name string // in case multiple instances are used
}

func NewBesuVM(path, name string) *BesuVM {
	return &BesuVM{
		path: path,
		name: name,
	}
}

func (evm *BesuVM) Name() string {
	return fmt.Sprintf("besu-%v", evm.name)
}

// RunStateTest implements the Evm interface
func (evm *BesuVM) RunStateTest(path string, out io.Writer, speedTest bool) (string, error) {
	var (
		stdout io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	if speedTest {
		cmd = exec.Command(evm.path, "--nomemory", "state-test", path)
	} else {
		cmd = exec.Command(evm.path, "--nomemory", "--json",
			"state-test", path) // exclude memory
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

func (vm *BesuVM) Close() {
}

func (vm *BesuVM) GetStateRoot(path string) (root, command string, err error) {
	// Run without tracing
	cmd := exec.Command(vm.path, "--nomemory", "state-test", path)

	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", cmd.String(), err
	}
	start := strings.Index(string(data), `"postHash":"`)
	if start > 0 {
		start = start + len(`"postHash":"`)
		root := string(data[start : start+2+64])
		return root, cmd.String(), nil
	}
	return "", cmd.String(), errors.New("no stateroot found")
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *BesuVM) Copy(out io.Writer, input io.Reader) {
	besuCopyUntilEnd(out, input)
}
