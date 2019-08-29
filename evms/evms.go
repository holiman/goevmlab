package evms

import (
	"bytes"
	"fmt"
)

// The Evm interface represents external EVM implementations, which can
// be e.g. docker instances or binaries
type Evm interface {
	// StartStateTest runs the statetest on the underlying EVM. It returns a channel
	// where the trace-output is delivered.
	StartStateTest(path string) (chan *OutputItem, error)
	//Open() // Preparare for execution
	Close() // Tear down processes
}

// OutputItem is what the EVMs spit out, representing an event during the trace
type OutputItem map[string]interface{}

// Diff returns a map containing differences between o and other
func (o *OutputItem) Diff(other *OutputItem) *map[string][]interface{} {

	var diff *map[string][]interface{}

	addDiff := func(name string, diffs ...interface{}) {
		if diff == nil {
			d := make(map[string][]interface{})
			diff = &d
		}
		(*diff)[name] = diffs
	}

	// The output items either are regular ops, or they contain the stateroot
	// Path one: regular op
	_, aRegular := (*o)["pc"]
	_, bRegular := (*other)["pc"]

	if aRegular && bRegular {
		// check pc, op, gas, depth
		for _, label := range []string{"pc", "op", "gas", "depth"} {
			a := (*o)[label]
			b := (*other)[label]
			if a != b {
				addDiff(label, a, b)
			}
		}
		// check stack
		stackA := fmt.Sprintf("%v", (*o)["stack"])
		stackB := fmt.Sprintf("%v", (*other)["stack"])
		if stackA != stackB {
			addDiff("stack", stackA, stackB)
		}
		return diff
	}

	_, aStateRoot := (*o)["stateRoot"]
	_, bStateRoot := (*other)["stateRoot"]

	if aStateRoot && bStateRoot {
		// Check stateroot
		for _, label := range []string{"stateRoot"} {
			a := (*o)[label]
			b := (*other)[label]
			if a == b {
				addDiff(label, a, b)
			}
		}
		return diff
	}
	return nil
}

// CompareVMs compares the outputs from the channels, returns a channel with
// error info
func CompareVms(a, b chan *OutputItem) chan string {
	output := make(chan string)

	go func() {
		for {
			var (
				op1, op2     *OutputItem
				more1, more2 bool
			)
			select {
			case op1, more1 = <-a:
				op2, more2 = <-b
			case op2, more2 = <-b:
				op1, more1 = <-a
			}
			if more1 != more2 {
				output <- fmt.Sprintf("Channel a: %v, chan b: %v", more1, more2)
			}
			if !(more1 && more2) {
				close(output)
				return
			}
			if diff := op1.Diff(op2); diff != nil {
				var info bytes.Buffer
				info.WriteString("Error:\n")
				for k, v := range *diff {
					info.WriteString(fmt.Sprintf("\t%v: %v != %v\n", k, v[0], v[1]))
				}
				output <- info.String()
			}
		}

	}()
	return output
}
