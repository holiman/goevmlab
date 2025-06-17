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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/urfave/cli/v2"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Test-case mutator. This app tries to mutate a testcase in order to trigger a differing stateroot. You probably " +
		"should run the `minimizer` first."
	app.Flags = append(app.Flags, common.VMFlags...)
	app.Flags = append(app.Flags, common.VerbosityFlag)
	app.Action = startMutator
	return app
}

var app = initApp()

func main() {

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startMutator(c *cli.Context) error {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr,
		slog.Level(c.Int(common.VerbosityFlag.Name)), true)))

	// TODO: Start a routine to listen for a keypress, to make it possible
	// for the user to skip forward in the mutation process

	if c.NArg() != 1 {
		return fmt.Errorf("input state test file needed")
	}
	var (
		testPath  = c.Args().First()
		compareFn func(path string, c *cli.Context) (bool, error)
	)
	compareFn = func(path string, c *cli.Context) (bool, error) {
		agree, err := common.RootsEqual(path, c)
		// An error here might mean that e.g the gas was changed so that
		// the intrinsic gas was made wrong, and this the tx is invalid.
		// we can ignore that error, and report it as 'agree', triggering the
		// revertal of whatever it was that caused it
		if err != nil {
			return true, nil
		}
		return agree, nil
	}
	if consensus, err := compareFn(testPath, c); err != nil {
		return err
	} else if !consensus {
		return errors.New("consensus failure -- the input statetest already produces a consensus error")
	}

	var (
		gst      fuzzing.GeneralStateTest
		testname string
		good     = fmt.Sprintf("%v.min", testPath)
		out      = fmt.Sprintf("%v.%v", testPath, "tmp")
	)
	if gstPtr, err := fuzzing.FromGeneralStateTest(testPath); err != nil {
		return err
	} else {
		gst = (*gstPtr)
		for t := range gst {
			testname = t
			break
		}
	}

	var inConsensus = func() bool {
		data, _ := json.MarshalIndent(gst, "", "  ")
		if err := os.WriteFile(out, data, 0777); err != nil {
			panic(err)
		}
		allAgree, err := compareFn(out, c)
		if err != nil {
			panic(err)
		}
		if allAgree {
			log.Info("Still in consensus")
			return true
		}
		if err := os.WriteFile(good, data, 0777); err != nil {
			panic(err)
		}
		log.Info("Good change, clients no longer in consensus")
		return false
	}
	var m = newReplacingCodeMutator()
	for target, acc := range gst[testname].Pre {
		if len(acc.Code) == 0 {
			continue
		}
		m.reset(acc.Code)
		log.Info("Mutating code", "target", target)

		for {
			if exhausted := m.proceed(); exhausted {
				break
			}
			acc := gst[testname].Pre[target]
			acc.Code = m.code()
			gst[testname].Pre[target] = acc
			if !inConsensus() {
				break
			}
			m.undo()
			acc.Code = m.code()
			gst[testname].Pre[target] = acc
		}
	}
	log.Info("Done", "result", good)
	return nil
}
