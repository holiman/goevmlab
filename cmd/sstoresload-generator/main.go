package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/goevmlab/fuzzing"
	"golang.org/x/crypto/sha3"
	"gopkg.in/urfave/cli.v1"
	"io"
	"os"
	"path"
	"path/filepath"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Martin Holst Swende"
	app.Usage = "Generator for tests targeting SSTORE and SLOAD"
	return app
}

var (
	app        = initApp()
	PrefixFlag = cli.StringFlag{
		Name:  "prefix",
		Usage: "prefix of output files",
	}
	LocationFlag = cli.StringFlag{
		Name:  "output-dir",
		Usage: "Location of where to place the generated files",
	}
	CountFlag = cli.IntFlag{
		Name:  "count",
		Usage: "number of tests to generate",
	}
)

func init() {
	app.Flags = []cli.Flag{
		PrefixFlag,
		LocationFlag,
		CountFlag,
	}
	app.Commands = []cli.Command{
		generateCommand,
	}
}

var generateCommand = cli.Command{
	Action:      generate,
	Name:        "generate",
	Usage:       "generate tests",
	ArgsUsage:   "<number of tests to generate>",
	Description: `The generate generates the tests.`,
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// fillTest uses go-ethereum internally to get the state root and logs
func fillTest(gst *fuzzing.GstMaker, traceOutput io.Writer) (root, logs common.Hash, err error) {
	test, err := gst.ToStateTest()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	subtest := test.Subtests()[0]
	cfg := vm.Config{}
	if traceOutput != nil {
		cfg.Debug = true
		cfg.Tracer = vm.NewJSONLogger(&vm.LogConfig{}, traceOutput)
	}
	statedb, _ := test.Run(subtest, cfg)

	root = statedb.IntermediateRoot(true)
	logs = rlpHash(statedb.Logs())

	fmt.Printf("root: %x, logs: %x\n", root, logs)
	return
}

func createTests(location, prefix string, limit int, trace bool) error {

	for i := 0; i < limit; i++ {
		testName := fmt.Sprintf("%v-storagetest-%04d", prefix, i)
		fileName := fmt.Sprintf("%v.json", testName)

		f, err := os.OpenFile(path.Join(location, "tests", fileName), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		close := func() {
			f.Close()
		}

		// Now, let's also dump out the trace, so we can investigate if the tests
		// are doing anything interesting
		if trace {
			traceName := fmt.Sprintf("%v-trace.jsonl", testName)
			traceOut, err := os.OpenFile(path.Join(location, "traces", traceName), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
			if err != nil {
				return err
			}
			close = func() {
				f.Close()
				traceOut.Close()
			}
		}

		// Generate new code
		base := fuzzing.Generate2200Test()

		// Get new state root and logs hash
		root, logs, err := fillTest(base, nil)
		if err != nil {
			close()
			return err
		}
		base.SetResult(root, logs)

		test := base.ToGeneralStateTest(testName)
		// Write to tfile
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", " ")
		if err = encoder.Encode(test); err != nil {
			close()
			return err
		}
		close()
	}
	return nil
}

func generate(ctx *cli.Context) error {

	var prefix = ""
	if ctx.GlobalIsSet(PrefixFlag.Name) {
		prefix = ctx.GlobalString(PrefixFlag.Name)
	}
	var location = ""
	if ctx.GlobalIsSet(LocationFlag.Name) {
		location = ctx.GlobalString(LocationFlag.Name)
		if err := os.MkdirAll(path.Join(location, "tests"), 0755); err != nil {
			return fmt.Errorf("could not create %v: %v", location, err)
		}
		//if err := os.MkdirAll(path.Join(location, "traces"), 0755); err != nil {
		//	return fmt.Errorf("could not create %v: %v", location, err)
		//}
	}
	var count = 0
	if ctx.GlobalIsSet(CountFlag.Name) {
		count = ctx.GlobalInt(CountFlag.Name)
	}
	return createTests(location, prefix, count, false)
}
