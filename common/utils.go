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
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"math/big"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/evms"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/utils"
	"github.com/urfave/cli/v2"
)

var (
	GethFlag = &cli.StringSliceFlag{
		Name:  "geth",
		Usage: "Location of go-ethereum 'evm' binary",
	}
	GethBatchFlag = &cli.StringSliceFlag{
		Name:  "gethbatch",
		Usage: "Location of go-ethereum 'evm' binary",
	}
	NethermindFlag = &cli.StringSliceFlag{
		Name:  "nethermind",
		Usage: "Location of nethermind 'nethtest' binary",
	}
	NethBatchFlag = &cli.StringSliceFlag{
		Name:  "nethbatch",
		Usage: "Location of nethermind 'nethtest' binary",
	}
	BesuFlag = &cli.StringSliceFlag{
		Name:  "besu",
		Usage: "Location of besu vm binary",
	}
	BesuBatchFlag = &cli.StringSliceFlag{
		Name:  "besubatch",
		Usage: "Location of besu vm binary",
	}
	ErigonFlag = &cli.StringSliceFlag{
		Name:  "erigon",
		Usage: "Location of erigon 'evm' binary",
	}
	ErigonBatchFlag = &cli.StringSliceFlag{
		Name:  "erigonbatch",
		Usage: "Location of erigon 'evm' binary",
	}
	NimbusFlag = &cli.StringSliceFlag{
		Name:  "nimbus",
		Usage: "Location of nimbus 'evmstate' binary",
	}
	ThreadFlag = &cli.IntFlag{
		Name:  "parallel",
		Usage: "Number of parallel executions to use.",
		Value: runtime.NumCPU(),
	}
	LocationFlag = &cli.StringFlag{
		Name:  "outdir",
		Usage: "Location to place artefacts",
		Value: "/tmp",
	}
	PrefixFlag = &cli.StringFlag{
		Name:  "prefix",
		Usage: "prefix of output files",
	}
	CountFlag = &cli.IntFlag{
		Name:  "count",
		Usage: "number of tests to generate",
		Value: 100,
	}
	TraceFlag = &cli.BoolFlag{
		Name: "trace",
		Usage: "if true, a trace will be generated along with the tests. \n" +
			"This is useful for debugging the usefulness of the tests",
	}
	SkipTraceFlag = &cli.BoolFlag{
		Name: "skiptrace",
		Usage: "If 'skiptrace' is set to true, then the evms will execute _without_ tracing, and only the final stateroot will be compared after execution.\n" +
			"This mode is faster, and can be used even if the clients-under-test has known errors in the trace-output, \n" +
			"but has a very high chance of missing cases which could be exploitable.",
	}
	VmFlags = []cli.Flag{
		GethFlag,
		GethBatchFlag,
		NethermindFlag,
		NethBatchFlag,
		BesuFlag,
		BesuBatchFlag,
		ErigonFlag,
		ErigonBatchFlag,
		NimbusFlag,
	}
	traceLengthSA = utils.NewSlidingAverage()
)

func initVMs(c *cli.Context) []evms.Evm {
	var (
		gethBins        = c.StringSlice(GethFlag.Name)
		gethBatchBins   = c.StringSlice(GethBatchFlag.Name)
		nethBins        = c.StringSlice(NethermindFlag.Name)
		nethBatchBins   = c.StringSlice(NethBatchFlag.Name)
		besuBins        = c.StringSlice(BesuFlag.Name)
		besuBatchBins   = c.StringSlice(BesuBatchFlag.Name)
		erigonBins      = c.StringSlice(ErigonFlag.Name)
		erigonBatchBins = c.StringSlice(ErigonBatchFlag.Name)
		nimBins         = c.StringSlice(NimbusFlag.Name)

		vms []evms.Evm
	)
	for i, bin := range gethBins {
		vms = append(vms, evms.NewGethEVM(bin, fmt.Sprintf("geth-%d", i)))
	}
	for i, bin := range gethBatchBins {
		vms = append(vms, evms.NewGethBatchVM(bin, fmt.Sprintf("gethbatch-%d", i)))
	}
	for i, bin := range nethBins {
		vms = append(vms, evms.NewNethermindVM(bin, fmt.Sprintf("nethermind-%d", i)))
	}
	for i, bin := range nethBatchBins {
		vms = append(vms, evms.NewNethermindBatchVM(bin, fmt.Sprintf("nethbatch-%d", i)))
	}
	for i, bin := range besuBins {
		vms = append(vms, evms.NewBesuVM(bin, fmt.Sprintf("besu-%d", i)))
	}
	for i, bin := range besuBatchBins {
		vms = append(vms, evms.NewBesuBatchVM(bin, fmt.Sprintf("besubatch-%d", i)))
	}
	for i, bin := range erigonBins {
		vms = append(vms, evms.NewErigonVM(bin, fmt.Sprintf("erigon-%d", i)))
	}
	for i, bin := range erigonBatchBins {
		vms = append(vms, evms.NewErigonBatchVM(bin, fmt.Sprintf("erigonbatch-%d", i)))
	}
	for i, bin := range nimBins {
		vms = append(vms, evms.NewNimbusEVM(bin, fmt.Sprintf("nimbus-%d", i)))
	}
	return vms

}

// RootsEqual executes the test on the given path on all vms, and returns true
// if they all report the same post stateroot.
func RootsEqual(path string, c *cli.Context) (bool, error) {
	var (
		vms   = initVMs(c)
		wg    sync.WaitGroup
		roots = make([]string, len(vms))
		errs  = make([]error, len(vms))
	)
	if len(vms) < 1 {
		return false, fmt.Errorf("No vms specified!")
	}
	wg.Add(len(vms))
	for i, vm := range vms {
		go func(index int, vm evms.Evm) {
			root, _, err := vm.GetStateRoot(path)
			roots[index] = root
			errs[index] = err
			vm.Close()
			wg.Done()
		}(i, vm)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return false, err
		}
	}
	for _, root := range roots[1:] {
		if root != roots[0] { // Consensus error
			return false, nil
		}
	}
	log.Info("Roots identical", "root", roots[0])
	return true, nil
}

// RunTests runs a test on all clients.
// Return values are :
// - true, nil: no consensus issue
// - true, err: test execution failed
// - false, err: a consensus issue found
// - false, nil: a consensus issue found
func RunSingleTest(path string, c *cli.Context) (bool, error) {
	var (
		vms     = initVMs(c)
		outputs []*os.File
	)
	if len(vms) < 1 {
		return true, fmt.Errorf("No vms specified!")
	}
	// Open/create outputs for writing
	for _, evm := range vms {
		out, err := os.OpenFile(fmt.Sprintf("./%v-output.jsonl", evm.Name()), os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			return true, fmt.Errorf("failed opening file %v", err)
		}
		outputs = append(outputs, out)
	}
	// Zero out the output files and reset offset
	for _, f := range outputs {
		_ = f.Truncate(0)
		_, _ = f.Seek(0, 0)
	}
	// Kick off the binaries
	var wg sync.WaitGroup
	wg.Add(len(vms))
	var commands = make([]string, len(vms))
	for i, vm := range vms {
		go func(evm evms.Evm, i int) {
			defer wg.Done()
			bufout := bufio.NewWriter(outputs[i])
			res, err := evm.RunStateTest(path, bufout, false)
			bufout.Flush()
			commands[i] = res.Cmd
			if err != nil {
				log.Error("Error running test", "err", err)
				return
			}
			log.Debug("Test done", "evm", evm.Name(), "time", res.ExecTime)
		}(vm, i)
	}
	wg.Wait()
	// Seek to beginning
	var readers []io.Reader
	for _, f := range outputs {
		_, _ = f.Seek(0, 0)
		readers = append(readers, f)
	}
	// Compare outputs
	if eq, _ := evms.CompareFiles(vms, readers); !eq {
		out := new(strings.Builder)
		fmt.Fprintf(out, "Consensus error\n")
		fmt.Fprintf(out, "Testcase: %v\n", path)
		for i, f := range outputs {
			fmt.Fprintf(out, "- %v: %v\n", vms[i].Name(), f.Name())
			fmt.Fprintf(out, "  - command: %v\n", commands[i])
		}
		fmt.Println(out)
		return false, fmt.Errorf("Consensus error")
	}
	for _, f := range outputs {
		f.Close()
	}
	log.Info("Execution done")
	return true, nil
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func TestSpeed(dir string, c *cli.Context) error {
	vms := initVMs(c)
	if len(vms) < 1 {
		return fmt.Errorf("No vms specified!")
	}
	if finfo, err := os.Stat(dir); err != nil {
		return err
	} else if !finfo.IsDir() {
		return fmt.Errorf("%v is not a directory", dir)
	}
	infoThreshold := time.Second
	warnThreshold := 5 * time.Second
	speedTest := func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, "json") {
			return nil
		}
		if err != nil {
			return err
		}
		// Run the binaries sequentially
		for _, evm := range vms {
			log.Debug("Starting test", "evm", evm.Name(), "file", path)
			res, err := evm.RunStateTest(path, noopWriter{}, true)
			if err != nil {
				log.Error("Error starting vm", "vm", evm.Name(), "err", err)
				return err
			}
			logger := log.Debug
			if res.ExecTime > warnThreshold {
				logger = log.Warn
			} else if res.ExecTime > infoThreshold {
				logger = log.Info
			}
			logger("Execution speed", "evm", evm.Name(), "file", path,
				"time", res.ExecTime, "cmd", res.Cmd)

		}
		return nil
	}
	return filepath.Walk(dir, speedTest)
}

type TestProviderFn func(index, threadId int) (string, error)

func testFnFromGenerator(fn GeneratorFn, name, location string) TestProviderFn {
	return func(index, threadId int) (string, error) {
		gstMaker := fn()
		testName := fmt.Sprintf("%08d-%v-%d", index, name, threadId)
		test := gstMaker.ToGeneralStateTest(testName)
		return storeTest(location, test, testName)
	}
}

type GeneratorFn func() *fuzzing.GstMaker

func GenerateAndExecute(c *cli.Context, generatorFn GeneratorFn, name string) error {
	var (
		location = c.String(LocationFlag.Name)
	)
	fn := testFnFromGenerator(generatorFn, name, location)
	return ExecuteFuzzer(c, fn, true)
}

func ExecuteFuzzer(c *cli.Context, providerFn TestProviderFn, cleanupFiles bool) error {

	var (
		vms        = initVMs(c)
		numThreads = c.Int(ThreadFlag.Name)
		skipTrace  = c.Bool(SkipTraceFlag.Name)
	)
	if len(vms) == 0 {
		return fmt.Errorf("need at least one vm to participate")
	}
	log.Info("Fuzzing started", "threads", numThreads)
	meta := &testMeta{
		errCh:       make(chan error, 10),  // Error channel
		testCh:      make(chan string, 10), // channel where we'll deliver tests
		removeCh:    make(chan string, 10), // channel for cleanup-taksks
		consensusCh: make(chan string, 10), // channel for signalling consensus errors
		slowCh:      make(chan string, 10), // channel for signalling slow tests
		vms:         vms,
	}
	// Routines to deliver tests
	meta.startTestFactories((numThreads+1)/2, providerFn)
	meta.startTestExecutors((numThreads+1)/2, skipTrace)
	// One goroutine to spit out some statistics
	ctx, cancel := context.WithCancel(context.Background())
	meta.wg.Add(1)
	go func() {
		defer meta.wg.Done()
		var (
			tStart    = time.Now()
			ticker    = time.NewTicker(8 * time.Second)
			testCount = uint64(0)
			ticks     = 0
		)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ticks++
				n := meta.numTests.Load()
				testsSinceLastUpdate := n - testCount
				testCount = n
				timeSpent := time.Since(tStart)
				// Update global counter
				globalCount := uint64(0)
				if content, err := os.ReadFile(".fuzzcounter"); err == nil {
					if count, err := strconv.Atoi((string(content))); err == nil {
						globalCount = uint64(count)
					}
				}
				globalCount += testsSinceLastUpdate
				if err := os.WriteFile(".fuzzcounter", []byte(fmt.Sprintf("%d", globalCount)), 0755); err != nil {
					log.Error("Error saving progress", "err", err)
				}
				log.Info("Executing",
					"tests", n,
					"time", common.PrettyDuration(timeSpent),
					"test/s", fmt.Sprintf("%.01f", float64(uint64(time.Second)*n)/float64(timeSpent)),
					"avg steps", fmt.Sprintf("%.01f", traceLengthSA.Avg()),
					"global", globalCount,
				)
				for _, vm := range vms {
					log.Info(fmt.Sprintf("Stats %v", vm.Name()), vm.Stats()...)
				}
				switch ticks {
				case 5:
					// Decrease stats-reporting after 40s
					ticker.Reset(time.Minute)
				case 65:
					// Decrease stats-reporting after one hour
					ticker.Reset(time.Hour)
				}
			case <-ctx.Done():
				return
			}
		}

	}()
	// One goroutine to clean up after ourselves
	meta.wg.Add(1)
	go func() {
		defer meta.wg.Done()
		for path := range meta.removeCh {
			if !cleanupFiles {
				continue
			}
			if err := os.Remove(path); err != nil {
				log.Error("Error deleting file", "file", path, "err", err)
			}
		}
	}()
	// One to handle slow tests
	meta.wg.Add(1)
	go func() {
		defer meta.wg.Done()
		for path := range meta.slowCh {
			newPath := filepath.Join(filepath.Dir(path),
				fmt.Sprintf("slowtest-%v", filepath.Base(path)))
			if err := Copy(path, newPath); err != nil {
				log.Error("Error copying file", "file", path, "err", err)
			}
		}
	}()

	// Cancel ability
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigs:
	case path := <-meta.consensusCh:
		log.Info("Possible consensus error", "file", path)
	case err := <-meta.errCh:
		log.Warn("Enocuntered error", "err", err)
	}
	log.Info("Waiting for processes to exit")
	meta.abort.Store(true)
	cancel()
	meta.wg.Wait()
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

type testMeta struct {
	abort       atomic.Bool
	errCh       chan error
	testCh      chan string
	removeCh    chan string
	consensusCh chan string
	slowCh      chan string
	wg          sync.WaitGroup
	vms         []evms.Evm
	numTests    atomic.Uint64
}

// startTestFactories creates a number of go-routines that write tests to disk, and delivers
// the paths on the testCh.
func (meta *testMeta) startTestFactories(numFactories int, providerFn TestProviderFn) {
	factories := int64(numFactories)
	meta.wg.Add(numFactories)
	factory := func(threadId int) {
		log.Info("Test factory thread started")
		defer func() {
			if f := atomic.AddInt64(&factories, -1); f == 0 {
				log.Info("Last test factory exiting\n")
				close(meta.testCh)
			}
			meta.wg.Done()
		}()
		for i := 0; !meta.abort.Load(); i++ {
			if fileName, err := providerFn(i, threadId); err != nil {
				log.Error("Error storing test", "err", err)
				meta.errCh <- err
				break
			} else {
				meta.testCh <- fileName
			}
		}
	}
	for i := 0; i < numFactories; i++ {
		go factory(i)
	}
}

func (meta *testMeta) startTestExecutors(numThreads int, skipTrace bool) {
	if skipTrace {
		meta.startNontracingTestExecutors(numThreads)
	} else {
		meta.startTracingTestExecutors(numThreads)
	}
}

func (meta *testMeta) startTracingTestExecutors(numThreads int) {
	var executors atomic.Int64

	execute := func(threadId int) {
		log.Info("Test executor routine started", "id", threadId)
		executors.Add(1)
		defer meta.wg.Done()
		defer func() {
			// clean-up tasks
			if f := executors.Add(-1); f == 0 {
				close(meta.removeCh)
				close(meta.slowCh)
				log.Info("Last executor exiting")
			}
		}()
		var (
			commands = make([]string, len(meta.vms))
			outputs  []*dualWriter
			vms      []evms.Evm
		)
		// Open/create outputs for writing
		for _, evm := range meta.vms {
			out, err := os.OpenFile(fmt.Sprintf("./%v-output-%d.jsonl", evm.Name(), threadId), os.O_CREATE|os.O_RDWR, 0755)
			if err != nil {
				log.Error("Failed opening file", "err", err)
				return
			}
			defer out.Close()
			outputs = append(outputs, newDualWriter(out))
		}
		// Start new thread-local instances of the VMs
		for _, vm := range meta.vms {
			vms = append(vms, vm.Instance(threadId))
		}
		// The fuzzing loop.
		for file := range meta.testCh {
			if meta.abort.Load() {
				// Continue to drain the testch
				continue
			}
			var slowTest atomic.Bool
			var vmWg sync.WaitGroup
			vmWg.Add(len(vms))
			// Kick off the binaries, which runs the test on all the vms in parallel
			for i, vm := range vms {
				go func(evm evms.Evm, i int) {
					defer vmWg.Done()
					outputs[i].Reset()
					res, err := evm.RunStateTest(file, outputs[i], false)
					commands[i] = res.Cmd
					if err != nil {
						log.Error("Error starting vm", "err", err, "command", res.Cmd)
						return
					}
					if res.Slow {
						// Flag test as slow
						log.Warn("Slow test found", "evm", evm.Name(), "time", res.ExecTime, "cmd", res.Cmd, "file", file)
						slowTest.Store(true)
					}
				}(vm, i)
			}
			vmWg.Wait()
			meta.numTests.Add(1)
			// All the tests are now executed, and we need to read and compare the outputs
			ok := true
			ref := outputs[0].Sum()
			for _, w := range outputs[1:] {
				have := w.Sum()
				ok = ok && bytes.Equal(have, ref)
			}
			if ok { // Output-hashes match, don't bother flushing to file and compare json.
				traceLengthSA.Add(outputs[0].nLines)

				if slowTest.Load() {
					meta.slowCh <- file // flag as slow
				} else {
					meta.removeCh <- file // flag for removal
				}
				continue
			}
			// Output-hashes do not match. At this point, we flush the full output
			// to file, and inspect line by line.
			var readers []io.Reader
			for _, w := range outputs {
				w.FlushAndSeek()
				readers = append(readers, w.f)
			}
			// Compare outputs
			if eq, _ := evms.CompareFiles(meta.vms, readers); !eq {
				meta.abort.Store(true)
				out := new(strings.Builder)
				fmt.Fprintf(out, "Consensus error\n")
				fmt.Fprintf(out, "Testcase: %v\n", file)
				for i, w := range outputs {
					fmt.Fprintf(out, "- %v: %v\n", meta.vms[i].Name(), w.f.Name())
					fmt.Fprintf(out, "  - command: %v\n", commands[i])
				}
				fmt.Println(out)
				meta.consensusCh <- file // flag as consensus-issue
			}
		}
	}
	numExecutors := numThreads / 2
	if numExecutors == 0 {
		numExecutors = 1
	}
	for i := 0; i < numExecutors; i++ {
		// Thread that executes the tests and compares the outputs
		meta.wg.Add(1)
		go execute(i)
	}
}

func (meta *testMeta) startNontracingTestExecutors(numThreads int) {
	executors := int64(0)
	execute := func(threadId int) {
		log.Info("Test executor routine started")
		defer meta.wg.Done()
		defer func() {
			// clean-up tasks
			if f := atomic.AddInt64(&executors, -1); f == 0 {
				close(meta.removeCh)
				close(meta.slowCh)
				log.Info("Last executor exiting")
			}
		}()
		atomic.AddInt64(&executors, 1)
		log.Info("Fuzzing thread started", "id", threadId)

		for file := range meta.testCh {
			if meta.abort.Load() {
				// Continue looping until testch is drained
				continue
			}
			// Zero out the output files and reset offset
			var slowTest uint32
			// Kick off the binaries, which runs the test on all the vms in parallel
			var vmWg sync.WaitGroup
			vmWg.Add(len(meta.vms))
			var roots = make([]string, len(meta.vms))
			var commands = make([]string, len(meta.vms))
			for i := range meta.vms {
				go func(i int) {
					defer vmWg.Done()
					var (
						t0             = time.Now()
						evm            = meta.vms[i]
						root, cmd, err = evm.GetStateRoot(file)
					)
					if err != nil {
						log.Error("Error starting vm", "vm", evm.Name(), "err", err)
						return
					}
					roots[i] = root
					commands[i] = cmd
					if execTime := time.Since(t0); execTime > 5*time.Second {
						log.Warn("Slow test found", "evm", evm.Name(), "time", execTime, "cmd", cmd, "file", file)
						// Flag test as slow
						atomic.StoreUint32(&slowTest, 1)
					}
				}(i)
			}
			vmWg.Wait()
			// All the tests are now executed, and we need to read and compare the roots
			meta.numTests.Add(1)
			// Compare roots
			consensusError := false
			for i, root := range roots[:len(roots)-1] {
				if root == roots[i+1] {
					continue
				}
				// Consensus error
				meta.abort.Store(true)
				consensusError = true
				break
			}
			if consensusError {
				out := new(strings.Builder)
				fmt.Fprintf(out, "Consensus error\n")
				for i, r := range roots {
					fmt.Fprintf(out, "  - %v stateroot: %v\n", meta.vms[i].Name(), r)
					fmt.Fprintf(out, "    - command: %v\n", commands[i])
				}
				fmt.Fprintf(out, "Testcase: %v\n", file)
				fmt.Println(out.String())
				meta.consensusCh <- file // flag as consensus-issue
			} else if slowTest != 0 {
				meta.slowCh <- file // flag as slow
			} else {
				meta.removeCh <- file // flag for removal
			}
		}
	}
	numExecutors := numThreads / 2
	if numExecutors == 0 {
		numExecutors = 1
	}
	for i := 0; i < numExecutors; i++ {
		// Thread that executes the tests and compares the outputs
		meta.wg.Add(1)
		go execute(i)
	}
}

// ConvertToStateTest is a utility to turn stuff into sharable state tests.
func ConvertToStateTest(name, fork string, alloc core.GenesisAlloc, gasLimit uint64, target common.Address) error {

	mkr := fuzzing.BasicStateTest(fork)
	// convert the genesisAlloc
	var fuzzGenesisAlloc = make(fuzzing.GenesisAlloc)
	for k, v := range alloc {
		fuzzAcc := fuzzing.GenesisAccount{
			Code:       v.Code,
			Storage:    v.Storage,
			Balance:    v.Balance,
			Nonce:      v.Nonce,
			PrivateKey: v.PrivateKey,
		}
		if fuzzAcc.Balance == nil {
			fuzzAcc.Balance = new(big.Int)
		}
		if fuzzAcc.Storage == nil {
			fuzzAcc.Storage = make(map[common.Hash]common.Hash)
		}
		fuzzGenesisAlloc[k] = fuzzAcc
	}
	// Also add the sender
	var sender = common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	if _, ok := fuzzGenesisAlloc[sender]; !ok {
		fuzzGenesisAlloc[sender] = fuzzing.GenesisAccount{
			Balance: big.NewInt(1000000000000000000), // 1 eth
			Nonce:   0,
			Storage: make(map[common.Hash]common.Hash),
		}
	}

	tx := &fuzzing.StTransaction{
		GasLimit:   []uint64{gasLimit},
		Nonce:      0,
		Value:      []string{"0x0"},
		Data:       []string{""},
		GasPrice:   big.NewInt(0x10),
		PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		To:         target.Hex(),
	}
	mkr.SetTx(tx)
	mkr.SetPre(&fuzzGenesisAlloc)
	if err := mkr.Fill(nil); err != nil {
		return err
	}
	gst := mkr.ToGeneralStateTest(name)
	dat, _ := json.MarshalIndent(gst, "", " ")
	fname := fmt.Sprintf("%v.json", name)
	if err := os.WriteFile(fname, dat, 0777); err != nil {
		return err
	}
	fmt.Printf("Wrote file %v\n", fname)
	return nil
}

// dualWriter is a io.Writer which writes to a hash, and also to a file, via a
// 5-MB buffer. Ideally, the file-write never happens.
// If the data is smaller than the internal buffer, and the hashes match up, the
// writer can be Reset, and thus the file-write aborted.
// This writer can thus be used to avoid both writing to (via buffer + discard)
// and reading from (via hasher-check) disk.
type dualWriter struct {
	f      *os.File
	out    *bufio.Writer
	sum    hash.Hash
	nLines int
}

func newDualWriter(f *os.File) *dualWriter {
	out := bufio.NewWriterSize(f, 5*1024*1024)
	w := &dualWriter{
		f:   f,
		out: out,
		sum: md5.New(),
	}
	w.Reset()
	return w
}

func (w *dualWriter) Write(data []byte) (int, error) {
	w.nLines += bytes.Count(data, []byte{'\n'})
	_, _ = w.out.Write(data)
	return w.sum.Write(data)
}

func (w *dualWriter) Sum() []byte {
	return w.sum.Sum(nil)
}

func (w *dualWriter) FlushAndSeek() {
	w.out.Flush()
	_ = w.f.Sync()
	_, _ = w.f.Seek(0, 0)
}

func (w *dualWriter) Reset() {
	_ = w.f.Truncate(0)
	_, _ = w.f.Seek(0, 0)
	w.out.Reset(w.f)
	w.sum.Reset()
	w.nLines = 0
}

// Close closes the underlying file handle.
func (w *dualWriter) Close() {
	w.f.Close()
}
