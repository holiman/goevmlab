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

package traces

import (
	"fmt"
	"path"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/ops"
	"os"
)

var (
	testDir = "testdata/"
)

// Tries to read all files in testdata/traces
func TestReaderBasics(t *testing.T) {
	files, err := os.ReadDir(testDir)
	if err != nil {
		t.Fatalf("error reading files: %v", err)
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "jsonl") ||
			strings.HasSuffix(f.Name(), "json") ||
			strings.HasSuffix(f.Name(), "snappy") {
			p := path.Join(testDir, f.Name())
			_, err = ReadFile(p)
			if err != nil {
				t.Errorf("err reading %v: %v", p, err)
			}
			//fmt.Printf("read %v\n", p)
		} else {
			t.Logf("skipped %v\n", f.Name())
		}
	}
}

func TestParityReading(t *testing.T) {
	p := path.Join(testDir, "parity_1352.jsonl")
	traces, err := ReadFile(p)
	if err != nil {
		t.Fatalf("err reading %v: %v", p, err)
	}
	{
		exp := 62
		if got := len(traces.Ops); got != exp {
			t.Fatalf("trace length wrong, got %d, expected %d", got, exp)
		}
	}
	{
		exp := 5
		if got := len(traces.Get(5).Stack()); got != exp {
			fmt.Printf("%v", traces.Get(0).log)
			t.Fatalf("stack length wrong, got %d, expected %d", got, exp)
		}
	}
}

func TestTraceTransactionReading(t *testing.T) {
	p := path.Join(testDir, "geth_traceTransaction.json")
	traces, err := ReadFile(p)
	if err != nil {
		t.Fatalf("err reading %v: %v", p, err)
	}
	{
		exp := 388
		if got := len(traces.Ops); got != exp {
			t.Fatalf("trace length wrong, got %d, expected %d", got, exp)
		}
	}
	{
		exp := 2
		if got := len(traces.Get(5).Stack()); got != exp {
			t.Fatalf("stack length wrong, got %d, expected %d", got, exp)
		}
	}

	{
		exp := byte(ops.PUSH1)
		if got := traces.Get(0).Op(); got != exp {
			t.Fatalf("op wrong, got %d, expected %d", got, exp)
		}
	}
	{
		exp := "PUSH1"
		if got := traces.Get(0).Get("opName"); got != exp {
			t.Fatalf("op wrong, got %v, expected %v", got, exp)
		}
	}
	{
		// There's an MSTORE at step 2, so mem at step four contains dat
		expMemSize := 96
		if got := len(traces.Get(3).Memory()); got != expMemSize {
			t.Fatalf("mem wrong, expected size %d, got %d", expMemSize, got)
		}
		if data := traces.Get(3).Memory()[0x5f]; data != 0x60 {
			t.Fatalf("mem wrong, expected 60 at byte %d, got %d", 0x5f, data)
		}
	}
}

// Tests the file 14a4a43b4e9759aac86bb0ae7e5926850406ff1c43ea571239563ff781474ae0.json,
// which was obtained from
// https://github.com/aragon/aragonOS/issues/549
func TestSecondTraceTxReading(t *testing.T) {
	p := path.Join(testDir, "14a4a43b4e9759aac86bb0ae7e5926850406ff1c43ea571239563ff781474ae0.json.snappy")
	traces, err := ReadFile(p)
	if err != nil {
		t.Fatalf("err reading %v: %v", p, err)
	}
	{
		exp := 13024
		if got := len(traces.Ops); got != exp {
			t.Fatalf("trace length wrong, got %d, expected %d", got, exp)
		}
	}
	{
		exp := byte(0x61)
		if got := traces.Get(12257).Op(); got != exp {
			t.Fatalf("op wrong, got %x, expected %x", got, exp)
		}
	}

	{
		exp := common.HexToAddress("0x3c307fefd3d71c3ca8a3c26539ef4d47c61b6565")
		if got := traces.Get(157).address; got == nil || *got != exp {
			t.Fatalf("op wrong, got %x, expected %x", got, exp)
		}
	}
	{ // At step 495, we should be back to depth 1,
		exp := traces.Get(1).address
		if got := traces.Get(495).address; got != exp {
			t.Fatalf("op wrong, got %x, expected %x", got, exp)
		}
	}

	// The example trace is a bit corrupt, reporting wrong gas values.
	// Nothing we can do about it here....
	//{
	//	// Push2 gascost is 3
	//	exp := "3"
	//	if got := traces.Get(12257).Get("gasCost"); got != exp {
	//		t.Fatalf("gascost wrong, got %v, expected %v", got, exp)
	//	}
	//
	//}
}

func TestParityVsGeth(t *testing.T) {
	pTrace, err := ReadFile(path.Join(testDir, "parity_1352.jsonl"))
	if err != nil {
		t.Fatalf("err reading  parity: %v", err)
	}
	gTrace, err := ReadFile(path.Join(testDir, "geth_1352.jsonl"))
	if err != nil {
		t.Fatalf("err reading geth: %v", err)
	}

	gLen := len(gTrace.Ops)
	pLen := len(pTrace.Ops)
	if gLen != pLen {
		t.Fatalf("geth length %d, parity length %d", gLen, pLen)
	}
	for step := 0; step < len(gTrace.Ops); step++ {
		gLog := gTrace.Get(step).log
		pLog := pTrace.Get(step).log
		if gLog.Op != pLog.Op {
			t.Fatalf("step %d, op %d != %d", step, gLog.Op, pLog.Op)
		}
		if gLog.Depth != pLog.Depth {
			t.Fatalf("step %d, depth %d != %d", step, gLog.Depth, pLog.Depth)
		}

		if len(gLog.Stack) != len(pLog.Stack) {
			t.Fatalf("step %d, stack size %d != %d", step, len(gLog.Stack), len(pLog.Stack))
		}
		for i, item := range gLog.Stack {
			if item != pLog.Stack[i] {
				t.Errorf("step %d, stack item %d diff: %v != %v", step, i, item, pLog.Stack[i])
			}
		}
	}
}

/*
// Convenience func to re-encode some input files
func TestReEncodeTrace(t *testing.T){
	src := path.Join(testDir, "14a4a43b4e9759aac86bb0ae7e5926850406ff1c43ea571239563ff781474ae0.json")
	dst := path.Join(testDir, "14a4a43b4e9759aac86bb0ae7e5926850406ff1c43ea571239563ff781474ae0.json.snappy")
	data, err := os.ReadFile(src)
	if err != nil{
		t.Fatal(err)
	}
	snapdata := snappy.Encode(nil, data)
	err = os.WriteFile(dst, snapdata, 0744)
	if err != nil{
		t.Fatal(err)
	}
}
*/
