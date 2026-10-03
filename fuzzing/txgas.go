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
	crand "crypto/rand"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
)

// fillTxGas fuzzes the transaction-level gas rules: intrinsic gas, the
// calldata floor and access list costs (EIP-2028, EIP-2930, EIP-7623 and, in
// Amsterdam, EIP-2780, EIP-7976 and EIP-7981). The transaction is always
// valid; the destination stores the gas it has left and the sender pays the
// floor, so a client charging differently ends up with a different root.
func fillTxGas(gst *GstMaker, fork string) {
	dest := common.HexToAddress("0x000000000000000000000000000000000000ac1d")
	slots := make(map[common.Hash]common.Hash)
	for i := range 8 {
		slots[common.BigToHash(big.NewInt(int64(i)))] = common.BigToHash(big.NewInt(int64(i + 1)))
	}
	balanceOf := []common.Address{sender, common.HexToAddress("0xc0ffee"), common.HexToAddress("0x01")}
	// Read the slots and balances an access list may warm, then store the gas
	// left and the calldata size, so any difference in the charges shows in the
	// state root.
	p := program.New()
	for i := range 8 {
		p.Push(i).Op(vm.SLOAD).Op(vm.POP)
	}
	for _, a := range balanceOf {
		p.Push(a).Op(vm.BALANCE).Op(vm.POP)
	}
	p.Op(vm.GAS).Push(0x100).Op(vm.SSTORE)
	p.Op(vm.CALLDATASIZE).Push(0x101).Op(vm.SSTORE)
	gst.AddAccount(dest, GenesisAccount{
		Code:    p.Bytes(),
		Balance: big.NewInt(10_000_000),
		Storage: slots,
	})

	data := randCalldata()
	accessList := randAccessList(append([]common.Address{dest, sender}, balanceOf...))
	to := dest
	value := "0x00"
	if rand.Intn(5) == 0 {
		// A value transfer to an account that doesn't exist.
		to = common.BytesToAddress(randBytes(20))
	}
	if rand.Intn(2) == 0 {
		value = randHex(4)
	}

	gst.SetTx(&StTransaction{
		GasLimit:    []uint64{randTxGasLimit(len(data), accessList)},
		Value:       []string{value},
		Data:        []string{hexutil.Encode(data)},
		AccessLists: []*types.AccessList{&accessList},
		GasPrice:    big.NewInt(0x10),
		To:          to.Hex(),
		Sender:      sender,
		PrivateKey:  pKey,
	})
}

// randCalldata returns calldata of a random size, mostly zero bytes, mostly
// non-zero bytes or a mix, so that either the standard calldata cost or the
// calldata floor can be the larger one.
func randCalldata() []byte {
	var size int
	switch rand.Intn(4) {
	case 0:
		size = 0
	case 1:
		size = rand.Intn(64)
	case 2:
		size = rand.Intn(1024)
	default:
		size = rand.Intn(8192)
	}
	data := randBytes(size)
	switch rand.Intn(3) {
	case 0: // mostly zero
		for i := range data {
			if rand.Intn(8) != 0 {
				data[i] = 0
			}
		}
	case 1: // no zero bytes
		for i := range data {
			if data[i] == 0 {
				data[i] = 1
			}
		}
	}
	return data
}

// randAccessList returns up to six addresses with up to four storage keys
// each, drawn from the given addresses, precompiles and random addresses, and
// from the slots the destination reads and random slots.
func randAccessList(addrs []common.Address) types.AccessList {
	var al types.AccessList
	for range rand.Intn(7) {
		var addr common.Address
		switch rand.Intn(3) {
		case 0:
			addr = addrs[rand.Intn(len(addrs))]
		case 1:
			addr = common.BigToAddress(big.NewInt(int64(1 + rand.Intn(0x11))))
		default:
			addr = common.BytesToAddress(randBytes(20))
		}
		keys := []common.Hash{}
		for range rand.Intn(5) {
			if rand.Intn(2) == 0 {
				keys = append(keys, common.BigToHash(big.NewInt(int64(rand.Intn(8)))))
			} else {
				keys = append(keys, common.BytesToHash(randBytes(32)))
			}
		}
		al = append(al, types.AccessTuple{Address: addr, StorageKeys: keys})
	}
	return al
}

// randTxGasLimit returns a gas limit above the intrinsic cost, the calldata
// floor and the execution cost on every fork, with some random slack.
func randTxGasLimit(dataLen int, al types.AccessList) uint64 {
	if rand.Intn(4) == 0 {
		return 8_000_000
	}
	limit := uint64(21000+64*dataLen+400_000) + uint64(len(al))*6000 + uint64(al.StorageKeys())*4000
	return limit + uint64(rand.Int63n(int64(limit/5)))
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = crand.Read(b)
	return b
}
