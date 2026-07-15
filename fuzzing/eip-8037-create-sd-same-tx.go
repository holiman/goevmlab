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
	program2 "github.com/holiman/goevmlab/program"
)

// fillCreateSdSameTx generates a state test where a factory contract
// CREATEs (or CREATE2s) a contract whose initcode immediately calls
// SELFDESTRUCT. Tests the EIP-6780 × EIP-8037 same-tx CREATE+SD path:
// state gas is charged for the new account but no refund is issued
// because the account is removed in-tx and never persists. See
// test_state_gas_selfdestruct.py::test_create_selfdestruct_no_refund_*.
//
// Beneficiary is varied across four scenarios that interact with the
// SELFDESTRUCT state-gas table:
//   - ZERO address
//   - factory (self)
//   - pre-existing EOA (no new-account state gas)
//   - nonexistent (charges 183,600 if SD originator has nonzero balance)
func fillCreateSdSameTx(gst *GstMaker, fork string) {
	factory := common.HexToAddress("0xF1")

	var beneficiary common.Address
	switch rand.Intn(4) {
	case 0:
		beneficiary = common.Address{}
	case 1:
		beneficiary = factory
	case 2:
		beneficiary = common.HexToAddress("0xBE")
	default:
		beneficiary = common.HexToAddress("0xDEAD")
	}

	// Init code: push beneficiary, SELFDESTRUCT.
	ctor := program.New()
	ctor.Push(beneficiary)
	ctor.Op(vm.SELFDESTRUCT)

	// Factory: CREATE/CREATE2 the SD-only contract, then no-op call to
	// the resulting (now-destroyed) address. Coin-flip CREATE vs CREATE2.
	p := program.New()
	useCreate2 := rand.Intn(2) == 0
	program2.CreateAndCall(p, ctor.Bytes(), useCreate2, vm.CALL)
	p.Op(vm.STOP)

	gst.AddAccount(factory, GenesisAccount{
		Code:    p.Bytes(),
		Balance: big.NewInt(1), // SD with nonzero balance triggers new-account state gas
		Storage: make(map[common.Hash]common.Hash),
	})
	// Pre-existing beneficiary (the "0xBE" branch) gets a genesis entry
	// half the time, so the SD-to-existing-vs-new branches both surface.
	// Balance is nonzero so the account isn't "empty" (EELS rejects
	// empty accounts in pre-state fixtures).
	if beneficiary == common.HexToAddress("0xBE") && rand.Intn(2) == 0 {
		gst.AddAccount(beneficiary, GenesisAccount{
			Balance: big.NewInt(1),
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
		To:                   factory.Hex(),
		Sender:               sender,
		PrivateKey: hexutil.MustDecode(
			"0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8",
		),
	}
	gst.SetTx(tx)
}
