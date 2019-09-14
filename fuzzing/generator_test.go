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

package fuzzing

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/evms"
	"os"
	"path"
	"testing"
)

func TestGenerator(t *testing.T) {
	st := GenerateStateTest("randoTest")

	data, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("data: \n")
	fmt.Printf(string(data))

}

func TestBlakeGenerator(t *testing.T) {
	st := GenerateBlakeTest("randoTest")

	data, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("data: \n")
	fmt.Printf(string(data))

}
func testCompare(a, b evms.Evm, testfile string) (int, int, error) {
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
	c := &evms.Comparer{}
	outCh := c.CompareVms(chA, chB)
	errors := 0

	for outp := range outCh {
		fmt.Printf("Output: %v\n", outp)
		errors++
	}
	if errors > 0 {
		return c.Steps, c.MaxDepth, fmt.Errorf("%d diffs encountered", errors)
	}
	return c.Steps, c.MaxDepth, nil
}

func testFuzzing(t *testing.T) (int, int, error) {
	testName := "some_rando_test"
	fileName := fmt.Sprintf("%v.json", testName)
	p := path.Join(os.TempDir(), fileName)
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		t.Fatal(err)
	}
	gst := GenerateStateTest(testName)
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", " ")
	if err = encoder.Encode(gst); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	geth := evms.NewGethEVM("/home/user/go/src/github.com/ethereum/go-ethereum/build/bin/evm")
	parity := evms.NewParityVM("/home/user/go/src/github.com/holiman/goevmlab/parity-evm")

	return testCompare(geth, parity, p)

}

func TestBlake(t *testing.T) {
	testName := "blake_test"
	fileName := fmt.Sprintf("%v.json", testName)
	p := path.Join(os.TempDir(), fileName)
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("file is %v \n", p)
	gst := GenerateBlakeTest(testName)
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", " ")
	if err = encoder.Encode(gst); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	geth := evms.NewGethEVM("/home/user/go/src/github.com/ethereum/go-ethereum/build/bin/evm")
	parity := evms.NewParityVM("/home/user/go/src/github.com/holiman/goevmlab/parity-evm")

	testCompare(geth, parity, p)

}

func TestFuzzing(t *testing.T) {
	testFuzzing(t)
}

func TestFuzzingCoverage(t *testing.T) {
	tot := 0
	totDepth := 0
	for i := 0; i < 100; i++ {
		numSteps, maxdepth, _ := testFuzzing(t)
		tot += numSteps
		fmt.Printf("numSteps %d maxDepth: %d\n", numSteps, maxdepth)
		totDepth += (maxdepth - 1)
	}
	fmt.Printf("total steps (100 tests): %d, total depth %d\n", tot, totDepth)
}

/*
BenchmarkGenerator-6   	  500000	      2419 ns/op

BenchmarkGenerator-6   	  500000	      3092 ns/op

# randomizing code from valid opcodes, with pushdata insertion and stack balance
BenchmarkGenerator-6   	   10000	    118339 ns/op

# using program, smart calldata
BenchmarkGenerator-6   	   20000	     60932 ns/op

# using blake2 generator
BenchmarkGenerator-6   	  100000	     13638 ns/op

# blake2, but only doing new randcall
BenchmarkGenerator-6   	  200000	      8413 ns/op
*/
func BenchmarkGenerator(b *testing.B) {
	t := GenerateBlake()
	alloc := *t.pre
	target := common.HexToAddress(t.tx.To)
	dest := alloc[target]
	for i := 0; i < b.N; i++ {
		dest.Code = RandCallBlake()
		t.ToGeneralStateTest("rando")
	}
}

func BenchmarkGeneratorWithMarshalling(b *testing.B) {
	/*
		With indent:
		BenchmarkGeneratorWithMarshalling-6   	   30000	     38208 ns/op
		Without indent:
		BenchmarkGeneratorWithMarshalling-6   	   50000	     22710 ns/op
	*/

	for i := 0; i < b.N; i++ {
		st := GenerateStateTest("randoTest")
		json.Marshal(st)
	}
}
