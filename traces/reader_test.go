package traces

import (
	"fmt"
	"github.com/holiman/goevmlab/ops"

	//"github.com/holiman/goevmlab/ops"
	"io/ioutil"
	"path"
	"testing"
)

var (
	testDir = "../testdata/traces"
)

// Tries to read all files in testdata/traces
func TestReaderBasics(t *testing.T) {
	files, err := ioutil.ReadDir(testDir)
	if err != nil {
		t.Fatalf("error reading files: %v", err)
	}
	for _, f := range files {
		p := path.Join(testDir, f.Name())
		_, err = ReadTrace(p)
		if err != nil {
			t.Errorf("err reading %v: %v", p, err)
		}
		fmt.Printf("read %v\n", p)
	}
}

func TestParityReading(t *testing.T) {
	p := path.Join(testDir, "parity_1352.jsonl")
	traces, err := ReadTrace(p)
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
	traces, err := ReadTrace(p)
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

func TestParityVsGeth(t *testing.T) {
	pTrace, err := ReadTrace(path.Join(testDir, "parity_1352.jsonl"))
	if err != nil {
		t.Fatalf("err reading  parity: %v", err)
	}
	gTrace, err := ReadTrace(path.Join(testDir, "geth_1352.jsonl"))
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
			if item.Cmp(pLog.Stack[i]) != 0 {
				t.Errorf("step %d, stack item %d diff: %v != %v", step, i, item, pLog.Stack[i])
			}
		}
	}
}
