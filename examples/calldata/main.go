// Copyright 2019 Martin Holst Swende, Hubert Ritzdorf
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

package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	common2 "github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Usage = "Generator for STATICCALL-attacks with large data (1.3M) in the call."
	return app
}

var (
	app      = initApp()
	destFlag = cli.StringFlag{
		Name:  "destination",
		Value: "0xdead",
		Usage: "Destination address to make staticcall to",
	}
	gasFlag = cli.IntFlag{
		Name:  "gas",
		Value: 10_000_000,
		Usage: "Sets the gas amount to use",
	}
	outFileFlag = cli.StringFlag{
		Name:  "out",
		Usage: "If set, causes a state-test to be written with the given name.",
	}
	forkFlag = cli.StringFlag{
		Name:  "fork",
		Value: "London",
		Usage: "What fork rules to use (e.g. Berlin, London)",
	}
	evaluateCommand = cli.Command{
		Action:      evaluate,
		Name:        "evaluate",
		Usage:       "evaluate the test using the built-in go-ethereum base",
		Description: `Evaluate the test using the built-in go-ethereum library.`,
	}
)

func init() {
	app.Flags = []cli.Flag{
		destFlag,
		gasFlag,
		forkFlag,
		outFileFlag,
	}
	app.Commands = []cli.Command{
		evaluateCommand,
	}
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func evaluate(ctx *cli.Context) error {
	var (
		// gas to use for the tx
		gas         = uint64(ctx.GlobalInt(gasFlag.Name))
		outFilename = ctx.GlobalString(outFileFlag.Name)
		fork        = ctx.GlobalString(forkFlag.Name)
		dest        = ctx.GlobalString(destFlag.Name)
	)
	// Validate ruleset
	ruleset, ok := common2.Forks[fork]
	if !ok {
		var valid []string
		for n, _ := range common2.Forks {
			valid = append(valid, n)
		}
		return fmt.Errorf("fork '%v' not defined. Valid values are %v", fork, strings.Join(valid, ","))
	}
	destAddr := new(big.Int).SetBytes(common.FromHex(dest))
	fmt.Printf(`
Call-address: 0x%x
Gas to use: %d
Fork: %v
`, destAddr.Bytes(), gas, fork)

	a := program.NewProgram()
	a.Op(ops.PC)          // Push 0
	a.Op(ops.DUP1)        // outsize = 0, on next iteration we use the return value of CALL
	label := a.Jumpdest() // Loop Head
	a.Op(ops.DUP2)        // outoffset = 0
	a.Push(1305700)       // insize = 1305700
	a.Op(ops.DUP2)        // inoffset = 0
	a.Push(destAddr)
	//	a.Push(0xdeadbeef)   // Push target address, alternatively we could call an empty contract here
	a.Op(ops.GAS) // Pass along all gas
	a.Op(ops.STATICCALL)
	a.Jump(label) // Jump back

	aAddr := common.HexToAddress("0xff0a")
	alloc := make(core.GenesisAlloc)
	alloc[aAddr] = core.GenesisAccount{
		Nonce:   0,
		Code:    a.Bytecode(),
		Balance: big.NewInt(0xffffffff),
	}
	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		sender     = common.BytesToAddress([]byte("sender"))
	)
	for addr, acc := range alloc {
		statedb.CreateAccount(addr)
		statedb.SetCode(addr, acc.Code)
		statedb.SetNonce(addr, acc.Nonce)
		if acc.Balance != nil {
			statedb.SetBalance(addr, acc.Balance)
		}
	}
	statedb.CreateAccount(sender)

	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    gas,
		Difficulty:  big.NewInt(0x200000),
		Time:        new(big.Int).SetUint64(0),
		Coinbase:    common.Address{},
		BlockNumber: new(big.Int).SetUint64(1),
		ChainConfig: ruleset,
		EVMConfig: vm.Config{
			Debug:  true,
			Tracer: &dumbTracer{},
		},
	}
	// Run with tracing
	_, _, err := runtime.Call(aAddr, nil, &runtimeConfig)
	// Diagnose it
	runtimeConfig.EVMConfig = vm.Config{}
	t0 := time.Now()
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	t1 := time.Since(t0)
	fmt.Printf("\nExecution time: %v\n", t1)
	if err != nil {
		fmt.Printf("Execution ended on error: %v\n", err)
	} else {
		fmt.Printf("Execution ended without error\n")
	}
	if len(outFilename) == 0 {
		return nil
	}
	return convertToStateTest(outFilename, fork, alloc, gas, aAddr)
}

// convertToStateTest is a utility to turn stuff into sharable state tests.
func convertToStateTest(name, fork string, alloc core.GenesisAlloc, gasLimit uint64,
	target common.Address) error {

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
	if err := ioutil.WriteFile(fname, dat, 0777); err != nil {
		return err
	}
	fmt.Printf("Wrote file %v\n", fname)
	return nil
}

type dumbTracer struct {
	counter uint64
}

func (d *dumbTracer) CaptureStart(env *vm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	fmt.Printf("captureStart\n")
	fmt.Printf("	from: %v\n", from.Hex())
	fmt.Printf("	to: %v\n", to.Hex())
}

func (d *dumbTracer) CaptureState(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
	if op == vm.STATICCALL {
		d.counter++
	}
	if op == vm.EXTCODESIZE {
		d.counter++
	}
}

func (d *dumbTracer) CaptureFault(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
	fmt.Printf("CaptureFault %v\n", err)
}

func (d *dumbTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {
	fmt.Printf("\nCaptureEnd\n")
	fmt.Printf("STATICCALLs: %d\n", d.counter)
}
