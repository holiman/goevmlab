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
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/goevmlab/ops"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/goevmlab/program"
)

func main() {

	if err := program.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runit() error {
	a := program.NewProgram()

	aAddr := common.HexToAddress("0xff0a")

	// Calling contract: Call contract B in a loop
	a.Op(ops.PC)         // Push 0
	a.Op(ops.DUP1)       // outsize = 0, on next iteration we use the return value of CALL
	dest := a.Jumpdest() // Loop Head
	a.Op(ops.DUP2)       // outoffset = 0
	a.Push(1305700)      // insize = 1305700
	a.Op(ops.DUP2)       // inoffset = 0
	a.Push(0xdeadbeef)   // Push target address, alternatively we could call an empty contract here
	a.Op(ops.GAS)        // Pass along all gas
	a.Op(ops.STATICCALL)
	a.Jump(dest) // Jump back

	alloc := make(core.GenesisAlloc)
	alloc[aAddr] = core.GenesisAccount{
		Nonce:   0,
		Code:    a.Bytecode(),
		Balance: big.NewInt(0xffffffff),
	}
	//-------------

	outp, err := json.MarshalIndent(alloc, "", " ")
	if err != nil {
		fmt.Printf("error : %v", err)
		os.Exit(1)
	}
	fmt.Printf("output \n%v\n", string(outp))
	//----------
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
		GasLimit:    10000000,
		Difficulty:  big.NewInt(0x200000),
		Time:        new(big.Int).SetUint64(0),
		Coinbase:    common.Address{},
		BlockNumber: new(big.Int).SetUint64(1),
		ChainConfig: &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			EIP150Block:         new(big.Int),
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
		},
		EVMConfig: vm.Config{
			Debug:  true,
			Tracer: &dumbTracer{},
		},
	}
	// Run with tracing
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	// Diagnose it
	runtimeConfig.EVMConfig = vm.Config{}
	t0 := time.Now()
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	t1 := time.Since(t0)
	fmt.Printf("Time elapsed: %v\n", t1)
	return err
}

type dumbTracer struct {
	counter uint64
}

func (d *dumbTracer) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
	if op == vm.STATICCALL {
		d.counter++
	}
	if op == vm.EXTCODESIZE {
		d.counter++
	}
}

func (d *dumbTracer) CaptureEnter(typ vm.OpCode, from common.Address, to common.Address, input []byte, gas uint64, value *big.Int) {
}

func (d *dumbTracer) CaptureExit(output []byte, gasUsed uint64, err error) {
}

func (d *dumbTracer) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
	fmt.Printf("CaptureFault %v\n", err)
}

func (d *dumbTracer) CaptureStart(env *vm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	fmt.Printf("captureStart\n")
	fmt.Printf("	from: %v\n", from.Hex())
	fmt.Printf("	to: %v\n", to.Hex())
}

func (d *dumbTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) {
	fmt.Printf("\nCaptureEnd\n")
	fmt.Printf("Counter: %d\n", d.counter)
}
