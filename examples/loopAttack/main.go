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
	app.Usage = "Generator for loop analysis tests"
	return app
}

var (
	app     = initApp()
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
		Value: "London",
		Usage: "What fork rules to use (e.g. Berlin, London)",
	}
	pushFlag = &cli.IntFlag{
		Name:  "push",
		Value: int(ops.PC),
		Usage: "What opcode to use for filling the stack with. These are all 2-gas ops: " +
			"MSIZE(0x59), GAS(0x5A), RETURNDATASIZE(0x3D), ADDRESS(0x30), ORIGIN(0x32)," +
			"CALLER(0x33), CALLVALUE(0x34), CALLDATASIZE(0x36), CODESIZE(0x38), GASPRICE(0x3A), COINBASE(0x41)," +
			"TIMESTAMP(0x42), NUMBER(0x43), DIFFICULTY(0x44), GASLIMIT(0x45), " +
			"CHAINID(0x46), BASEFEE(0x48)",
	}
	popFlag = &cli.IntFlag{
		Name:  "pop",
		Value: int(ops.POP),
		Usage: "What opcode to use for emptying the stack with.",
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
		gasFlag,
		forkFlag,
		pushFlag,
		popFlag,
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
		// gas to use for the tx
		gas         = uint64(ctx.Int(gasFlag.Name))
		outFilename = ctx.String(outFileFlag.Name)
		fork        = ctx.String(forkFlag.Name)
		pusher      = ctx.Int(pushFlag.Name)
		popper      = ctx.Int(popFlag.Name)
		// The attacker code
		attackerAddr = common.HexToAddress("0x31337")
	)
	ruleset, err := ops.LookupChainConfig(fork)
	if err != nil {
		return err
	}
	if pusher > 255 {
		return fmt.Errorf("pusher %d is not a byte opcode", pusher)
	}
	if popper > 255 {
		return fmt.Errorf("popper %d is not a byte opcode", popper)
	}
	a, b := ops.OpCode(byte(pusher)), ops.OpCode(byte(popper))
	if a.Stackdelta()+b.Stackdelta() != 0 {
		return fmt.Errorf("operations %v (stackdelta %d) and %v (stackdelta %d) do not balance push/pop",
			a, a.Stackdelta(), b, b.Stackdelta())
	}
	pushpop := program.New()
	if a == ops.STOP {
		// No filling used
	} else if a.Stackdelta() > 0 {
		stack := 0
		for stack < 1023 {
			pushpop.Op(vm.OpCode(a))
			stack += a.Stackdelta()
		}
		for stack > 0 {
			pushpop.Op(vm.OpCode(b))
			stack += b.Stackdelta()
		}
	} else {
		for i := 0; i < 1024; i++ {
			pushpop.Op(vm.OpCode(a))

		}
		for i := 0; i < 1024; i++ {
			pushpop.Op(vm.OpCode(b))
		}
	}
	payload := program.New()
	_, start := payload.Jumpdest()
	payload.Append(pushpop.Bytes())
	payload.Jump(start)

	// And dump it into state
	alloc := make(types.GenesisAlloc)
	desc := fmt.Sprintf(`
Pusher: %v
Popper: %v
Gas to use: %d
Fork: %v
`, a, b, gas, fork)
	fmt.Println(desc)
	alloc[attackerAddr] = types.Account{
		Nonce:   1,
		Code:    payload.Bytes(),
		Balance: big.NewInt(0xffffffff),
	}
	var (
		statedb = common2.StateDBWithAlloc(alloc)
		sender  = common.BytesToAddress([]byte("sender"))
	)
	statedb.CreateAccount(sender)
	tracer := &customTracer{}
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
	jumpCount uint64
	opCount   uint64
}

func (d *customTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnOpcode: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
			d.opCount++
			if vm.OpCode(op) == vm.JUMP {
				d.jumpCount++
			}

		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			fmt.Printf("CaptureFault %v\n", err)
		},
		OnTxEnd: func(receipt *types.Receipt, err error) {
			gasUsed := receipt.GasUsed
			fmt.Printf(`
# Stats

Total steps: %d
Avg gas/op : %.02f
Gas used   : %d
Jumps made : %d
`, d.opCount, float64(gasUsed)/float64(d.opCount), gasUsed,
				d.jumpCount)
		},
	}
}
