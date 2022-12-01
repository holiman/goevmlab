// Copyright 2019 Martin Holst Swende
// This file is part of the go-evmlab library.
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
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/urfave/cli/v2"
	"io"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Generator for tests targeting SSTORE and SLOAD"
	return app
}

var (
	app = initApp()
)

func init() {
	app.Flags = []cli.Flag{
		common.PrefixFlag,
		common.LocationFlag,
		common.CountFlag,
		common.TraceFlag,
	}
	app.Commands = []*cli.Command{
		generateCommand,
	}
}

var generateCommand = &cli.Command{
	Action:      generate,
	Name:        "generate",
	Usage:       "generate tests",
	ArgsUsage:   "<number of tests to generate>",
	Description: `The generate generates the tests.`,
}

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	log.Info("Generating tests")
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func createTests(location, prefix, fork string, limit int, trace bool) error {
	log.Info("Generating tests", "location", location, "prefix", prefix, "fork", fork, "limit", limit, "tracing", trace)
	createFn := fuzzing.Factory2200("London")
	for i := 0; i < limit; i++ {
		testName := fmt.Sprintf("%vstoragetest-%04d", prefix, i)
		fileName := fmt.Sprintf("%v.json", testName)
		p := path.Join(location, fileName)
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		close := func() {
			f.Close()
		}
		// Now, let's also dump out the trace, so we can investigate if the tests
		// are doing anything interesting
		var traceOutput io.Writer
		if trace {
			traceName := fmt.Sprintf("%v-trace.jsonl", testName)
			traceOut, err := os.OpenFile(path.Join(location, traceName), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
			if err != nil {
				return err
			}
			traceOutput = traceOut
			close = func() {
				f.Close()
				traceOut.Close()
			}
		}
		// Generate new code
		base := createFn()
		// Get new state root and logs hash
		if err := base.Fill(traceOutput); err != nil {
			close()
			return err
		}
		test := base.ToGeneralStateTest(testName)
		// Write to tfile
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", " ")
		if err = encoder.Encode(test); err != nil {
			close()
			return err
		}
		close()
	}
	return nil
}

func generate(ctx *cli.Context) error {
	var (
		fork     = "London"
		prefix   = ""
		count    = ctx.Int(common.CountFlag.Name)
		location = ctx.String(common.LocationFlag.Name)
	)
	if err := os.MkdirAll(location, 0755); err != nil {
		return fmt.Errorf("could not create %v: %v", location, err)
	}
	if ctx.IsSet(common.PrefixFlag.Name) {
		prefix = fmt.Sprintf("%v-", ctx.String(common.PrefixFlag.Name))
	}
	return createTests(location, prefix, fork, count, ctx.Bool(common.TraceFlag.Name))
}
