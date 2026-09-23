// Copyright 2026 Martin Holst Swende
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

package fuzzing

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/holiman/goevmlab/ops"
)

// This file contains block-level fillers: transaction kinds which are not
// interesting for state tests, but which produce block access list entries
// (EIP-7928) of a kind that the state test fillers do not: balance-only
// changes, nonce changes, code changes and withdrawals.

// BtOptions configures the blocktest factory.
type BtOptions struct {
	// Fillers are the names of the state test fillers to draw transactions from.
	Fillers []string
	// Blocks is the number of blocks per test.
	Blocks int
	// TxsPerBlock is the number of transactions per block.
	TxsPerBlock int
}

// BtFactory returns a function which generates blocktest makers for the given
// fork, drawing transactions from the given state test fillers, interleaved
// with block-level transaction kinds.
func BtFactory(fork string, opts BtOptions) (func() (*BtMaker, error), error) {
	if opts.Blocks <= 0 {
		opts.Blocks = 2
	}
	if opts.TxsPerBlock <= 0 {
		opts.TxsPerBlock = 3
	}
	var fills []func(*GstMaker, string)
	for _, name := range opts.Fillers {
		filler, ok := fillers[name]
		if !ok {
			return nil, fmt.Errorf("unknown filler %q", name)
		}
		fills = append(fills, filler)
	}
	if len(fills) == 0 {
		return nil, fmt.Errorf("no fillers")
	}
	return func() (*BtMaker, error) {
		bt, err := NewBtMaker(fork)
		if err != nil {
			return nil, err
		}
		for i := 0; i < opts.Blocks; i++ {
			bt.AddBlock()
			for j := 0; j < opts.TxsPerBlock; j++ {
				switch rand.Intn(8) {
				case 0:
					addValueTransfer(bt)
				case 1:
					addCreateTx(bt, fork)
				default:
					gst := BasicStateTest(fork)
					fills[rand.Intn(len(fills))](gst, fork)
					bt.AddTxFromGst(gst)
				}
			}
			if rand.Intn(2) == 0 {
				addRandomWithdrawals(bt)
			}
		}
		return bt, nil
	}, nil
}

// btTargets are addresses which the block-level fillers use as recipients.
// Some of them exist in the state (via the state test fillers), some do not.
var btTargets = []common.Address{
	common.HexToAddress("0xF1"),
	common.HexToAddress("0xF2"),
	common.HexToAddress("0x01"),
	common.HexToAddress("0xb94f5374fce5edbc8e2a8697c15331677e6ebf0b"), // coinbase
	common.HexToAddress("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
	common.HexToAddress("0xcccccccccccccccccccccccccccccccccccccccc"),
	common.HexToAddress("0xdddddddddddddddddddddddddddddddddddddddd"),
}

func randomTarget() common.Address {
	return btTargets[rand.Intn(len(btTargets))]
}

// addValueTransfer adds a plain value transfer, optionally with calldata.
func addValueTransfer(bt *BtMaker) {
	dest := randomTarget()
	bt.AddTx(&StTransaction{
		To:       dest.Hex(),
		GasLimit: []uint64{100_000},
		Value:    []string{randHex(4)},
		Data:     []string{randHex(1 + rand.Intn(20))},
		GasPrice: big.NewInt(0x10),
		Sender:   sender,
	})
}

// addCreateTx adds a contract creation, whose init code writes storage,
// deploys random code, and optionally creates further contracts.
func addCreateTx(bt *BtMaker, fork string) {
	forkDef := ops.LookupFork(fork)
	if forkDef == nil {
		panic("bad fork")
	}
	initcode := program.New()
	// Write a few storage slots
	for i := 0; i < rand.Intn(4); i++ {
		initcode.Sstore(rand.Intn(8), rand.Intn(100))
	}
	// Maybe create a nested contract with random runtime code
	if rand.Intn(2) == 0 {
		nested := program.New().ReturnData(randomBytecode(forkDef))
		initcode.Create2(nested.Bytes(), rand.Intn(10))
	}
	// Return some random runtime code
	code := randomBytecode(forkDef)
	if len(code) > 0 && code[0] == 0xEF {
		code[0] = 0xEE
	}
	initcode.ReturnData(code)
	bt.AddTx(&StTransaction{
		To:       "",
		GasLimit: []uint64{2_000_000},
		Value:    []string{"0x01"},
		Data:     []string{fmt.Sprintf("0x%x", initcode.Bytes())},
		GasPrice: big.NewInt(0x10),
		Sender:   sender,
	})
}

// addRandomWithdrawals adds a few withdrawals to the current block.
func addRandomWithdrawals(bt *BtMaker) {
	for i := 0; i < 1+rand.Intn(3); i++ {
		bt.AddWithdrawal(uint64(rand.Intn(100)), randomTarget(), uint64(rand.Intn(1000)))
	}
}
