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
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"gopkg.in/urfave/cli.v1"

	"github.com/holiman/goevmlab/evms"
	"github.com/holiman/goevmlab/fuzzing"
)

var (
	GethFlag = cli.StringFlag{
		Name:  "geth",
		Usage: "Location of go-ethereum 'evm' binary",
		//Required: true,
	}
	ParityFlag = cli.StringFlag{
		Name:  "parity",
		Usage: "Location of go-ethereum 'parity-vm' binary",
		//Required: true,
	}
	NethermindFlag = cli.StringFlag{
		Name:  "nethermind",
		Usage: "Location of nethermind 'nethtest' binary",
		//Required: true,
	}
	AlethFlag = cli.StringFlag{
		Name:  "testeth",
		Usage: "Location of aleth 'testeth' binary",
		//Required: true,
	}
	ThreadFlag = cli.IntFlag{
		Name:  "parallel",
		Usage: "Number of parallel executions to use.",
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

func initVMs(c *cli.Context) []evms.Evm {
	var (
		gethBin   = c.GlobalString(GethFlag.Name)
		parityBin = c.GlobalString(ParityFlag.Name)
		nethBin   = c.GlobalString(NethermindFlag.Name)
		alethBin  = c.GlobalString(AlethFlag.Name)
	)
	vms := []evms.Evm{evms.NewGethEVM(gethBin)}
	if parityBin != "" {
		vms = append(vms, evms.NewParityVM(parityBin))
	}
	if nethBin != "" {
		vms = append(vms, evms.NewNethermindVM(nethBin))
	}
	if alethBin != "" {
		vms = append(vms, evms.NewAlethVM(alethBin))
	}
	return vms

}

func RunOneTest(path string, c *cli.Context) error {
	var (
		vms     = initVMs(c)
		outputs []*os.File
		readers []io.Reader
	)
	if len(vms) < 1 {
		return fmt.Errorf("No vms specified!")
	}
	// Open/create outputs for writing
	for _, evm := range vms {
		out, err := os.OpenFile(fmt.Sprintf("./%v-output.jsonl", evm.Name()), os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			return fmt.Errorf("failed opening file %v", err)
		}
		outputs = append(outputs, out)
	}
	// Kick off the binaries
	var wg sync.WaitGroup
	wg.Add(len(vms))
	for i, vm := range vms {
		go func(evm evms.Evm, out io.Writer) {
			defer wg.Done()
			t0 := time.Now()
			if _, err := evm.RunStateTest(path, out, false); err != nil {
				fmt.Fprintf(os.Stderr, "error running test: %v\n", err)
				return
			}
			execTime := time.Since(t0)
			fmt.Printf("%10v done in %v\n", evm.Name(), execTime)
		}(vm, outputs[i])
	}
	wg.Wait()
	// Seek to beginning
	for _, f := range outputs {
		_, _ = f.Seek(0, 0)
		readers = append(readers, f)
	}
	// Compare outputs
	if eq := evms.CompareFiles(vms, readers); !eq {
		fmt.Printf("output files: %v, %v, %v\n", outputs[0].Name(), outputs[1].Name(), outputs[2].Name())
		return fmt.Errorf("Consensus error")
	}
	fmt.Printf("all agree!")
	return nil
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestSpeed(path string, c *cli.Context) (bool, error) {
	var (
		vms = initVMs(c)
	)
	if len(vms) < 1 {
		return false, fmt.Errorf("No vms specified!")
	}
	// Kick off the binaries
	var wg sync.WaitGroup
	var slowTest uint32
	wg.Add(len(vms))
	for _, vm := range vms {
		go func(evm evms.Evm) {
			defer wg.Done()
			t0 := time.Now()
			if _, err := evm.RunStateTest(path, noopWriter{}, true); err != nil {
				fmt.Fprintf(os.Stderr, "error running test: %v\n", err)
				return
			}
			execTime := time.Since(t0)
			if execTime > 2*time.Second {
				fmt.Printf("%v: %10v done in %v\n", path, evm.Name(), execTime)
				atomic.StoreUint32(&slowTest, 1)
			}
		}(vm)
	}
	wg.Wait()
	return slowTest != 0, nil
}

type GeneratorFn func() *fuzzing.GstMaker

func ExecuteFuzzer(c *cli.Context, generatorFn GeneratorFn, name string) error {

	var (
		vms        = initVMs(c)
		numThreads = c.GlobalInt(ThreadFlag.Name)
		location   = c.GlobalString(LocationFlag.Name)
		numTests   uint64
	)

	if len(vms) == 0 {
		return fmt.Errorf("need at least onee vm to participate")
	}

	fmt.Printf("numThreads: %d\n", numThreads)
	var wg sync.WaitGroup
	// The channel where we'll deliver tests
	testCh := make(chan string, 10)
	// The channel for cleanup-taksks
	removeCh := make(chan string, 10)
	// channel for signalling consensus errors
	consensusCh := make(chan string, 10)
	// channel for signalling slow tests
	slowCh := make(chan string, 10)

	// Cancel ability
	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	abort := int64(0)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Thread that creates tests, spits out filenames
	numFactories := numThreads / 2
	factories := int64(numFactories)
	for i := 0; i < numFactories; i++ {
		wg.Add(1)
		go func(threadId int) {
			defer wg.Done()
			defer func() {
				if f := atomic.AddInt64(&factories, -1); f == 0 {
					fmt.Printf("closing testCh\n")
					close(testCh)
				}
			}()
			for i := 0; atomic.LoadInt64(&abort) == 0; i++ {
				gstMaker := generatorFn()
				testName := fmt.Sprintf("%08d-%v-%d", i, name, threadId)
				test := gstMaker.ToGeneralStateTest(testName)
				fileName, err := storeTest(location, test, testName)
				if err != nil {
					fmt.Printf("Error: %v", err)
					break
				}
				testCh <- fileName
			}
		}(i)
	}
	executors := int64(0)
	for i := 0; i < numThreads/2; i++ {
		// Thread that executes the tests and compares the outputs
		wg.Add(1)
		go func(threadId int) {
			defer wg.Done()
			atomic.AddInt64(&executors, 1)
			var outputs []*os.File
			defer func() {
				if f := atomic.AddInt64(&executors, -1); f == 0 {
					close(removeCh)
					close(slowCh)
				}
			}()
			defer func() {
				for _, f := range outputs {
					f.Close()
				}
			}()
			fmt.Printf("Fuzzing started \n")
			// Open/create outputs for writing
			for _, evm := range vms {
				out, err := os.OpenFile(fmt.Sprintf("./%v-output-%d.jsonl", evm.Name(), threadId), os.O_CREATE|os.O_RDWR, 0755)
				if err != nil {
					fmt.Printf("failed opening file %v", err)
					return
				}
				outputs = append(outputs, out)
			}
			for file := range testCh {
				if atomic.LoadInt64(&abort) == 1 {
					// Continue to drain the testch
					continue
				}
				// Zero out the output files and reset offset
				for _, f := range outputs {
					_ = f.Truncate(0)
					_, _ = f.Seek(0, 0)
				}
				var slowTest uint32
				// Kick off the binaries
				var wg sync.WaitGroup
				wg.Add(len(vms))
				for i, vm := range vms {
					go func(evm evms.Evm, out io.Writer) {
						t0 := time.Now()
						cmd, _ := evm.RunStateTest(file, out, false)
						execTime := time.Since(t0)
						if execTime > 20*time.Second {
							fmt.Printf("%10v done in %v (slow!). Cmd: %v\n", evm.Name(), execTime, cmd)
							// Flag test as slow
							atomic.StoreUint32(&slowTest, 1)
						}
						wg.Done()
					}(vm, outputs[i])
				}
				wg.Wait()
				var readers []io.Reader
				// Flush file and set reset offset
				for _, f := range outputs {
					_ = f.Sync()
					_, _ = f.Seek(0, 0)
					readers = append(readers, f)
				}
				atomic.AddUint64(&numTests, 1)
				// Compare outputs
				eq := evms.CompareFiles(vms, readers)
				if !eq {
					atomic.StoreInt64(&abort, 1)
					s := "output files: "
					for _, f := range outputs {
						s = fmt.Sprintf("%v %v", s, f.Name())
					}
					fmt.Println(s)
					consensusCh <- file
				} else if slowTest != 0 {
					slowCh <- file
				} else {
					removeCh <- file

				}
			}
		}(i)
	}
	// One goroutine to spit out some statistics
	wg.Add(1)
	go func() {
		defer wg.Done()
		tStart := time.Now()
		ticker := time.NewTicker(60 * time.Second)
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

				if err := ioutil.WriteFile(".fuzzcounter", []byte(fmt.Sprintf("%d", globalCount)), 0755); err != nil {
					fmt.Fprintf(os.Stderr, "error saving progress: %v\n", err)
				}
			case <-ctx.Done():
				return
			}
		}

	}()
	// One goroutine to clean up after ourselves
	wg.Add(1)
	go func() {
		defer wg.Done()
		for path := range removeCh {
			if err := os.Remove(path); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting file %v, : %v\n", path, err)
			}
		}
	}()
	// One to handle slow tests
	wg.Add(1)
	go func() {
		defer wg.Done()
		for path := range slowCh {
			newPath := filepath.Join(filepath.Dir(path),
				fmt.Sprintf("slowtest-%v", filepath.Base(path)))
			if err := Copy(path, newPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error copying file file %v, : %v\n", path, err)
			}
		}
	}()
	select {
	case <-sigs:
	case path := <-consensusCh:
		fmt.Printf("Possible consensus error!\nFile: %v\n", path)
	}
	fmt.Printf("waiting for procs to exit\n")
	atomic.StoreInt64(&abort, 1)
	cancel()
	wg.Wait()
	return nil
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

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
