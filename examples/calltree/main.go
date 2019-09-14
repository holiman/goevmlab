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


package main

import (
	"encoding/json"
	"fmt"
	"github.com/holiman/goevmlab/ops"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/goevmlab/program"
)

type dumbTracer struct {
	counter uint64
}

func (d *dumbTracer) CaptureStart(from common.Address, to common.Address, call bool, input []byte, gas uint64, value *big.Int) error {
	fmt.Printf("captureStart\n")
	fmt.Printf("	from: %v\n", from.Hex())
	fmt.Printf("	to: %v\n", to.Hex())
	return nil
}

func (d *dumbTracer) CaptureState(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, contract *vm.Contract, depth int, err error) error {
	if op == vm.CALL {
		if depth == 1 {
			fmt.Println("")
		} else {
			d.counter++
		}
		if depth < 2 {
			fmt.Printf("(%d: %d)", depth, gas)
		}

	}
	return nil
}

func (d *dumbTracer) CaptureFault(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, contract *vm.Contract, depth int, err error) error {
	fmt.Printf("CaptureFault\n")
	return nil
}

func (d *dumbTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error {
	fmt.Printf("\nCaptureEnd\n")
	fmt.Printf("Counter: %d\n", d.counter)
	return nil
}

func main() {

	if err := program.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func runit() error {
	a := program.NewProgram()
	b := program.NewProgram()

	aAddr := common.HexToAddress("0xff0a")
	bAddr := common.HexToAddress("0xff0b")

	dest := a.Jumpdest()
	a.Call(nil, bAddr, nil, 0, 0, 0, 0)
	a.Jump(dest)

	// The self-call can be done a bit more clever, gas-wise

	b.Op(ops.PC)      // get zero on stack (out size)
	b.Op(ops.DUP1)    // out offset
	b.Op(ops.DUP1)    // insize
	b.Op(ops.DUP1)    // inoffset
	b.Op(ops.DUP1)    // value
	b.Op(ops.ADDRESS) // address
	b.Op(ops.GAS)     // Gas
	b.Op(ops.CALL)

	alloc := make(core.GenesisAlloc)
	alloc[aAddr] = core.GenesisAccount{
		Nonce:   0,
		Code:    a.Bytecode(),
		Balance: big.NewInt(0xffffffff),
	}
	alloc[bAddr] = core.GenesisAccount{
		Nonce:   0,
		Code:    b.Bytecode(),
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
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()))
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
		//EVMConfig: vm.Config{
		//	Debug:  true,
		//	Tracer: &dumbTracer{},
		//},
	}
	// Diagnose it
	t0 := time.Now()
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	t1 := time.Since(t0)
	fmt.Printf("Time elapsed: %v\n", t1)
	t0 = time.Now()
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	t1 = time.Since(t0)
	fmt.Printf("Time elapsed: %v\n", t1)
	return err
}
