package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/holiman/goevmlab/evms"
	"github.com/holiman/goevmlab/fuzzing"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Martin Holst Swende"
	app.Usage = "Generator for blake (state-)tests"
	return app
}

var (
	app      = initApp()
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
		Usage: "Number of paralell executions to use. (default = numcpu)",
		Value: runtime.NumCPU(),
	}
)

func init() {
	app.Flags = []cli.Flag{
		GethFlag,
		ParityFlag,
		ThreadFlag,
	}
	app.Action = testBlake
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func testBlake(c *cli.Context) error {

	var (
		gethBin    = c.GlobalString(GethFlag.Name)
		parityBin  = c.GlobalString(ParityFlag.Name)
		numThreads = c.GlobalInt(ThreadFlag.Name)
		numTests   uint64
	)
	fmt.Printf("numThreads: %d\n", numThreads)
	var wg sync.WaitGroup
	// The channel where we'll deliver tests
	testCh := make(chan string, 2)
	// Cancel ability
	sigs := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	// Thread that creates tests, spits out filenames
	for i := 0; i < 1; i++ {
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Generator started \n")
			base := fuzzing.GenerateBlake()
			target := base.GetDestination()
			prefix := fmt.Sprintf("blake-%d", id)

			for i := 0; ; i++ {
				// Generate new code
				base.SetCode(target, fuzzing.RandCallBlake())
				testName := fmt.Sprintf("%v-blaketest-%d", prefix, i)
				test := base.ToGeneralStateTest(testName)

				fileName, err := dumpTest(test, testName)
				if err != nil {
					fmt.Printf("Error occurred: %v", err)
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
	for i := 0; i < numThreads; i++ {
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
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				n := atomic.LoadUint64(&numTests)
				timeSpent := time.Since(tStart)
				execPerSecond := float64(uint64(time.Second)*n) / float64(timeSpent)
				fmt.Printf("%d tests executed, in %v (%.02f tests/s)\n", n, timeSpent, execPerSecond)
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

// dumpTest saves a testcase to disk
func dumpTest(test *fuzzing.GeneralStateTest, testName string) (string, error) {

	fileName := fmt.Sprintf("%v.json", testName)
	fullPath := path.Join("/tmp/", fileName)

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
