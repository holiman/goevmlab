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
	"fmt"
	"github.com/ethereum/go-ethereum/core/vm"
	"io"
)

// The Evm interface represents external EVM implementations, which can
// be e.g. docker instances or binaries
type Evm interface {
	// RunStateTest runs the statetest on the underlying EVM, and writes
	// the output to the given writer
	RunStateTest(path string, writer io.Writer) error

	// Copy takes the 'raw' output from the VM, and writes the
	// canonical output to the given writer
	Copy(out io.Writer, input io.Reader)
	//Open() // Preparare for execution
	Close() // Tear down processes
	Name() string
}

type stateRoot struct {
	StateRoot string `json:"stateRoot"`
}

// logString provides a human friendly string
func logString(log *vm.StructLog) string {
	return fmt.Sprintf("pc: %3d op: %18v depth: %2v gas: %5d stack size %d",
		log.Pc, log.Op, log.Depth, log.Gas, len(log.Stack))

}

func CompareFiles(vms []Evm, readers []io.Reader) bool {
	var scanners []*bufio.Scanner
	for _, r := range readers {
		scanners = append(scanners, bufio.NewScanner(r))
	}
	refOut := scanners[0]
	refVm := vms[0]
	for refOut.Scan() {
		//fmt.Printf("ref: %v\n", string(refOut.Bytes()))
		for i, scanner := range scanners[1:] {
			scanner.Scan()
			if !bytes.Equal(refOut.Bytes(), scanner.Bytes()) {
				fmt.Printf("diff: \n%v: %v\n%v: %v\n",
					refVm.Name(),
					string(refOut.Bytes()),
					vms[i+1].Name(),
					string(scanner.Bytes()))
				return false
			}
		}
	}
	return true
}
