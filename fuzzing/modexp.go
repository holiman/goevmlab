// Copyright 2025 Martin Holst Swende
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
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/holiman/uint256"
)

func fillModexp(gst *GstMaker, fork string) {
	// Add a contract which calls BLS
	dest := common.HexToAddress("0x00ca1130de4f")
	gst.AddAccount(dest, GenesisAccount{
		Code:    randCallModexp(),
		Balance: big.NewInt(10_000_000),
		Storage: make(map[common.Hash]common.Hash),
	})
	// The transaction
	gst.SetTx(&StTransaction{
		GasLimit:   []uint64{16_000_000},
		Nonce:      0,
		Value:      []string{randHex(4)},
		Data:       []string{randHex(100)},
		GasPrice:   big.NewInt(0x10),
		To:         dest.Hex(),
		Sender:     sender,
		PrivateKey: pKey,
	})
}

func randModexpInt() *big.Int {
	b := make([]byte, 4)
	_, _ = crand.Read(b)
	// 32/256 chance of zero
	if b[0] < 32 {
		return big.NewInt(0)
	}
	// 128/256 chance of a uint64
	if b[1] < 128 {
		return big.NewInt(0).SetUint64(rand.Uint64())
	}
	// Random size, up to 2048 in size
	size := rand.Intn(2846)
	val := make([]byte, size)
	_, _ = crand.Read(val)
	return (new(big.Int)).SetBytes(val)
}

func randCallModexp() []byte {
	p := program.New()

	base := randModexpInt()
	exp := randModexpInt()
	mod := randModexpInt()

	// 32 bytes each for baselen, expLen and modlen
	buf := make([]byte, 0)

	buf = append(buf, uint256.NewInt(uint64((base.BitLen()+7)/8)).PaddedBytes(32)...)
	buf = append(buf, uint256.NewInt(uint64((exp.BitLen()+7)/8)).PaddedBytes(32)...)
	buf = append(buf, uint256.NewInt(uint64((mod.BitLen()+7)/8)).PaddedBytes(32)...)

	buf = append(buf, base.Bytes()...)
	buf = append(buf, exp.Bytes()...)
	buf = append(buf, mod.Bytes()...)

	// Now mutate it randomly a bit
	mutate(buf)

	p.Mstore(buf, 0)

	p.Call(nil, 0x5, 0, 0, len(buf), 0, 64)

	// pop the ret value
	p.Op(vm.POP)
	// Store the output in some slot, to make sure the stateroot changes
	p.MemToStorage(0, 32, 0)
	return p.Bytes()
}
