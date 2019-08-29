package main

import (
	"fmt"
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

func main() {

	//file := "/home/user/workspace/tests/GeneralStateTests/stPreCompiledContracts2/CALLBlake2f.json"
	file := "/home/user/go/src/github.com/holiman/goevmlab/testdata/statetests/0003--randomStatetestmartin-Wed_10_02_29-14338-0-3-test.json"

	geth := evms.NewGethEVM("/home/user/go/src/github.com/ethereum/go-ethereum/build/bin/evm")
	//testEvm(geth, file)
	parity := evms.NewParityVM("/home/user/go/src/github.com/holiman/goevmlab/parity-evm")
	//testEvm(parity, file)
	testCompare(geth, parity, file)
}
