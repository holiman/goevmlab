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
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	common2 "github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func main() {
	if err := program.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func runit() error {
	a := program.NewProgram()
	// "input" is (hash, v, r, s), each 32 bytes
	hash := make([]byte, 32)
	v := make([]byte, 32)
	r := make([]byte, 32)
	s := make([]byte, 32)
	copy(v, hexutil.MustDecode("0x000000000000000000000000000000000000000000000000000000000000001b"))
	copy(r, hexutil.MustDecode("0x79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"))
	copy(s, hexutil.MustDecode("0x6b8d2c81b11b2d699528dde488dbdf2f94293d0d33c32e347f255fa4a6c1f0a9"))
	copy(hash, hexutil.MustDecode("0x6b8d2c81b11b2d699528dde488dbdf2f94293d0d33c32e347f255fa4a6c1f0a9"))

	a.Mstore(hash, 0)
	a.Mstore(v, 32)
	a.Mstore(r, 64)
	a.Mstore(s, 96)

	a.Call(big.NewInt(1_000_000),
		1,
		big.NewInt(0),   // value
		big.NewInt(0),   // inoffset
		big.NewInt(128), // insize
		big.NewInt(0),   // outoffset
		big.NewInt(32),  // outsize
	)
	a.Op(ops.POP)
	// Move the output (mem 0:32) into the stack
	a.Push(0)
	a.Op(ops.MLOAD)
	a.Push(0)
	a.Op(ops.SSTORE)
	aAddr := common.HexToAddress("0xff0a")
	alloc := make(core.GenesisAlloc)
	alloc[aAddr] = core.GenesisAccount{
		Nonce:   0,
		Code:    a.Bytecode(),
		Balance: big.NewInt(0xffffffff),
	}
	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		sender     = common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")
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
	var (
		gas  = uint64(10_000_000)
		fork = "London"
	)
	ruleset, err := ops.LookupChainConfig(fork)
	if err != nil {
		panic(err)
	}
	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    gas,
		Difficulty:  big.NewInt(0x200000),
		BlockNumber: new(big.Int).SetUint64(1),
		ChainConfig: ruleset,
		EVMConfig: vm.Config{
			Tracer: logger.NewJSONLogger(nil, os.Stderr),
		},
	}
	// Diagnose it
	t0 := time.Now()
	_, _, err = runtime.Call(aAddr, nil, &runtimeConfig)
	t1 := time.Since(t0)
	fmt.Printf("\nExecution time: %v\n", t1)
	if err != nil {
		fmt.Printf("Execution ended on error: %v\n", err)
	} else {
		fmt.Printf("Execution ended without error\n")
	}
	return common2.ConvertToStateTest("ecRecoverTest", fork, alloc, gas, aAddr)
}
