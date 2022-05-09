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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/holiman/goevmlab/traces"
	"github.com/holiman/goevmlab/ui"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "filename")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, `
Reads the given trace-file, and displays the tracing in a nice CLI user interface`)
	}
}

func main() {
	testTraces := []string{
		"../../traces/testdata/geth_nomemory.jsonl",
		"../../traces/testdata/geth_memory.jsonl",
		"../../traces/testdata/geth_traceTransaction.json",
		"../../traces/testdata/14a4a43b4e9759aac86bb0ae7e5926850406ff1c43ea571239563ff781474ae0.json.snappy",
	}

	hasChunking := flag.Bool("chunking", false, "enable code chunking info in traceview")
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
	ui.NewViewManager(trace, &ui.Config{HasChunking: *hasChunking})
}
