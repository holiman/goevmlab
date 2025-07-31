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
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"github.com/holiman/uint256"
	"golang.org/x/exp/rand"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
)

func fillSecp256R(gst *GstMaker, fork string) {
	// Add a contract which calls BLS
	dest := common.HexToAddress("0x00ca115ec9")
	gst.AddAccount(dest, GenesisAccount{
		Code:    randCallSecp256R(),
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

func randCallSecp256R() []byte {
	p := program.New()

	hash := make([]byte, 32)
	_, _ = rand.Read(hash)

	privKey, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)

	r, s, _ := ecdsa.Sign(crand.Reader, privKey, hash)
	data := append([]byte{}, hash...)

	data = append(data, uint256.MustFromBig(r).PaddedBytes(32)...)
	data = append(data, uint256.MustFromBig(s).PaddedBytes(32)...)

	data = append(data, uint256.MustFromBig(privKey.PublicKey.X).PaddedBytes(32)...)
	data = append(data, uint256.MustFromBig(privKey.PublicKey.Y).PaddedBytes(32)...)

	// Mutate it randomly a bit
	mutate(data)

	p.Mstore(data, 0)

	p.Call(nil, 0x100, 0, 0, len(data), 0, 64)

	// pop the ret value
	p.Op(vm.POP)
	// Store the output in some slot, to make sure the stateroot changes
	p.MemToStorage(0, 32, 0)
	return p.Bytes()
}
