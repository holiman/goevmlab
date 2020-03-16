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
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/evms"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestGenerator(t *testing.T) {
	st := GenerateStateTest("randoTest")

	_, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		t.Fatal(err)
	}
}

func TestBlakeGenerator(t *testing.T) {
	st := GenerateBlakeTest("randoTest")

	_, err := json.MarshalIndent(st, "", " ")
	if err != nil {
		t.Fatal(err)
	}
}

var binMu sync.Mutex

// We need to decompress the evm binaries prior to execution
func setupBinaries(t *testing.T) {
	binMu.Lock()
	defer binMu.Unlock()

	err := filepath.Walk("../binaries", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".zip") {
			// ignore
			return nil
		}
		destPath := strings.TrimSuffix(path, ".zip")
		if _, err := os.Stat(destPath); err == nil {
			// Already exists, skip
			return nil
		}
		var src io.Reader
		fmt.Printf("Uncompressing %v into %v \n", path, destPath)
		{ // Decompress
			zr, err := zip.OpenReader(path)
			if err != nil {
				return fmt.Errorf("failed opening reader: %v", err)
			}
			f, err := zr.File[0].Open()
			if err != nil {
				return fmt.Errorf("failed opening file inside archive: %v", err)
			}
			defer f.Close()
			src = f
		}
		dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			return err
		}
		defer dest.Close()
		size, err := io.Copy(dest, src)
		if err != nil {
			return err
		}
		fmt.Printf("Uncompressed %v OK (%d bytes)\n", path, size)
		return nil
	})
	if err != nil {
		t.Fatalf("error constructing evm binaries: %v", err)
	}
}

func testCompare(a, b evms.Evm, testfile string) error {

	wa := bytes.NewBuffer(nil)
	if _, err := a.RunStateTest(testfile, wa, false); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	wb := bytes.NewBuffer(nil)
	if _, err := b.RunStateTest(testfile, wb, false); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	eq := evms.CompareFiles([]evms.Evm{a, b}, []io.Reader{wa, wb})
	if !eq {
		return fmt.Errorf("diffs encountered")
	}
	return nil
}

func testFuzzing(t *testing.T) error {
	setupBinaries(t)

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

	geth := evms.NewGethEVM("../binaries/evm")
	parity := evms.NewParityVM("../binaries/parity-evm")

	return testCompare(geth, parity, p)

}

func TestBlake(t *testing.T) {
	setupBinaries(t)
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
	geth := evms.NewGethEVM("../binaries/evm")
	parity := evms.NewParityVM("../binaries/parity-evm")

	if err := testCompare(geth, parity, p); err != nil {
		t.Fatal(err)
	}
}

func TestFuzzing(t *testing.T) {
	if err := testFuzzing(t); err != nil {
		t.Fatal(err)
	}
}

//func TestFuzzingCoverage(t *testing.T) {
//	tot := 0
//	totDepth := 0
//	for i := 0; i < 100; i++ {
//		numSteps, maxdepth, _ := testFuzzing(t)
//		tot += numSteps
//		fmt.Printf("numSteps %d maxDepth: %d\n", numSteps, maxdepth)
//		totDepth += (maxdepth - 1)
//	}
//	fmt.Printf("total steps (100 tests): %d, total depth %d\n", tot, totDepth)
//}

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
		if _, err := json.Marshal(st); err != nil {
			b.Fatal(err)
		}
	}
}
