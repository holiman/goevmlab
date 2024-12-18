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
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
	common2 "github.com/holiman/goevmlab/common"
	program2 "github.com/holiman/goevmlab/utils"
)

func main() {

	if err := program2.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func runit() error {
	a := program.New()

	aAddr := common.HexToAddress("0xff0a")

	/*
		nop
		JUMPDEST    // []
		MSIZE       // [0] out size
		MSIZE       // [0,0] out offset
		MSIZE       // [0,0,0] insize
		MSIZE       // [0,0,0,0] inoffset
		PUSH1 4     // [0,0,0,0,4] address
		GAS         // [0,0,0,0,4,gas] Gas
		STATICCALL  // [1] pops 6, pushes 1
		JUMP        // []

	*/

	a.Op(vm.PC, vm.JUMPDEST, vm.MSIZE, vm.MSIZE, vm.MSIZE, vm.MSIZE)
	a.Push(4)
	a.Op(vm.GAS, vm.STATICCALL, vm.JUMP)

	alloc := make(types.GenesisAlloc)
	alloc[aAddr] = types.Account{
		Nonce:   0,
		Code:    a.Bytes(),
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
		statedb = common2.StateDBWithAlloc(alloc)
		sender  = common.BytesToAddress([]byte("sender"))
	)
	statedb.CreateAccount(sender)

	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    10000000,
		Difficulty:  big.NewInt(0x200000),
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
			Tracer: new(dumbTracer).Hooks(),
		},
	}
	// Diagnose it
	t0 := time.Now()
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	t1 := time.Since(t0)
	fmt.Printf("Time elapsed: %v\n", t1)
	//for i:=0 ; i < 3; i++{
	//	t0 = time.Now()
	//	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	//	t1 = time.Since(t0)
	//	fmt.Printf("Time elapsed: %v\n", t1)
	//}
	return err
}

type dumbTracer struct {
	common2.BasicTracer
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

func (d *dumbTracer) CaptureStart(env *vm.EVM, from common.Address, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	fmt.Printf("captureStart\n")
	fmt.Printf("	from: %v\n", from.Hex())
	fmt.Printf("	to: %v\n", to.Hex())
}

func (d *dumbTracer) CaptureEnd(output []byte, gasUsed uint64, err error) {
	fmt.Printf("\nCaptureEnd\n")
	fmt.Printf("Counter: %d\n", d.counter)
}
