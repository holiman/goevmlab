// Copyright 2026 Spencer Taylor-Brown (terminus-31)
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fuzzing

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
)

// fillDeepCalls generates a state test with a deep CALL chain (5-50
// frames) where each contract calls the next and the leaf does SSTORE.
// Exercises the EIP-8037 reservoir-passing-through-frames pathway at
// depths well beyond what RandCall2200's compile-time-recursion cap of
// 10 reaches. See test_state_gas_call.py::test_nested_calls_*.
func fillDeepCalls(gst *GstMaker, fork string) {
	depth := 5 + rand.Intn(46) // 5..50
	contracts := make([]common.Address, depth)
	for i := 0; i < depth; i++ {
		contracts[i] = common.HexToAddress(fmt.Sprintf("0xc%03x", i))
	}

	for i := 0; i < depth; i++ {
		p := program.New()
		if i == depth-1 {
			// Leaf: SSTORE so the deepest frame draws state gas
			// from (the parent chain's) reservoir.
			p.Sstore(0, 1)
			p.Op(vm.STOP)
		} else {
			// Intermediate: CALL the next contract, forwarding all
			// gas. CALL stack (top → bottom): gas, to, value, argOff,
			// argSize, retOff, retSize.
			p.Push(0) // retSize
			p.Push(0) // retOff
			p.Push(0) // argSize
			p.Push(0) // argOff
			p.Push(0) // value
			p.Push(contracts[i+1])
			p.Op(vm.GAS)
			p.Op(vm.CALL)
			p.Op(vm.POP)
			p.Op(vm.STOP)
		}
		gst.AddAccount(contracts[i], GenesisAccount{
			Code:    p.Bytes(),
			Balance: new(big.Int),
			Storage: make(map[common.Hash]common.Hash),
		})
	}

	gasLimit := gasLimitOptions8037[rand.Intn(len(gasLimitOptions8037))]
	tx := &StTransaction{
		GasLimit:             []uint64{gasLimit},
		Nonce:                0,
		Value:                []string{"0x0"},
		Data:                 []string{"0x"},
		MaxFeePerGas:         big.NewInt(0x10),
		MaxPriorityFeePerGas: big.NewInt(0x10),
		To:                   contracts[0].Hex(),
		Sender:               sender,
		PrivateKey: hexutil.MustDecode(
			"0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8",
		),
	}
	gst.SetTx(tx)
}
