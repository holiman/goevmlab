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

	gnark "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
)

var modulo, _ = big.NewInt(0).SetString("1a0111ea397fe69a4b1ba7b6434bacd764774b84f38512bf6730d2a0f6b0f6241eabfffeb153ffffb9feffffffffaaab", 16)

type blsPrec struct {
	addr    int
	newData func() []byte
	outsize int
}

var precompilesBLS = []blsPrec{
	{0xb, newG1Add, 128},    // G1Add
	{0xc, newG1MSM, 128},    // G1Mul
	{0xd, newG2Add, 256},    // G2Add
	{0xe, newG2MSM, 256},    // G2Mul
	{0x0f, newPairing, 32},  // Pairing
	{0x10, newFPtoG1, 128},  // FP to G1
	{0x11, newFP2toG2, 256}, // FP2 to G2
}

func fillBls(gst *GstMaker, fork string) {
	// Add a contract which calls BLS
	dest := common.HexToAddress("0x00ca110b15012381")
	code := RandCallBLS()
	gst.AddAccount(dest, GenesisAccount{
		Code:    code,
		Balance: new(big.Int),
		Storage: make(map[common.Hash]common.Hash),
	})
	// The transaction
	gst.SetTx(&StTransaction{
		// 8M gaslimit
		GasLimit:   []uint64{8000000},
		Nonce:      0,
		Value:      []string{randHex(4)},
		Data:       []string{randHex(100)},
		GasPrice:   big.NewInt(0x10),
		To:         dest.Hex(),
		Sender:     sender,
		PrivateKey: pKey,
	})
}

// mutate does some bit-twiddling.
func mutate(data []byte) {
	if len(data) == 0 {
		return
	}
	for rand.Intn(2) == 0 {
		bit := rand.Intn(len(data) * 8) // // 13
		data[bit/8] = data[bit/8] ^ (1 << bit % 8)
	}
}

func RandCallBLS() []byte {
	p := program.New()
	offset := 0
	for _, precompile := range precompilesBLS {
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

func newG1Add() []byte {
	a := newG1Point()
	b := newG1Point()
	return append(a, b...)
}

func newG1MSM() []byte {
	a := newG1Point()
	mul := make([]byte, 32)
	_, _ = crand.Read(mul)
	return append(a, mul...)
}

func newG2Add() []byte {
	a := newG2Point()
	b := newG2Point()
	return append(a, b...)
}

func newG2MSM() []byte {
	a := newG2Point()
	mul := make([]byte, 32)
	_, _ = crand.Read(mul)
	return append(a, mul...)
}

// TODO This should return 64 bytes
// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2537.md#abi-for-mapping-fp-element-to-g1-point
func newFPtoG1() []byte {
	return newFieldElement()
}

// TODO This should return 128 bytes
// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2537.md#abi-for-mapping-fp2-element-to-g2-point
func newFP2toG2() []byte {
	a := newFieldElement()
	b := newFieldElement()
	return append(a, b...)
}

// randInt64 returns a new random int64
// With 3% probability it outputs 0
// With 92% probability it outputs a number [0..30)
// With 5% probability it outputs a number [0..150)
func randInt64() int64 {
	b := rand.Int31n(100)
	// Zero or not?
	if b < 3 {
		return 0
	}
	if b < 95 {
		return rand.Int63n(30)
	}
	return rand.Int63n(150)
}

// newPairing creates a new valid pairing.
// We create the following pairing:
// e(aMul1 * G1, bMul1 * G2) * e(aMul2 * G1, bMul2 * G2) * ... * e(aMuln * G1, bMuln * G2) == e(G1, G2) ^ s
// with s = sum(x: 1 -> n: (aMulx * bMulx))
// TODO!
// Check that it returns k * 384 bytes!
// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-2537.md#abi-for-pairing-check
func newPairing() []byte {
	_, _, _, genG2 := gnark.Generators()
	pairs := randInt64()
	var res []byte
	target := new(big.Int)
	// LHS: sum(x: 1->n: e(aMulx * G1, bMulx * G2))
	for k := 0; k < int(pairs); k++ {
		aMul := randScalar()
		bMul := randScalar()
		g1 := new(gnark.G1Affine).ScalarMultiplicationBase(aMul)
		g2 := new(gnark.G2Affine).ScalarMultiplication(&genG2, bMul)
		res = append(res, g1.Marshal()...)
		res = append(res, g2.Marshal()...)
		// Add to s
		target = target.Add(target, aMul.Mul(aMul, bMul))
	}
	// RHS: e(G1, G2) ^ s
	ta := new(gnark.G1Affine).ScalarMultiplicationBase(target)
	ta.Neg(ta)
	res = append(res, ta.Marshal()...)
	res = append(res, genG2.Marshal()...)
	return res
}

func randScalar() *big.Int {
	ret, err := crand.Int(crand.Reader, modulo)
	if err != nil {
		panic(err)
	}
	return ret
}

func newFieldElement() []byte {
	bytes := randScalar().Bytes()
	buf := make([]byte, 64)
	copy(buf[16+48-len(bytes):], bytes)
	return buf
}

// newG1Point generates a random G1 and returns it as a 128-byte slice.
func newG1Point() []byte {
	// sample a random scalar
	s := randScalar()
	// compute a random point
	cp := new(gnark.G1Affine)
	_, _, g1Gen, _ := gnark.Generators()
	cp.ScalarMultiplication(&g1Gen, s)
	return encodePointG1(cp)
}

// encodePointG1 encodes a point into 128 bytes.
func encodePointG1(p *gnark.G1Affine) []byte {
	out := make([]byte, 128)
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(out[16:]), p.X)
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(out[64+16:]), p.Y)
	return out
}

// newG2Point generates a random G2 and returns it as a 256-byte byte slice.
func newG2Point() []byte {
	s := randScalar()
	_, _, _, g2gen := gnark.Generators()
	cp := new(gnark.G2Affine)
	cp.ScalarMultiplication(&g2gen, s)
	return encodePointG2(cp)
}

// encodePointG2 encodes a point into 256 bytes.
func encodePointG2(p *gnark.G2Affine) []byte {
	out := make([]byte, 256)
	// encode x
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(out[16:16+48]), p.X.A0)
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(out[80:80+48]), p.X.A1)
	// encode y
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(out[144:144+48]), p.Y.A0)
	fp.BigEndian.PutElement((*[fp.Bytes]byte)(out[208:208+48]), p.Y.A1)
	return out
}
