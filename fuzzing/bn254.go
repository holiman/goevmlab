// Copyright 2020 Martin Holst Swende, Marius van der Wijden
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
	bn2562 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

type prec struct {
	addr    int
	newData func() []byte
	outsize int
}

var precompilesBn254 = []prec{
	// Takes two 64-byte points as inputs, output is one 64-byte point
	{0x6, newBnAdd, 64},
	// Takes one 64-byte point, and one 32-byte scalar as input. Output is 64-byte point
	{0x7, newBnScalarMul, 64},
	// Input is multiple of 192 (bn256.G1: 64 byte, bn256.G2: 128 byte). Output is
	// 32 bytes, boolean true or false.
	{0x8, newBnPairing, 32},
}

func fillBn254(gst *GstMaker, fork string) {
	// Add a contract which calls the Bn precompiles
	dest := common.HexToAddress("0x00ca110b15012381")
	code := RandCallBn()
	gst.AddAccount(dest, GenesisAccount{
		Code:    code,
		Balance: new(big.Int),
		Storage: make(map[common.Hash]common.Hash),
	})
	// The transaction
	gst.SetTx(&StTransaction{
		// 8M gaslimit
		GasLimit:   []uint64{16_000_000},
		Nonce:      0,
		Value:      []string{randHex(2)},
		Data:       []string{randHex(2)},
		GasPrice:   big.NewInt(0x10),
		To:         dest.Hex(),
		Sender:     sender,
		PrivateKey: pKey,
	})
}

func RandCallBn() []byte {
	p := program.New()
	offset := 0
	for _, precompile := range precompilesBn254 {
		data := precompile.newData()
		mutate(data) // don't always use valid data
		p.Mstore(data, 0)
		memInFn := func() (offset, size interface{}) {
			offset, size = 0, len(data)
			return
		}
		sizeOut := precompile.outsize
		memOutFn := func() (offset, size interface{}) {
			offset, size = 0, sizeOut
			return
		}
		addrGen := func() interface{} {
			return precompile.addr
		}
		p2 := RandCall(GasRandomizer(), addrGen, ValueRandomizer(), memInFn, memOutFn)
		p.Append(p2)
		// pop the ret value
		p.Op(vm.POP)
		// Store the output in some slot, to make sure the stateroot changes
		p.MemToStorage(0, sizeOut, offset)
		offset += sizeOut
	}
	return p.Bytes()
}

func newBnAdd() []byte {
	// Takes two 64-byte points as inputs
	a := makeBadBn254G1()
	b := makeBadBn254G1()
	return append(a, b...)
}

func newBnScalarMul() []byte {
	// Takes one 64-byte point, and one 32-byte scalar as input
	a := makeBadBn254G1()
	b := make([]byte, 32)
	_, _ = crand.Read(b)
	return append(a, b...)
}

func newBnPairing() []byte {
	// Input is multiple of 192 (bn256.G1: 64 byte, bn256.G2: 128 byte). Output is
	// 32 bytes, boolean true or false.
	k := 1 + randInt64()
	var res []byte
	for i := 0; i < int(k); i++ {
		a := makeBadBn254G1()
		res = append(res, a...)
		b := makeBadBn254G2()
		res = append(res, b...)
	}
	return res
}

func makeBadBn254G1() []byte {
	var retval []byte
	if c := rand.Intn(10); c == 0 {
		// Produces crappy G1s which are (usually not) on curve
		retval = make([]byte, 64)
		_, _ = crand.Read(retval)
	} else {
		_, g1, err := bn2562.RandomG1(crand.Reader)
		if err != nil {
			panic(err)
		}
		retval = g1.Marshal()
	}
	// Potentially mutate it a bit
	if rand.Intn(10) == 0 {
		retval[rand.Intn(len(retval))] = byte(rand.Int())
	}
	return retval
}

func makeBadBn254G2() []byte {
	var retval []byte
	if c := rand.Intn(10); c == 0 {
		// Produces crappy G2s which are (usually not) on curve
		retval = make([]byte, 128)
		_, _ = crand.Read(retval)
	} else {
		_, g2, err := bn2562.RandomG2(crand.Reader)
		if err != nil {
			panic(err)
		}
		retval = g2.Marshal()
	}
	// Potentially mutate it a bit
	if rand.Intn(10) == 0 {
		retval[rand.Intn(len(retval))] = byte(rand.Int())
	}
	return retval
}
