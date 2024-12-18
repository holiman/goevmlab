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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
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
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runit() error {
	a := program.New()

	aAddr := common.HexToAddress("0xff0a")

	// Calling contract: Call contract B in a loop
	a.Op(vm.PC)             // Push 0
	a.Op(vm.DUP1)           // outsize = 0, on next iteration we use the return value of CALL
	_, dest := a.Jumpdest() // Loop Head
	a.Op(vm.DUP2)           // outoffset = 0
	a.Push(1305700)         // insize = 1305700
	a.Op(vm.DUP2)           // inoffset = 0
	a.Push(0xdeadbeef)      // Push target address, alternatively we could call an empty contract here
	a.Op(vm.GAS)            // Pass along all gas
	a.Op(vm.STATICCALL)
	a.Jump(dest) // Jump back

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
	// Run with tracing
	_, _, _ = runtime.Call(aAddr, nil, &runtimeConfig)
	// Work-around missing tx-end, todo remove later
	runtimeConfig.EVMConfig.Tracer.OnTxEnd(nil, nil)
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

func (d *dumbTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnTxStart: func(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
			fmt.Printf("captureStart\n")
			fmt.Printf("	from: %v\n", from.Hex())
			fmt.Printf("	to: %v\n", tx.To().Hex())
		},
		OnOpcode: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
			if op == byte(vm.STATICCALL) {
				d.counter++
			}
			if op == byte(vm.EXTCODESIZE) {
				d.counter++
			}
		},
		OnTxEnd: func(receipt *types.Receipt, err error) {
			fmt.Printf("\nCaptureEnd\n")
			fmt.Printf("Counter: %d\n", d.counter)
		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			fmt.Printf("OnFault %v\n", err)
		},
		OnClose: func() {
			fmt.Printf("OnClose")
		},
	}
}
