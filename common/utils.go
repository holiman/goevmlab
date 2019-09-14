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

package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/holiman/goevmlab/evms"
	"github.com/holiman/goevmlab/fuzzing"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	GethFlag = cli.StringFlag{
		Name:     "geth",
		Usage:    "Location of go-ethereum 'evm' binary",
		Required: true,
	}
	ParityFlag = cli.StringFlag{
		Name:     "parity",
		Usage:    "Location of go-ethereum 'parity-vm' binary",
		Required: true,
	}
	ThreadFlag = cli.IntFlag{
		Name:  "paralell",
		Usage: "Number of paralell executions to use.",
		Value: runtime.NumCPU(),
	}
	LocationFlag = cli.StringFlag{
		Name:  "outdir",
		Usage: "Location to place artefacts",
		Value: "/tmp",
	}
	PrefixFlag = cli.StringFlag{
		Name:  "prefix",
		Usage: "prefix of output files",
	}
	CountFlag = cli.IntFlag{
		Name:  "count",
		Usage: "number of tests to generate",
	}
)

type GeneratorFn func() *fuzzing.GstMaker

func ExecuteFuzzer(c *cli.Context, generatorFn GeneratorFn, name string) error {

	var (
		gethBin    = c.GlobalString(GethFlag.Name)
		parityBin  = c.GlobalString(ParityFlag.Name)
		numThreads = c.GlobalInt(ThreadFlag.Name)
		location   = c.GlobalString(LocationFlag.Name)
		numTests   uint64
	)
	fmt.Printf("numThreads: %d\n", numThreads)
	var wg sync.WaitGroup
	// The channel where we'll deliver tests
	testCh := make(chan string, 10)
	// Cancel ability
	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	// Thread that creates tests, spits out filenames

	for i := 0; i < numThreads/2; i++ {
		go func(threadId int) {
			defer wg.Done()
			for i := 0; ; i++ {
				gstMaker := generatorFn()
				testName := fmt.Sprintf("%d-%v-%d", threadId, name, i)
				test := gstMaker.ToGeneralStateTest(testName)
				fileName, err := storeTest(location, test, testName)
				if err != nil {
					fmt.Printf("Error: %v", err)
					break
				}
				select {
				case testCh <- fileName:
				case <-ctx.Done():
					break
				}
			}
		}(i)
	}
	for i := 0; i < numThreads/2; i++ {
		// Thread that executes the tests and compares the outputs
		wg.Add(1)
		go func() {
			defer wg.Done()
			geth := evms.NewGethEVM(gethBin)
			parity := evms.NewParityVM(parityBin)
			fmt.Printf("Fuzzing started \n")
			for file := range testCh {
				if err := compareOutputs(geth, parity, file); err != nil {
					fmt.Printf("Error occurred in executor: %v", err)
					break
				}
				atomic.AddUint64(&numTests, 1)
			}
		}()
	}
	// One goroutine to spit out some statistics
	wg.Add(1)
	go func() {
		defer wg.Done()
		tStart := time.Now()
		ticker := time.NewTicker(5 * time.Second)
		testCount := uint64(0)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				n := atomic.LoadUint64(&numTests)
				testsSinceLastUpdate := n - testCount
				testCount = n
				timeSpent := time.Since(tStart)
				execPerSecond := float64(uint64(time.Second)*n) / float64(timeSpent)
				fmt.Printf("%d tests executed, in %v (%.02f tests/s)\n", n, timeSpent, execPerSecond)
				// Update global counter
				globalCount := uint64(0)
				if content, err := ioutil.ReadFile(".fuzzcounter"); err == nil {
					if count, err := strconv.Atoi((string(content))); err == nil {
						globalCount = uint64(count)
					}
				}
				globalCount += testsSinceLastUpdate

				ioutil.WriteFile(".fuzzcounter", []byte(fmt.Sprintf("%d", globalCount)), 0755)
			case <-ctx.Done():
				break
			}
		}

	}()

	<-sigs
	fmt.Println("Exiting...")
	cancel()
	return nil
}

func compareOutputs(a, b evms.Evm, testfile string) error {
	comparer := evms.Comparer{}
	chA, err := a.StartStateTest(testfile)
	if err != nil {
		return fmt.Errorf("failed [1] starting testfile %v: %v", testfile, err)
	}
	chB, err := b.StartStateTest(testfile)
	if err != nil {
		return fmt.Errorf("failed [2] starting testfile %v: %v", testfile, err)
	}
	outCh := comparer.CompareVms(chA, chB)
	for outp := range outCh {
		fmt.Printf("Output: %v\n", outp)
		err = errors.New("consensus error")
	}
	fmt.Printf("file %v: stats: %v\n", testfile, comparer.Stats())
	return err
}

// storeTest saves a testcase to disk
func storeTest(location string, test *fuzzing.GeneralStateTest, testName string) (string, error) {

	fileName := fmt.Sprintf("%v.json", testName)
	fullPath := path.Join(location, fileName)

	f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		return "", err
	}
	defer f.Close()
	// Write to file
	encoder := json.NewEncoder(f)
	if err = encoder.Encode(test); err != nil {
		return fullPath, err
	}
	return fullPath, nil
}
