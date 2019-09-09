package main

import (
	"flag"
	"fmt"
	"github.com/holiman/goevmlab/traces"
	"github.com/holiman/goevmlab/ui"
	"os"
	"strconv"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "filename")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Reads the given trace-file, and displays the tracing in a nice CLI user interface
`)
	}
}

func main() {

	testTraces := []string{
		"../../traces/testdata/geth_nomemory.jsonl",
		"../../traces/testdata/geth_memory.jsonl",
		"../../traces/testdata/geth_traceTransaction.json",
		"../../traces/testdata/14a4a43b4e9759aac86bb0ae7e5926850406ff1c43ea571239563ff781474ae0.json.snappy",
	}

	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Printf("Expected one argument\n")
		flag.Usage()
		os.Exit(1)
	}

	fName := flag.Arg(0)
	// Some debugging help here
	if n, err := strconv.Atoi(fName); err == nil {
		if n < len(testTraces) {
			fName = testTraces[n]
		}
	}
	trace, err := traces.ReadFile(fName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	mgr := ui.NewViewManager(trace)
	mgr.Run()
}
