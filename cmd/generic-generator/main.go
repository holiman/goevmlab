// Copyright 2022 Martin Holst Swende
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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/urfave/cli/v2"
)

var (
	targetFlag = &cli.StringSliceFlag{
		Name:  "target",
		Usage: "fuzzing-target",
	}
	app = initApp()
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Generator for tests"
	app.Flags = []cli.Flag{
		common.PrefixFlag,
		common.LocationFlag,
		common.CountFlag,
		common.TraceFlag,
		targetFlag,
	}
	app.Action = generate
	return app
}

func main() {
	log.Root().SetHandler(log.LvlFilterHandler(log.LvlInfo, log.StreamHandler(os.Stderr, log.TerminalFormat(true))))
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type config struct {
	fork     string
	prefix   string
	count    int
	location string
	factory  func() *fuzzing.GstMaker
	target   string
	tracing  bool
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
	fNames := ctx.StringSlice(targetFlag.Name)
	// At this point, we only do one at a time
	if len(fNames) == 0 {
		log.Error("At least one fuzzer target needed. ", "available", fuzzing.FactoryNames())
		return errors.New("missing target")
	}
	if len(fNames) > 1 {
		log.Info("Only one target supported\n")
	}
	factory := fuzzing.Factory(fNames[0], "London")
	if factory == nil {
		log.Error("Unknown target needed. ", "target", fNames[0], "available", fuzzing.FactoryNames())
		return fmt.Errorf("unknown target %v", fNames[0])
	}
	return createTests(&config{
		fork:     fork,
		prefix:   prefix,
		count:    count,
		location: location,
		factory:  factory,
		target:   fNames[0],
		tracing:  ctx.Bool(common.TraceFlag.Name),
	})
}

func createTests(conf *config) error {
	log.Info("Generating tests",
		"location", conf.location,
		"prefix", conf.prefix,
		"fork", conf.fork,
		"limit", conf.count,
		"tracing", conf.tracing)
	for i := 0; i < conf.count; i++ {
		testName := fmt.Sprintf("%v%v-%04d", conf.prefix, conf.target, i)
		p := path.Join(conf.location, fmt.Sprintf("%v.json", testName))
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
		if conf.tracing {
			if traceOut, err := os.OpenFile(path.Join(conf.location, fmt.Sprintf("%v-trace.jsonl", testName)),
				os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755); err != nil {
				close()
				return err
			} else {
				traceOutput = traceOut
				close = func() {
					f.Close()
					traceOut.Close()
				}
			}
		}
		// Generate new code
		base := conf.factory()
		// Get new state root and logs hash
		if err := base.Fill(traceOutput); err != nil {
			close()
			return err
		}
		test := base.ToGeneralStateTest(testName)
		// Write to file
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
