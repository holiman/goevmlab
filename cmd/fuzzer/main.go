package main

import (
	"fmt"
	"github.com/holiman/goevmlab/fuzzing"
	"os"

	"github.com/holiman/goevmlab/evms"
)

func testEvm(vm evms.Evm, testfile string) {
	ch, err := vm.StartStateTest(testfile)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	for item := range ch {
		fmt.Printf("data: \n%v\n", item)
	}
	fmt.Printf("exiting")
	vm.Close()
}

func testCompare(a, b evms.Evm, testfile string) {
	chA, err := a.StartStateTest(testfile)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	chB, err := b.StartStateTest(testfile)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	outCh := evms.CompareVms(chA, chB)
	for outp := range outCh {
		fmt.Printf("Output: %v\n", outp)
	}
}

func testBlake(){

	geth := evms.NewGethEVM("/home/user/go/src/github.com/ethereum/go-ethereum/build/bin/evm")
	file := "/home/user/go/src/github.com/holiman/goevmlab/evms/testdata/blaketest1.json"
	fuzzing.GenerateStateTest("blaketest")

}

func main() {

	// generate a test

	//file := "/home/user/workspace/tests/GeneralStateTests/stPreCompiledContracts2/CALLBlake2f.json"
	file := "/home/user/go/src/github.com/holiman/goevmlab/evms/testdata/statetest1.json"
	geth := evms.NewGethEVM("/home/user/go/src/github.com/ethereum/go-ethereum/build/bin/evm")
	parity := evms.NewParityVM("/home/user/go/src/github.com/holiman/goevmlab/parity-evm")
	testCompare(geth, parity, file)
	fmt.Printf("Done\n")
}
