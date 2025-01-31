// Copyright 2021 Martin Holst Swende, Hubert Ritzdorf
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
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	common2 "github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/ops"
	"github.com/urfave/cli/v2"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Generator for jumpdest analysis tests"
	return app
}

var (
	app            = initApp()
	multiplierFlag = &cli.IntFlag{
		Name:  "multiplier",
		Value: 65, // 65 * 0x6000 = 1_597_440 bytes
		Usage: "Sets the memory payload size, which becomes 'multiplier x 0x6000'",
	}
	gasFlag = &cli.IntFlag{
		Name:  "gas",
		Value: 10_000_000,
		Usage: "Sets the gas amount to use",
	}
	outFileFlag = &cli.StringFlag{
		Name:  "out",
		Usage: "If set, causes a state-test to be written with the given name.",
	}
	forkFlag = &cli.StringFlag{
		Name:  "fork",
		Value: "Berlin",
		Usage: "What fork rules to use (e.g. Berlin, London)",
	}
	payloadFlag = &cli.IntFlag{
		Name:  "payload",
		Value: 0x5b,
		Usage: "What opcode to fill the space with.",
	}
	evaluateCommand = &cli.Command{
		Action:      evaluate,
		Name:        "evaluate",
		Usage:       "evaluate the test using the built-in go-ethereum base",
		Description: `Evaluate the test using the built-in go-ethereum library.`,
	}
)

func init() {
	app.Flags = []cli.Flag{
		multiplierFlag,
		gasFlag,
		forkFlag,
		payloadFlag,
		outFileFlag,
	}
	app.Commands = []*cli.Command{
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
		// codeMultiplier determines the size of the memory area to use.
		// The resulting size will be multiplied by 0x6000
		codeMultiplier = ctx.Int(multiplierFlag.Name)
		// gas to use for the tx
		gas          = uint64(ctx.Int(gasFlag.Name))
		outFilename  = ctx.String(outFileFlag.Name)
		fork         = ctx.String(forkFlag.Name)
		payload      = ctx.Int(payloadFlag.Name)
		initCodeSize = codeMultiplier * 0x6000
		// initcode is different from the mainnet attack code, which just did
		// a small jump across the first bytes and then stopped. This code instead
		// jumps to the last byte of the code, forcing memory reads more all over the place.
		// It also prevents optimistic jumpdest analysis optimisation based on not analyzing unless
		// jumping to (or across) it
		initcode = common.HexToHash(fmt.Sprintf("0x62%v565b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b5b",
			strconv.FormatInt(int64(initCodeSize-1), 16)))
		// Where the payload resides
		payloadAddr = common.HexToAddress("0xc0de")
		// The attacker code
		attackerAddr = common.HexToAddress("0x31337")
	)
	ruleset, err := ops.LookupChainConfig(fork)
	if err != nil {
		return err
	}
	if payload > 255 {
		return fmt.Errorf("payload %d is not a byte opcode", payload)
	}

	// Create the full payload of 0x6000, the max codesize
	payloadCode := make([]byte, 0x6000)
	for i := range payloadCode {
		payloadCode[i] = byte(payload)
	}
	// We need to fill the last 32 byte with something other than push, to be
	// sure that the very last jumpdest is a _valid_ jumpdest.
	ending := payloadCode[len(payloadCode)-33:]
	for i := range ending {
		ending[i] = byte(ops.JUMPDEST)
	}
	// And dump it into state
	alloc := make(types.GenesisAlloc)
	alloc[payloadAddr] = types.Account{
		Nonce: 1,
		Code:  payloadCode,
	}
	// The main show.
	a := program.New()
	fmt.Printf(`
Payload: %v
Init code size: %d (multiplier %d)
Gas to use: %d
Fork: %v
`, ops.OpCode(payload), initCodeSize, codeMultiplier, gas, fork)

	a.Mstore([]byte{0}, uint32(initCodeSize)) // Expand memory
	for i := 0; i <= codeMultiplier; i++ {
		a.ExtcodeCopy(payloadAddr, i*0x6000, 0, 0x6000)
	}
	// Memory filled with payload. We insert actual runnable code at the beginning
	a.Mstore(initcode[:], 0)
	// Memory is now filled. Phase two, creation mode.
	_, loc := a.Jumpdest()
	a.Push(initCodeSize).Push(0 /* offset */).Op(vm.DUP1 /*0 value*/)
	a.Op(vm.CREATE)
	a.Op(vm.POP)
	// flip a bit
	a.Push(0).Op(vm.MLOAD)
	a.Push(1).Op(vm.ADD)    //
	a.Push(0).Op(vm.MSTORE) // mem location
	a.Jump(loc)

	alloc[attackerAddr] = types.Account{
		Nonce:   1,
		Code:    a.Bytes(),
		Balance: big.NewInt(0xffffffff),
	}
	var (
		statedb = common2.StateDBWithAlloc(alloc)
		sender  = common.BytesToAddress([]byte("sender"))
	)
	statedb.CreateAccount(sender)
	tracer := &customTracer{startGas: gas}
	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    gas,
		Difficulty:  big.NewInt(0x200000),
		BlockNumber: new(big.Int).SetUint64(1),
		ChainConfig: ruleset,
		EVMConfig: vm.Config{
			Tracer: tracer.Hooks(),
		},
	}
	// Run with tracing
	_, _, _ = runtime.Call(attackerAddr, nil, &runtimeConfig)
	// Diagnose it
	runtimeConfig.EVMConfig = vm.Config{}
	t0 := time.Now()
	_, _, err = runtime.Call(attackerAddr, nil, &runtimeConfig)
	t1 := time.Since(t0)

	fmt.Printf("Amount of code analyzed (CREATE x initcode size): %.02f MB",
		float64(tracer.createCount*uint64(initCodeSize))/(1024*1024))

	fmt.Printf("\nExecution time: %v\n", t1)
	if err != nil {
		fmt.Printf("Execution ended on error: %v\n", err)
	} else {
		fmt.Printf("Execution ended without error\n")
	}
	if len(outFilename) == 0 {
		return nil
	}
	return convertToStateTest(outFilename, fork, alloc, gas, attackerAddr)
}

// convertToStateTest is a utility to turn stuff into sharable state tests.
func convertToStateTest(name, fork string, alloc types.GenesisAlloc, gasLimit uint64,
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
	if err := os.WriteFile(fname, dat, 0777); err != nil {
		return err
	}
	fmt.Printf("Wrote file %v\n", fname)
	return nil
}

type customTracer struct {
	createCount uint64
	copyCount   uint64
	memSize     uint64
	opCount     uint64

	startGas  uint64
	phase1Gas uint64
	phase2Gas uint64
}

func (d *customTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnOpcode: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
			d.opCount++
			if vm.OpCode(op) == vm.EXTCODECOPY {
				d.copyCount++
				if d.phase1Gas == 0 {
					d.memSize = uint64(len(scope.MemoryData()))
					d.phase1Gas = gas
				}
			}
			if vm.OpCode(op) == vm.CREATE {
				d.createCount++
				if d.phase2Gas == 0 {
					d.phase2Gas = gas
				}
			}
		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			fmt.Printf("CaptureFault %v\n", err)
		},
		OnTxEnd: func(receipt *types.Receipt, err error) {
			fmt.Printf(`
# Stats

Phase 1 - Mem expansion cost: %d (%d bytes expansion) 
Phase 2 - Code copying cost : %d (%d EXTCODECOPY calls)
Phase 3 - initcode-exec cost: %d (%d CREATE calls)

Total steps: %d
Total gas spent: %d
`, d.startGas-d.phase1Gas, d.memSize,
				d.phase1Gas-d.phase2Gas, d.copyCount,
				d.phase2Gas, d.createCount,
				d.opCount,
				d.startGas)
		},
	}
}
