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
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/ops"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Martin Holst Swende"
	app.Usage = "Test-case minimizer"
	app.Flags = append(app.Flags, common.VmFlags...)
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

type codeMutator struct {
	current  []byte
	lastGood []byte
}

// proceed tells the mutator to continue one mutation
func (m *codeMutator) proceed() {
	m.lastGood = m.current
	// Now mutate current
	var next []byte
	max := ops.InstructionCount(m.lastGood)
	removed := 0
	for removed == 0 {
		cutPoint := rand.Intn(max)
		next = make([]byte, 0)
		it := ops.NewInstructionIterator(m.lastGood)
		for it.Next() {
			if removed == 0 && it.PC() > uint64(cutPoint) {
				// Remove until the stack balances out
				delta := 0
				for {
					removed += 1
					delta += it.Op().Stackdelta()
					if delta == 0 {
						break
					}
					// Skip next one too
					if !it.Next() {
						break
					}
				}
			}
			next = append(next, byte(it.Op()))
			if arg := it.Arg(); arg != nil {
				next = append(next, arg...)
			}
		}
	}
	fmt.Printf("code length: %d (was %d)\n", len(next), len(m.lastGood))
	m.current = next
}

// undo tells the mutator to revert the last change
func (m *codeMutator) undo() {
	m.current = m.lastGood
}

func startFuzzer(c *cli.Context) error {

	if c.NArg() != 1 {
		return fmt.Errorf("input state test file needed")
	}
	testPath := c.Args().First()

	consensus, err := common.RootsEqual(testPath, c)

	if err != nil {
		return err
	}
	if consensus {
		return errors.New("No consensus failure -- " +
			"the input statetest needs to be a test which produces differing stateroot")

	}
	gst, err := fuzzing.FromGeneralStateTest(testPath)
	if err != nil {
		return err
	}
	gst2 := (*gst)

	var testname string
	for t := range gst2 {
		testname = t
		break
	}

	for target, acc := range gst2[testname].Pre {
		if len(acc.Code) > 0 {
			fmt.Printf("Target: %x\n", target)
		} else {
			continue
		}
		//target := common2.HexToAddress("0x00000000000000000000000000ca110b15012381")
		code := acc.Code
		if err != nil {
			return err
		}
		m := codeMutator{current: code, lastGood: code}
		// Alright, we're in business
		i := 0
		good := fmt.Sprintf("%v.min", testPath)
		for {
			out := fmt.Sprintf("%v.%v", testPath, "tmp")
			m.proceed()
			acc := gst2[testname].Pre[target]
			acc.Code = m.current
			gst2[testname].Pre[target] = acc
			data, _ := json.MarshalIndent(gst2, "", "  ")
			if err := ioutil.WriteFile(out, data, 0777); err != nil {
				return err
			}
			inConsensus, err := common.RootsEqual(out, c)
			if err != nil {
				return err
			}
			if !inConsensus {
				fmt.Print("still failing!")
				i++
				//good = fmt.Sprintf("%v.good.%v", testPath, "latest")
				if err := ioutil.WriteFile(good, data, 0777); err != nil {
					return err
				}
			} else {
				fmt.Printf("oops, broke it\n")
				m.undo()
			}
		}
	}
	// Find some account with code for minimization
	return nil
}
