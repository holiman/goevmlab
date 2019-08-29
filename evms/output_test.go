package evms

import (
	"os"
	"testing"
)

// TestGethVsParityStatetestOutput in this test, we simulate two external
// vms, using printouts from actual evm binaries. The two outputs should
// not produce any differences
func TestGethVsParityStatetestOutput(t *testing.T) {
	var (
		gethChan   = make(chan *OutputItem)
		parityChan = make(chan *OutputItem)
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
		output, err := os.Open("testdata/statetest1_geth_stderr.jsonl")
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
