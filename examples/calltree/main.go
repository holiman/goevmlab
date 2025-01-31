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
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
	common2 "github.com/holiman/goevmlab/common"
	program2 "github.com/holiman/goevmlab/utils"
)

type customTracer struct {
	counter uint64
}

func (d *customTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnOpcode: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
			if vm.OpCode(op) == vm.CALL {
				if depth == 1 {
					fmt.Println("")
				} else {
					d.counter++
				}
				if depth < 2 {
					fmt.Printf("(%d: %d)", depth, gas)
				}
			}
		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			fmt.Printf("CaptureFault %v\n", err)
		},
		OnTxStart: func(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
			fmt.Printf("captureStart\n")
			fmt.Printf("	from: %v\n", from.Hex())
			fmt.Printf("	to: %v\n", tx.To())

		},
		OnTxEnd: func(receipt *types.Receipt, err error) {
			fmt.Printf("\nCaptureEnd\n")
			fmt.Printf("Counter: %d\n", d.counter)
		},
	}
}

func main() {

	if err := program2.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func runit() error {
	a := program.New()
	b := program.New()

	aAddr := common.HexToAddress("0xff0a")
	bAddr := common.HexToAddress("0xff0b")

	_, dest := a.Jumpdest()
	a.Call(nil, bAddr, nil, 0, 0, 0, 0)
	a.Jump(dest)

	// The self-call can be done a bit more clever, gas-wise

	b.Op(vm.PC)              // get zero on stack (out size)
	b.Op(vm.DUP1, vm.DUP1)   // out offset, insize
	b.Op(vm.DUP1, vm.DUP1)   // inoffset, value
	b.Op(vm.ADDRESS, vm.GAS) // address, gas
	b.Op(vm.CALL)

	alloc := make(types.GenesisAlloc)
	alloc[aAddr] = types.Account{
		Nonce:   0,
		Code:    a.Bytes(),
		Balance: big.NewInt(0xffffffff),
	}
	alloc[bAddr] = types.Account{
		Nonce:   0,
		Code:    b.Bytes(),
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
	var vmConf vm.Config
	if true {
		vmConf = vm.Config{
			Tracer: new(customTracer).Hooks(),
		}
	}
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
		EVMConfig: vmConf,
	}
	// Diagnose it
	t0 := time.Now()
	_, _, _ = runtime.Call(aAddr, nil, &runtimeConfig)
	t1 := time.Since(t0)
	fmt.Printf("Time elapsed: %v\n", t1)
	t0 = time.Now()
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	t1 = time.Since(t0)
	fmt.Printf("Time elapsed: %v\n", t1)
	return err
}
