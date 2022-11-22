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
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/ops"
	"github.com/urfave/cli/v2"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
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
// returns 'true' is the mutator is exhausted
func (m *codeMutator) proceed() bool {
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
	return len(next) == len(m.lastGood)
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

	if consensus, err := common.RootsEqual(testPath, c); err != nil {
		return err
	} else if consensus {
		return errors.New("No consensus failure -- " +
			"the input statetest needs to be a test which produces differing stateroot")
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
	var identicalRoots = func() bool {
		data, _ := json.MarshalIndent(gst, "", "  ")
		if err := ioutil.WriteFile(out, data, 0777); err != nil {
			panic(err)
		}
		inConsensus, err := common.RootsEqual(out, c)
		if err != nil {
			panic(err)
		}
		if !inConsensus {
			fmt.Printf("Still ok! (failing testcase) ...\n")
			if err := ioutil.WriteFile(good, data, 0777); err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("Not ok! (clients in consensus)...\n")
		}
		return inConsensus
	}

	// Try decreasing gas
	gas := sort.Search(int(gst[testname].Tx.GasLimit[0]), func(i int) bool {
		gst[testname].Tx.GasLimit[0] = uint64(i)
		fmt.Printf("Testing gas %d\n", i)
		return !identicalRoots()
	})
	// And restore the gas again
	gst[testname].Tx.GasLimit[0] = uint64(gas)

	// Try removing accounts
	for target, acc := range gst[testname].Pre {
		delete(gst[testname].Pre, target)
		fmt.Printf("Testing dropping %x\n", target)
		if !identicalRoots() {
			continue
		}
		fmt.Printf("oops, broke it, restoring %x\n", target)
		gst[testname].Pre[target] = acc
	}
	// Try reducing code
	for target, acc := range gst[testname].Pre {
		if len(acc.Code) == 0 {
			continue
		}
		fmt.Printf("Target: %x\n", target)
		code := acc.Code
		m := codeMutator{current: code, lastGood: code}
		// Alright, we're in business
		fails := 0
		for {
			if exhausted := m.proceed(); exhausted {
				break
			}
			acc := gst[testname].Pre[target]
			acc.Code = m.current
			gst[testname].Pre[target] = acc
			if !identicalRoots() {
				fails = 0
				continue
			} else {
				fmt.Printf("oops, broke it, restoring\n")
				fails++
				m.undo()
				// restore it
				acc.Code = m.current
				gst[testname].Pre[target] = acc
				if fails > 5 {
					break
				}
			}
		}
	}
	fmt.Printf("Reducing slots...\n")
	// Try removing storage
	for target, acc := range gst[testname].Pre {
		for k, v := range acc.Storage {
			delete(gst[testname].Pre[target].Storage, k)
			fmt.Printf("Testing removing slot %x from %x\n", k, target)
			if !identicalRoots() {
				continue
			}
			gst[testname].Pre[target].Storage[k] = v
		}
	}

	return nil
}
