// Copyright 2026 Martin Holst Swende
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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/evms"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/urfave/cli/v2"
)

// Blocktest modes, as given to the --blocktest flag.
const (
	BlockTestOff       = "off"
	BlockTestConstruct = "construct"
	BlockTestConsume   = "consume"
	BlockTestBoth      = "both"
)

// engineSuffix is the filename suffix of blocktests in engine flavour.
const engineSuffix = ".engine.json"

var (
	BlockTestFlag = &cli.StringFlag{
		Name: "blocktest",
		Usage: "Generate and execute blocktests instead of statetests. Modes:\n" +
			"  'off': statetests (default)\n" +
			"  'construct': blockchain_test fixtures. Blocks are given as RLP, and the clients construct the block access list (EIP-7928) themselves.\n" +
			"  'consume': blockchain_test_engine fixtures. Blocks are given as engine payloads including the block access list, exercising the BAL-driven (parallel) execution. Executed without tracing.\n" +
			"  'both': generate both flavours from each test.",
		Value: BlockTestOff,
	}
	BlocksFlag = &cli.IntFlag{
		Name:  "blocks",
		Usage: "Number of blocks per blocktest",
		Value: 2,
	}
	TxsPerBlockFlag = &cli.IntFlag{
		Name:  "txsperblock",
		Usage: "Number of transactions per block in blocktests",
		Value: 3,
	}
	BlockTestFlags = []cli.Flag{
		BlockTestFlag,
		BlocksFlag,
		TxsPerBlockFlag,
	}
)

// TestKind is the kind of a test file.
type TestKind int

const (
	StateTest TestKind = iota
	BlockTest
)

func (k TestKind) String() string {
	if k == BlockTest {
		return "blocktest"
	}
	return "statetest"
}

// testMode describes how a test file is to be executed.
type testMode struct {
	kind      TestKind
	engine    bool // for blocktests: engine flavour
	skipTrace bool
}

// blockTestMode returns the evms mode for this testMode.
func (m testMode) blockTestMode() evms.BlockTestMode {
	return evms.BlockTestMode{
		Engine: m.engine,
		// The engine flavour exercises parallel execution, which cannot be traced.
		Trace: !m.skipTrace && !m.engine,
	}
}

// modeForFile determines the execution mode of a test file, given whether
// blocktests are expected. Engine-flavour blocktests are recognized by
// their filename suffix.
func modeForFile(file string, blockTests, skipTrace bool) testMode {
	m := testMode{skipTrace: skipTrace}
	if blockTests {
		m.kind = BlockTest
		m.engine = strings.HasSuffix(file, engineSuffix)
	}
	return m
}

// runTest executes the given test on the evm, according to the mode.
func runTest(evm evms.Evm, m testMode, file string, out io.Writer) (*evms.TracingResult, error) {
	if m.kind == BlockTest {
		return evm.RunBlockTest(file, out, m.blockTestMode())
	}
	return evm.RunStateTest(file, out, m.skipTrace)
}

// BlockTestsRequested returns true if the --blocktest flag selects blocktests.
func BlockTestsRequested(c *cli.Context) bool {
	mode := c.String(BlockTestFlag.Name)
	return mode != "" && mode != BlockTestOff
}

// filterBlockTestVMs drops the vms which cannot execute blocktests.
func filterBlockTestVMs(vms []evms.Evm) []evms.Evm {
	var res []evms.Evm
	for _, vm := range vms {
		if evms.SupportsBlockTests(vm) {
			res = append(res, vm)
		} else {
			log.Warn("Dropping vm, blocktests not supported", "vm", vm.Name())
			vm.Close()
		}
	}
	return res
}

// DetectTestKind determines whether the given file is a state test or a
// blocktest, by inspecting its content.
func DetectTestKind(file string) (TestKind, error) {
	if strings.HasSuffix(file, engineSuffix) {
		return BlockTest, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return StateTest, err
	}
	var tests map[string]struct {
		Blocks      json.RawMessage `json:"blocks"`
		Payloads    json.RawMessage `json:"engineNewPayloads"`
		Transaction json.RawMessage `json:"transaction"`
	}
	if err := json.Unmarshal(data, &tests); err != nil {
		return StateTest, err
	}
	for _, t := range tests {
		if len(t.Blocks) > 0 || len(t.Payloads) > 0 {
			return BlockTest, nil
		}
		if len(t.Transaction) > 0 {
			return StateTest, nil
		}
	}
	return StateTest, errors.New("unrecognized test format")
}

// BtGeneratorFn generates blocktest makers.
type BtGeneratorFn func() (*fuzzing.BtMaker, error)

// WriteBlockTest builds the blocktest and writes the flavours selected by the
// mode ('construct', 'consume' or 'both') to the given directory. The paths of
// the written files are returned. If indent is set, the JSON is indented.
func WriteBlockTest(bt *fuzzing.BtMaker, dir, name, mode string, indent bool) ([]string, error) {
	if err := bt.Build(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	var (
		files []string
		write = func(fname string, v any) error {
			fullPath := path.Join(dir, fname)
			f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
			if err != nil {
				return err
			}
			defer f.Close()
			encoder := json.NewEncoder(f)
			if indent {
				encoder.SetIndent("", " ")
			}
			if err := encoder.Encode(v); err != nil {
				return err
			}
			files = append(files, fullPath)
			return nil
		}
	)
	if mode == BlockTestConstruct || mode == BlockTestBoth {
		test, err := bt.ToBlockchainTest(name)
		if err != nil {
			return nil, err
		}
		if err := write(fmt.Sprintf("%v.json", name), test); err != nil {
			return nil, err
		}
	}
	if mode == BlockTestConsume || mode == BlockTestBoth {
		test, err := bt.ToEngineTest(name)
		if err != nil {
			return nil, err
		}
		if err := write(fmt.Sprintf("%v%v", name, engineSuffix), test); err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("unknown blocktest mode %q", mode)
	}
	return files, nil
}

// btTestFnFromGenerator creates a test provider from a blocktest generator.
// Since one generated test may result in several files (flavours), the
// provider keeps a per-thread queue of files to deliver.
func btTestFnFromGenerator(fn BtGeneratorFn, name, location, mode string) TestProviderFn {
	var (
		mu      sync.Mutex
		pending = make(map[int][]string)
	)
	return func(index, threadId int) (string, error) {
		mu.Lock()
		if q := pending[threadId]; len(q) > 0 {
			pending[threadId] = q[1:]
			mu.Unlock()
			return q[0], nil
		}
		mu.Unlock()
		// Generation can fail if e.g. a filler produces a transaction which
		// cannot be applied in a block. Retry a few times.
		for attempt := 0; ; attempt++ {
			bt, err := fn()
			if err != nil {
				return "", err
			}
			testName := fmt.Sprintf("%08d-%v-%d", index, name, threadId)
			files, err := WriteBlockTest(bt, location, testName, mode, false)
			if err != nil {
				if attempt < 10 {
					log.Debug("Blocktest generation failed, retrying", "err", err)
					continue
				}
				return "", err
			}
			mu.Lock()
			pending[threadId] = append(pending[threadId], files[1:]...)
			mu.Unlock()
			return files[0], nil
		}
	}
}

// GenerateAndExecuteBlockTests runs the fuzzing loop with blocktests produced
// by the given generator.
func GenerateAndExecuteBlockTests(c *cli.Context, generatorFn BtGeneratorFn, name string) error {
	mode := c.String(BlockTestFlag.Name)
	switch mode {
	case BlockTestConstruct, BlockTestConsume, BlockTestBoth:
	default:
		return fmt.Errorf("invalid blocktest mode %q", mode)
	}
	fn := btTestFnFromGenerator(generatorFn, name, c.String(LocationFlag.Name), mode)
	return ExecuteFuzzerWithOptions(c, fn, FuzzOptions{
		CleanupFiles: c.Bool(RemoveFilesFlag.Name),
		BlockTests:   true,
	})
}
