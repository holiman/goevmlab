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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/evms"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/utils"
	"github.com/urfave/cli/v2"
	"net/http"
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
	EelsFlag = &cli.StringSliceFlag{
		Name:  "eels",
		Usage: "Location of 'ethereum-spec-evm' binary",
	}
	EelsBatchFlag = &cli.StringSliceFlag{
		Name:  "eelsbatch",
		Usage: "Location of 'ethereum-spec-evm' binary",
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
	EvmoneFlag = &cli.StringSliceFlag{
		Name:  "evmone",
		Usage: "Location of evmone 'evmone' binary",
	}
	RethFlag = &cli.StringSliceFlag{
		Name:  "revme",
		Usage: "Location of reth 'revme' binary",
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
	NotifyFlag = &cli.StringFlag{
		Name:  "ntfy",
		Usage: "Topic to sent 'https://ntfy.sh/'-ping on exit (e.g. due to consensus issue)",
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
	VerbosityFlag = &cli.IntFlag{
		Name:  "verbosity",
		Usage: "sets the verbosity level (-4: DEBUG, 0: INFO, 4: WARN, 8: ERROR)",
		Value: 0,
	}

	VmFlags = []cli.Flag{
		GethFlag,
		GethBatchFlag,
		EelsFlag,
		EelsBatchFlag,
		NethermindFlag,
		NethBatchFlag,
		BesuFlag,
		BesuBatchFlag,
		ErigonFlag,
		ErigonBatchFlag,
		NimbusFlag,
		EvmoneFlag,
		RethFlag,
	}
	traceLengthSA = utils.NewSlidingAverage()
)

func initVMs(c *cli.Context) []evms.Evm {
	var (
		gethBins        = c.StringSlice(GethFlag.Name)
		gethBatchBins   = c.StringSlice(GethBatchFlag.Name)
		eelsBins        = c.StringSlice(EelsFlag.Name)
		eelsBatchBins   = c.StringSlice(EelsBatchFlag.Name)
		nethBins        = c.StringSlice(NethermindFlag.Name)
		nethBatchBins   = c.StringSlice(NethBatchFlag.Name)
		besuBins        = c.StringSlice(BesuFlag.Name)
		besuBatchBins   = c.StringSlice(BesuBatchFlag.Name)
		erigonBins      = c.StringSlice(ErigonFlag.Name)
		erigonBatchBins = c.StringSlice(ErigonBatchFlag.Name)
		nimBins         = c.StringSlice(NimbusFlag.Name)
		evmoneBins      = c.StringSlice(EvmoneFlag.Name)
		revmBins        = c.StringSlice(RethFlag.Name)

		vms []evms.Evm
	)
	for i, bin := range gethBins {
		vms = append(vms, evms.NewGethEVM(bin, fmt.Sprintf("geth-%d", i)))
	}
	for i, bin := range gethBatchBins {
		vms = append(vms, evms.NewGethBatchVM(bin, fmt.Sprintf("gethbatch-%d", i)))
	}
	for i, bin := range eelsBins {
		vms = append(vms, evms.NewEelsEVM(bin, fmt.Sprintf("eels-%d", i)))
	}
	for i, bin := range eelsBatchBins {
		vms = append(vms, evms.NewEelsBatchVM(bin, fmt.Sprintf("eelsbatch-%d", i)))
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
	for i, bin := range evmoneBins {
		vms = append(vms, evms.NewEvmoneVM(bin, fmt.Sprintf("%d", i)))
	}
	for i, bin := range revmBins {
		vms = append(vms, evms.NewRethVM(bin, fmt.Sprintf("%d", i)))
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
			log.Info("Root found", "stateroot", root, "vm", vm.Name(), "err", err)
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
		outdir  = c.String(LocationFlag.Name)
	)
	if len(vms) < 1 {
		return true, fmt.Errorf("No vms specified!")
	}
	// Open/create outputs for writing
	for _, evm := range vms {
		out, err := os.OpenFile(fmt.Sprintf("%v/%v-output.jsonl", outdir, evm.Name()), os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0755)
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
			if res != nil {
				commands[i] = res.Cmd
			}
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
	if eq, _, diff := evms.CompareFiles(vms, readers); !eq {
		fmt.Print(diff)
		out := new(strings.Builder)
		fmt.Fprintf(out, "Consensus error\n")
		fmt.Fprintf(out, "Testcase: %v\n", path)
		for i, f := range outputs {
			fmt.Fprintf(out, "- %v: %v\n", vms[i].Name(), f.Name())
			fmt.Fprintf(out, "  - command: %v\n", commands[i])
		}
		fmt.Fprintf(out, "\nTo view the difference with tracediff:\n\ttracediff %v %v\n", outputs[0].Name(), outputs[0].Name())
		fmt.Println(out)
		return false, fmt.Errorf("Consensus error")
	}

	for _, f := range outputs {
		f.Close()
	}
	log.Info("Execution done")
	return true, nil
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
			res, err := evm.RunStateTest(path, io.Discard, true)
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
	fn := testFnFromGenerator(generatorFn, name, c.String(LocationFlag.Name))
	return ExecuteFuzzer(c, false, fn, true)
}

func ExecuteFuzzer(c *cli.Context, allClients bool, providerFn TestProviderFn, cleanupFiles bool) error {
	var (
		vms        = initVMs(c)
		numThreads = c.Int(ThreadFlag.Name)
		skipTrace  = c.Bool(SkipTraceFlag.Name)
		numClients = 2
	)
	if allClients {
		numClients = len(vms)
	}
	if len(vms) == 0 {
		return fmt.Errorf("need at least one vm to participate")
	}
	log.Info("Fuzzing started", "threads", numThreads)
	meta := &testMeta{
		testCh:              make(chan string, 4), // channel where we'll deliver tests
		consensusCh:         make(chan string, 4), // channel for signalling consensus errors
		vms:                 vms,
		deleteFilesWhenDone: cleanupFiles,
		outdir:              c.String(LocationFlag.Name),
		notifyTopic:         c.String(NotifyFlag.Name),
	}
	// Routines to deliver tests
	meta.startTestFactories((numThreads+1)/2, providerFn)
	meta.wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		meta.fuzzingLoop(skipTrace, numClients)
		cancel()
	}()
	// One goroutine to spit out some statistics
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
	// Cancel ability
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigs:
	case <-ctx.Done():
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
	testCh      chan string
	consensusCh chan string
	wg          sync.WaitGroup
	vms         []evms.Evm
	numTests    atomic.Uint64
	outdir      string
	notifyTopic string

	deleteFilesWhenDone bool
}

// startTestFactories creates a number of go-routines that write tests to disk, and delivers
// the paths on the testCh.
func (meta *testMeta) startTestFactories(numFactories int, providerFn TestProviderFn) {
	var factories atomic.Int64
	factories.Add(int64(numFactories))
	meta.wg.Add(numFactories)
	factory := func(threadId int) {
		log.Info("Test factory thread started")
		defer func() {
			log.Info("Factory exiting")
			if f := factories.Add(-1); f == 0 {
				log.Info("Last test factory exiting\n")
				close(meta.testCh)
			}
			meta.wg.Done()
		}()
		for i := 0; !meta.abort.Load(); i++ {
			fileName, err := providerFn(i, threadId)
			if err == io.EOF {
				log.Info("Test provider done, exiting")
				break
			}
			if err != nil {
				log.Error("Error generating test, exiting", "err", err)
				break
			}
			log.Trace("Shipping a test", "file", fileName)
			meta.testCh <- fileName
		}
	}
	for i := 0; i < numFactories; i++ {
		go factory(i)
	}
}

type task struct {
	// pre-execution fields:
	file      string // file is the input statetest
	testIdx   int    // testIdx is a global index of the test
	vmIdx     int    // vmIdx is a global index of the vm
	skipTrace bool   // skipTrace: if true, ignore output and just exec as fast as possible

	// post-execution fields:
	execSpeed time.Duration
	slow      bool   // set by the executor if the test is deemed slow.
	result    []byte // result is the md5 hash of the execution output
	nLines    int    // number of lines of output
	command   string // command used to execute the test
	err       error  // if error occurred
}

type lineCountingHasher struct {
	h     hash.Hash
	lines int
}

func newLineCountingHasher() *lineCountingHasher {
	return &lineCountingHasher{md5.New(), 0}
}

func (l *lineCountingHasher) Write(p []byte) (n int, err error) {
	var count int
	for _, c := range p {
		if c == '\n' {
			count++
		}
	}
	l.lines += count
	return l.h.Write(p)
}

func (l *lineCountingHasher) Reset() {
	l.h.Reset()
	l.lines = 0
}

func (meta *testMeta) vmLoop(evm evms.Evm, taskCh, resultCh chan *task) {
	defer meta.wg.Done()
	var hasher = newLineCountingHasher()
	for t := range taskCh {
		hasher.Reset()
		res, err := evm.RunStateTest(t.file, hasher, t.skipTrace)
		if err != nil {
			log.Error("Error starting vm", "err", err, "evm", evm.Name())
			t.err = fmt.Errorf("error starting vm %v: %w", evm.Name(), err)
			// Send back
			resultCh <- t
			continue
		}
		if res.Slow {
			log.Warn("Slow test found", "evm", evm.Name(), "time", res.ExecTime, "cmd", res.Cmd, "file", t.file)
		} else {
			log.Debug("Test executed", "evm", evm.Name(), "time", res.ExecTime, "cmd", res.Cmd, "file", t.file)
		}
		t.slow = res.Slow
		t.result = hasher.h.Sum(nil)
		t.nLines = hasher.lines
		t.command = res.Cmd
		t.execSpeed = res.ExecTime
		// Send back
		resultCh <- t
	}
	log.Debug("vmloop exiting")
}

type cleanTask struct {
	slow   string // path to a file considered 'slow'
	remove string // path to a file to be removed
}

func (meta *testMeta) cleanupLoop(cleanCh chan *cleanTask) {
	defer meta.wg.Done()
	for task := range cleanCh {
		if path := task.slow; path != "" {
			newPath := filepath.Join(filepath.Dir(path), fmt.Sprintf("slowtest-%v", filepath.Base(path)))
			if err := Copy(path, newPath); err != nil {
				log.Error("Error copying file", "file", path, "err", err)
			}
		}
		if path := task.remove; path != "" && meta.deleteFilesWhenDone {
			if err := os.Remove(path); err != nil {
				log.Error("Error deleting file", "file", path, "err", err)
			}
		}
	}
	log.Debug("CleanupLoop exiting")
}

func (meta *testMeta) handleConsensusFlaw(testfile string) {
	output := new(strings.Builder)
	fmt.Fprintf(output, "Consensus error\n")
	fmt.Fprintf(output, "Testcase: %v\n", testfile)
	var readers []io.Reader
	var diffargs []string
	for _, evm := range meta.vms {
		filename := fmt.Sprintf("%v/%v-output.jsonl", meta.outdir, evm.Name())
		out, err := os.Create(filename)
		if err != nil {
			log.Error("Failed opening file", "err", err)
			panic(err)
		}
		res, err := evm.RunStateTest(testfile, out, false)
		if err != nil {
			log.Error("Failed running vm", "err", err)
			panic(err)
		}
		fmt.Fprintf(output, "- %v: %v\n", evm.Name(), filename)
		fmt.Fprintf(output, "  - command: %v\n", res.Cmd)
		diffargs = append(diffargs, filename)
		_ = out.Sync()
		_, _ = out.Seek(0, 0)
		readers = append(readers, out)
	}
	fmt.Fprintf(output, "\nTo view the difference with tracediff:\n\ttracediff %v %v\n", diffargs[0], diffargs[1])

	// Compare outputs (and show diff)
	_, _, diff := evms.CompareFiles(meta.vms, readers)
	fmt.Fprint(output, diff)
	fmt.Println(output.String())
	if meta.notifyTopic != "" {
		if _, err := http.Post(fmt.Sprintf("https://ntfy.sh/%v", meta.notifyTopic), "text/plain",
			strings.NewReader(output.String())); err != nil {
			fmt.Printf("Failed to post notification: %v\n", err)
		}
	}

	for _, f := range readers {
		f.(*os.File).Close()
	}
}

func (meta *testMeta) fuzzingLoop(skipTrace bool, clientCount int) {
	var (
		ready        []int
		testIndex    = 0
		taskChannels []chan *task
		resultCh     = make(chan *task)
		cleanCh      = make(chan *cleanTask)
	)
	defer meta.wg.Done()
	defer close(cleanCh)
	// Start n vmLoops.
	for i, vm := range meta.vms {
		var taskCh = make(chan *task)
		taskChannels = append(taskChannels, taskCh)
		meta.wg.Add(1)
		go meta.vmLoop(vm, taskCh, resultCh)
		ready = append(ready, i)
	}

	meta.wg.Add(1)
	go meta.cleanupLoop(cleanCh)

	type execResult struct {
		hash          []byte // hash of the output
		slow          bool   // whether it was considered slow
		consensusFlaw bool   // whether it triggered a consensus flaw
		waiting       int    // the number of clients we're waiting the results from
	}
	var executing = make(map[string]*execResult)
	readResults := func(count int) {
		for i := 0; i < count; i++ {
			t := <-resultCh                // result delivery
			ready = append(ready, t.vmIdx) // add client to ready-set
			if t.err != nil {
				log.Error("Error", "err", t.err)
				meta.abort.Store(true)
				continue
			}
			execRs := executing[t.file]
			execRs.waiting--

			if t.slow {
				execRs.slow = true
			}
			// check results
			if execRs.hash == nil { // first
				execRs.hash = t.result
			}
			if !bytes.Equal(execRs.hash, t.result) {
				log.Info("Consensus flaw", "file", t.file)
				execRs.consensusFlaw = true
			}
			if execRs.waiting > 0 {
				continue
			}
			traceLengthSA.Add(t.nLines)
			// No more results in the pipeline
			delete(executing, t.file)
			meta.numTests.Add(1)
			switch {
			case execRs.consensusFlaw:
				meta.consensusCh <- t.file
				meta.abort.Store(true)
			case execRs.slow:
				cleanCh <- &cleanTask{slow: t.file}
			default:
				cleanCh <- &cleanTask{remove: t.file}
			}
		}
	}
	for testfile := range meta.testCh {
		testIndex++
		// First, make sure we have N clients to execute the test on
		if clientsNeeded := clientCount - len(ready); clientsNeeded > 0 {
			readResults(clientsNeeded)
		}
		if meta.abort.Load() {
			log.Info("Shortcutting through abort")
			continue
		}
		// Dispatch the testfile to the ready clients
		log.Trace("Dispatching test to clients", "count", clientCount)
		executing[testfile] = &execResult{waiting: clientCount}
		for i := 0; i < clientCount; i++ {
			id := ready[0]
			taskChannels[id] <- &task{
				file:      testfile,
				testIdx:   testIndex,
				vmIdx:     id,
				skipTrace: skipTrace,
			}
			ready = ready[1:]
		}
	}
	// Close all task channels
	for _, taskCh := range taskChannels {
		close(taskCh)
	}
	// drain resultchanne;
	for len(ready) < len(meta.vms) {
		readResults(len(meta.vms) - len(ready))
	}
	log.Debug("Fuzzing loop exiting")
	// We might have a consensus issue to investigate
	select {
	case testfile := <-meta.consensusCh:
		meta.handleConsensusFlaw(testfile)
	default:
	}
}

// ConvertToStateTest is a utility to turn stuff into sharable state tests.
func ConvertToStateTest(name, fork string, alloc types.GenesisAlloc, gasLimit uint64, target common.Address) error {

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
		Sender:     sender,
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
