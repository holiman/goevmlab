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
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/urfave/cli/v2"
)

var fullTraceFlag = &cli.BoolFlag{
	Name: "fulltrace",
	Usage: "If set to true, the minimizer examines the full trace, instead of just " +
		"looking for a differing stateroot.",
}

var patienceFlag = &cli.UintFlag{
	Name:  "patience",
	Usage: "If set to a high value, the minmizer will spend more time retrying it's various minimization routines",
	Value: 5,
}

var skipGasFlag = &cli.BoolFlag{
	Name:  "skipgas",
	Usage: "If true, gas minimization is skipped",
	Value: false,
}

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Test-case minimizer"
	app.Flags = append(app.Flags, common.VMFlags...)
	app.Flags = append(app.Flags, fullTraceFlag, patienceFlag, skipGasFlag)
	app.Flags = append(app.Flags, common.VerbosityFlag)
	app.Flags = append(app.Flags, common.LocationFlag)
	app.Action = startFuzzer
	return app
}

var app = initApp()

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startFuzzer(c *cli.Context) error {
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
		patience  = c.Int(patienceFlag.Name)
		vms       = common.InitVMs(c)
		outdir    = c.String(common.LocationFlag.Name)
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
	if len(vms) == 1 {
		log.Info("Only one VM supplied, entering crash-mode (assuming the vm crashes)")
		// In this mode, we just execute one VM which crashes.
		evm := vms[0]
		compareFn = func(path string, c *cli.Context) (bool, error) {
			if _, err := evm.RunStateTest(path, io.Discard, false); err != nil {
				log.Info("EVM crash occurred", "err", err)
				return false, nil
			}
			log.Info("No crash occurred")
			return true, nil
		}
	} else if c.Bool(fullTraceFlag.Name) {
		compareFn = func(path string, c *cli.Context) (bool, error) {
			agree, err := common.RunSingleTest(path, outdir, vms)
			if !agree {
				return false, nil
			}
			return true, err
		}
	}
	if consensus, err := compareFn(testPath, c); err != nil {
		return err
	} else if consensus {
		msg := "No consensus failure -- the input statetest needs to be a test which produces a difference"
		if !c.Bool(fullTraceFlag.Name) {
			msg = "No consensus failure -- the input statetest needs to be a test which produces a difference.\n" +
				"(Perhaps retry with --fulltrace enabled?)"
		}
		return errors.New(msg)
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
		if !allAgree {
			log.Info("Change ok")
			if err := os.WriteFile(good, data, 0777); err != nil {
				panic(err)
			}
		} else {
			log.Info("Bad change, clients in consensus - reverting")
		}
		return allAgree
	}

	// Try removing auths
	if len(gst[testname].Tx.AuthorizationList) > 0 {
		log.Info("Reducing auths")
		for i := 0; ; {
			current := gst[testname].Tx.AuthorizationList
			if i >= len(current) {
				break
			}
			gst[testname].Tx.AuthorizationList = fuzzing.CopyAndDropAuth(current, i)
			log.Info("Dropped auth", "index", i)
			if !inConsensus() {
				continue
			}
			log.Info("Restoring change")
			gst[testname].Tx.AuthorizationList = current
			i++
		}
	}

	// Try decreasing gas
	if !c.Bool(skipGasFlag.Name) {
		gas := sort.Search(int(gst[testname].Tx.GasLimit[0]), func(i int) bool {
			gst[testname].Tx.GasLimit[0] = uint64(i)
			log.Info("Mutating gas", "value", i)
			return !inConsensus()
		})
		// And restore the gas again
		gst[testname].Tx.GasLimit[0] = uint64(gas)
	}

	// Try removing accounts
	for target, acc := range gst[testname].Pre {
		delete(gst[testname].Pre, target)
		log.Info("Removing account", "target", target)
		if !inConsensus() {
			continue
		}
		log.Info("Restoring", "target", target)
		gst[testname].Pre[target] = acc
	}
	mutators := []codeMutator{
		newCodeShortener(),
		newBalancedCodeMutator(),
		newCodeRandomMutator(),
	}
	for i, m := range mutators {
		for target, acc := range gst[testname].Pre {
			if len(acc.Code) == 0 {
				continue
			}
			m.reset(acc.Code)
			log.Info("Reducing code", "mutator", i, "target", target)

			for fails := 0; fails < patience; {
				if exhausted := m.proceed(); exhausted {
					break
				}
				acc := gst[testname].Pre[target]
				acc.Code = m.code()
				gst[testname].Pre[target] = acc
				if !inConsensus() {
					fails = 0
					continue
				}
				log.Info("Restoring change")
				fails++
				m.undo()
				acc.Code = m.code()
				gst[testname].Pre[target] = acc
			}
		}
	}

	log.Info("Reducing slots")
	// Try removing storage
	for target, acc := range gst[testname].Pre {
		for k, v := range acc.Storage {
			delete(gst[testname].Pre[target].Storage, k)
			log.Info("Reducing slot", "target", target, "slot", k)
			if !inConsensus() {
				continue
			}
			log.Info("Restoring change")
			gst[testname].Pre[target].Storage[k] = v
		}
	}

	log.Info("Done", "result", good)
	return nil
}
