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
	"fmt"
	"math/big"
	"os"
	"testing"

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

func staticCallAttack() []byte {
	// Causes 13928 staticcalls
	//30          39521949 ns/op // 39 ms
	//  9            112185165 ns/op // 112 ms
	a, dest := program.New().Jumpdest()
	reps := 800
	for i := 0; i < reps; i++ {
		a.Push(0)
		a.Op(vm.DUP1, vm.DUP1, vm.DUP1, vm.GAS, vm.GAS, vm.STATICCALL, vm.POP)
	}
	return a.Jump(dest).Bytes()
}

func extCodeSizeAttack() []byte {
	// Causes 14205 EXTCODESIZEs
	//       63          17686763 ns/op

	a, dest := program.New().Jumpdest()
	reps := 800
	for i := 0; i < reps; i++ {
		a.Op(vm.GAS, vm.EXTCODESIZE, vm.POP)
	}
	return a.Jump(dest).Bytes()

}

func runit() error {
	/*
		if opcode == "FA":
		code = "5b%s600056" % (("60008080805a5a%s50"%opcode) * code_repitions)  # JUMPDEST (PUSH 0 DUP1 DUP1 DUP1 GAS GAS STATICCALL POP) * 8'000 PUSH1 0x0 JUMP
		else:
		code = "5b%s600056" % (("5a%s50"%opcode) * code_repitions)  # JUMPDEST (GAS BALANCE/EXTCODESIZE/EXTCODEHASH POP) * 8'000 PUSH1 0x0 JUMP
	*/
	code := staticCallAttack()
	if false {
		code = extCodeSizeAttack()
	}

	aAddr := common.HexToAddress("0xff0a")
	alloc := make(types.GenesisAlloc)
	alloc[aAddr] = types.Account{
		Nonce:   0,
		Code:    code,
		Balance: big.NewInt(0xffffffff),
	}
	var err error
	//-------------

	//outp, err := json.MarshalIndent(alloc, "", " ")
	//if err != nil {
	//	fmt.Printf("error : %v", err)
	//	os.Exit(1)
	//}
	//fmt.Printf("output \n%v\n", string(outp))
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
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	// Diagnose it
	runtimeConfig.EVMConfig = vm.Config{}
	res := testing.Benchmark(func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)

		}
	})
	//t0 := time.Now()
	//t1 := time.Since(t0)
	fmt.Print(res.String())
	fmt.Println()
	//fmt.Printf("Time elapsed: %v\n", t1)
	return err
}

type dumbTracer struct {
	counter uint64
}

func (d *dumbTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnTxStart: func(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
			fmt.Printf("OnTxStart\n")
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
			fmt.Printf("\nOnTxEnd\n")
			fmt.Printf("Counter: %d\n", d.counter)
		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			fmt.Printf("OnFault %v\n", err)
		},
	}
}
