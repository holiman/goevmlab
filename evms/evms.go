package evms

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/vm"
)

// The Evm interface represents external EVM implementations, which can
// be e.g. docker instances or binaries
type Evm interface {
	// StartStateTest runs the statetest on the underlying EVM. It returns a channel
	// where the trace-output is delivered.
	StartStateTest(path string) (chan *vm.StructLog, error)
	//Open() // Preparare for execution
	Close() // Tear down processes
}

// logString provides a human friendly string
func logString(log *vm.StructLog) string {
	return fmt.Sprintf("pc: %3d op: %18v depth: %2v gas: %5d stack size %d",
		log.Pc, log.Op, log.Depth, log.Gas, len(log.Stack))

}
func DiffLogs(a, b *vm.StructLog) string {
	if a.Pc != b.Pc {
		return fmt.Sprintf("pc %d != %d", a.Pc, b.Pc)
	}
	if a.Op != b.Op {
		return fmt.Sprintf("op %d != %d", a.Op, b.Op)
	}
	if a.Depth != b.Depth {
		return fmt.Sprintf("depth %d != %d", a.Depth, b.Depth)
	}
	if a.Gas != b.Gas {
		return fmt.Sprintf("gas %d != %d", a.Gas, b.Gas)
	}
	// Parity seems to be lacking gasCost
	//if a.GasCost != b.GasCost {
	//	return fmt.Sprintf("gasCost %d != %d", a.GasCost, b.GasCost)
	//}
	if len(a.Stack) != len(b.Stack) {
		return fmt.Sprintf("stack size %d != %d", len(a.Stack), len(b.Stack))

	}
	for i, item := range a.Stack {
		if b.Stack[i].Cmp(item) != 0 {
			return fmt.Sprintf("stack item %d, %x != %x", i, item, b.Stack[i])
		}
	}
	return ""
}

type Comparer struct {
	Steps    int
	MaxDepth int
}

// CompareVMs compares the outputs from the channels, returns a channel with
// error info
func (c *Comparer) CompareVms(a, b chan *vm.StructLog) chan string {
	output := make(chan string)

	go func() {
		// This whole thing is ugly. Needs to be rewritten
		for {
			var (
				op1, op2     *vm.StructLog
				more1, more2 bool
			)
			select {
			case op1, more1 = <-a:
				op2, more2 = <-b
			case op2, more2 = <-b:
				op1, more1 = <-a
			}
			if more1 != more2 {
				output <- fmt.Sprintf("Channel a done: %v, chan b done: %v", !more1, !more2)
				fmt.Printf("op1 %v op2 %v\n", op1, op2)

			}
			if !(more1 && more2) {
				close(output)
				return
			}
			if diff := DiffLogs(op1, op2); len(diff) != 0 {
				info := fmt.Sprintf("Diff detected, step %d: %v\n\t%v\n\t%v\n", c.Steps, diff, logString(op1), logString(op2))
				output <- info
			}
			c.Steps++
			if depth := op1.Depth; depth > c.MaxDepth {
				c.MaxDepth = depth
			}
		}

	}()
	return output
}
