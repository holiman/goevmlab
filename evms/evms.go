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
	"fmt"
	"io"
	"strings"
	"sync"
)

// The Evm interface represents external EVM implementations, which can
// be e.g. docker instances or binaries
type Evm interface {
	// RunStateTest runs the statetest on the underlying EVM, and writes
	// the output to the given writer
	RunStateTest(path string, writer io.Writer, skipTrace bool) (*tracingResult, error)
	// GetStateRoot runs the test and returns the stateroot
	GetStateRoot(path string) (root, command string, err error)
	// ParseStateRoot reads the stateroot from the combined output.
	ParseStateRoot([]byte) (string, error)
	// Copy takes the 'raw' output from the VM, and writes the
	// canonical output to the given writer
	Copy(out io.Writer, input io.Reader)
	//Open() // Preparare for execution
	Close() // Tear down processes
	Name() string
	Stats() []any

	// Instance delivers an instance of the EVM which will be executed per-thread.
	// This method may deliver the same instance each time, but it may also
	// deliver e.g. a unique version which has preallocated buffers. Such an instance
	// is not concurrency-safe, but is fine to deliver in this method.
	Instance(threadID int) Evm
}

type stateRoot struct {
	StateRoot string `json:"stateRoot"`
}

// CompareFiles returns true if the files are equal, along with the number of line s
// compared
func CompareFiles(vms []Evm, readers []io.Reader) (bool, int, string) {
	var output = new(strings.Builder)
	var scanners []*bufio.Scanner
	for _, r := range readers {
		scanner := bufio.NewScanner(r)
		buf := bufferPool.Get().([]byte)
		//lint:ignore SA6002: argument should be pointer-like to avoid allocations.
		defer bufferPool.Put(buf)
		scanner.Buffer(buf, len(buf))
		scanners = append(scanners, scanner)

	}
	var (
		count     = 0
		prevLine  = ""
		curLines  = make([]string, len(scanners))
		diffFound = false
	)
	for curLines[0] != "EOF" {
		// Scan next line of output from every VM
		for i, scanner := range scanners {
			if !scanner.Scan() {
				curLines[i] = "EOF"
			} else {
				curLines[i] = scanner.Text()
			}
			diffFound = diffFound || (curLines[i] != curLines[0])
		}
		// If diff, show output and return
		if diffFound {
			fmt.Fprintf(output, "-------\n%4d: %15v: %v\n", count, "all", prevLine)
			for i, line := range curLines {
				fmt.Fprintf(output, "%4d: %15v: %v\n", count+1, vms[i].Name(), line)
			}
			return false, count, output.String()
		}
		count++
		prevLine = curLines[0]
	}
	return true, count, output.String()
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		// A 5Mb buffer
		return make([]byte, 5*1024*1025)
	},
}
