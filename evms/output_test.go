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
	//"fmt"
	"github.com/ethereum/go-ethereum/core/vm"
	"os"
	"testing"
)

// TestGethVsParityStatetestOutput in this test, we simulate two external
// vms, using printouts from actual evm binaries. The two outputs should
// not produce any differences
func TestGethVsParityStatetestOutput(t *testing.T) {
	var (
		gethChan   = make(chan *vm.StructLog)
		parityChan = make(chan *vm.StructLog)
	)

	// Set up geth
	{
		output, err := os.Open("testdata/statetest1_geth_stderr.jsonl")
		if err != nil {
			t.Fatal(err)
		}
		go func(file *os.File) {
			vm := NewGethEVM("")
			vm.wg.Add(1)
			vm.feed(file, gethChan)
		}(output)
	}
	// Set up parity
	{
		output, err := os.Open("testdata/statetest1_parity_stderr.jsonl")
		if err != nil {
			t.Fatal(err)
		}
		go func(file *os.File) {
			vm := NewParityVM("")
			vm.wg.Add(1)
			vm.feed(file, parityChan)
		}(output)
	}
	c := &Comparer{}
	// Now we have two channels that will spit out OutputItems
	outCh := c.CompareVms(gethChan, parityChan)
	for outp := range outCh {
		t.Errorf("Expected no diff, got %v", outp)
	}
}

// TestDifferentLengthOutput checks that
// - the comparer detects some trivial flaws
// - if one chan exits earlier, we don't halt
func TestDifferentLengthOutput(t *testing.T) {
	var (
		gethChan   = make(chan *vm.StructLog)
		parityChan = make(chan *vm.StructLog)
	)

	// Set up geth
	{
		output, err := os.Open("testdata/statetest1_geth_stderr.jsonl")
		if err != nil {
			t.Fatal(err)
		}
		go func(file *os.File) {
			vm := NewGethEVM("")
			vm.wg.Add(1)
			vm.feed(file, gethChan)
		}(output)
	}
	// Set up parity
	{
		output, err := os.Open("testdata/statetest1_parity_stderr.jsonl")
		if err != nil {
			t.Fatal(err)
		}
		go func(file *os.File) {
			evm := NewParityVM("")
			evm.wg.Add(1)
			filter := make(chan *vm.StructLog)
			go func() {
				// This little filter drops the first two items
				// Therefore, the errors will be
				// 3 errors from the ops,
				// + 1 error from different lengths
				i := 0
				for item := range filter {
					if i > 2 {
						parityChan <- item
					}
					i++
				}
				close(parityChan)
			}()
			evm.feed(file, filter)
		}(output)
	}

	// Now we have two channels that will spit out OutputItems
	c := &Comparer{}
	outCh := c.CompareVms(gethChan, parityChan)
	expErrors := 4
	errors := 0
	for range outCh {
		//fmt.Printf("diff: %v\n", diff)
		errors++
	}
	if expErrors != errors {
		t.Errorf("got %d errors, expected %d", errors, expErrors)
	}
}
