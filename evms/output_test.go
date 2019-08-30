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

	// Now we have two channels that will spit out OutputItems
	outCh := CompareVms(gethChan, parityChan)
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
	outCh := CompareVms(gethChan, parityChan)
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
