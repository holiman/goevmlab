// Copyright 2026 Spencer Taylor-Brown (terminus-31)
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fuzzing

import (
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
)

// fillSstoreRestoration generates a state test that exercises the
// EIP-8037 SSTORE restoration refund pathway: a slot transitions
// 0 → x → 0 within a single transaction, refunding the
// (STATE_BYTES_PER_STORAGE_SET × CPSB = 97,920) state-gas charge on
// the initial write back to the reservoir.
//
// The contract runs N cycles (1-20) of (write nonzero, clear) across
// distinct slots, randomising the intermediate value. Inline reservoir
// replenishment is exercised when N is large enough that the refund
// from cycle i seeds the budget for cycle i+1. See
// test_state_gas_sstore.py::test_sstore_restoration_*.
func fillSstoreRestoration(gst *GstMaker, fork string) {
	target := common.HexToAddress("0xF1")

	p := program.New()
	cycles := 1 + rand.Intn(20)
	for i := 0; i < cycles; i++ {
		slot := i
		val := 1 + rand.Intn(100)
		p.Sstore(slot, val) // 0 → nonzero (charges 97,920 state gas)
		p.Sstore(slot, 0)   // nonzero → 0 (refunds 97,920 to reservoir)
	}
	p.Op(vm.STOP)

	gst.AddAccount(target, GenesisAccount{
		Code:    p.Bytes(),
		Balance: new(big.Int),
		Storage: make(map[common.Hash]common.Hash),
	})

	gasLimit := gasLimitOptions8037[rand.Intn(len(gasLimitOptions8037))]
	tx := &StTransaction{
		GasLimit:             []uint64{gasLimit},
		Nonce:                0,
		Value:                []string{"0x0"},
		Data:                 []string{"0x"},
		MaxFeePerGas:         big.NewInt(0x10),
		MaxPriorityFeePerGas: big.NewInt(0x10),
		To:                   target.Hex(),
		Sender:               sender,
		PrivateKey: hexutil.MustDecode(
			"0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8",
		),
	}
	gst.SetTx(tx)
}
