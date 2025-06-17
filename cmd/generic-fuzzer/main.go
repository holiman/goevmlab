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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/ops"
	"github.com/urfave/cli/v2"
)

var (
	engineFlag = &cli.StringSliceFlag{
		Name:  "engine",
		Usage: "fuzzing-engine",
		Value: cli.NewStringSlice(fuzzing.FactoryNames()...),
	}
	forkFlag = &cli.StringFlag{
		Name:  "fork",
		Usage: fmt.Sprintf("Fork to use %v", ops.ForkNames()),
		Value: "Prague",
	}
	app = initApp()
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Fuzzer with various targets"
	app.Flags = append(app.Flags, common.VMFlags...)
	app.Flags = append(app.Flags,
		common.SkipTraceFlag,
		common.ThreadFlag,
		common.LocationFlag,
		engineFlag,
		forkFlag,
		common.VerbosityFlag,
		common.NotifyFlag,
		common.RemoveFilesFlag,
		common.RawDebugFlag,
	)
	app.Action = startFuzzer
	return app
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startFuzzer(ctx *cli.Context) (err error) {
	if topic := ctx.String(common.NotifyFlag.Name); topic != "" {
		_, _ = http.Post(fmt.Sprintf("https://ntfy.sh/%v", topic), "text/plain",
			strings.NewReader("Fuzzer starting"))
	}
	loglevel := slog.Level(ctx.Int(common.VerbosityFlag.Name))
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, loglevel, true)))
	log.Root().Write(loglevel, "Set loglevel", "level", loglevel)
	var (
		fNames = ctx.StringSlice(engineFlag.Name)
		fork   = ctx.String(forkFlag.Name)
	)
	if len(fNames) == 0 {
		fmt.Printf("At least one fuzzer engine needed. ")
		fmt.Printf("Available targets: %v\n", fuzzing.FactoryNames())
		return errors.New("missing engine")
	}
	var factory common.GeneratorFn
	if len(fNames) == 1 {
		factory = fuzzing.Factory(fNames[0], fork)
		if factory == nil {
			return fmt.Errorf("unknown target %v", fNames[0])
		}
	} else {
		// Need to put together a meta-factory
		var factories []common.GeneratorFn
		for _, fName := range fNames {
			if f := fuzzing.Factory(fName, fork); f == nil {
				return fmt.Errorf("unknown target %v", fName)
			} else {
				factories = append(factories, f)
			}
			log.Info("Added factory", "name", fName)
		}
		var index atomic.Uint64
		factory = func() *fuzzing.GstMaker {
			i := int(index.Add(1))
			i %= len(factories)
			fn := factories[i]
			return fn()
		}
	}
	return common.GenerateAndExecute(ctx, factory, "mixed")
}
