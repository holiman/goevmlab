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

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

var modulo, _ = big.NewInt(0).SetString("0x1a0111ea397fe69a4b1ba7b6434bacd764774b84f38512bf6730d2a0f6b0f6241eabfffeb153ffffb9feffffffffaaab", 0)

type blsPrec struct {
	addr    int
	newData func() []byte
	outsize int
}

var precompilesBLS = []blsPrec{
	{0xb, newG1Add, 128},    // G1Add
	{0xc, newG1Mul, 128},    // G1Mul
	{0xd, newG1Exp, 128},    // G1MultiExp
	{0xe, newG2Add, 256},    // G2Add
	{0xf, newG2Mul, 256},    // G2Mul
	{0x10, newG2Exp, 256},   // G2MultiExp
	{0x11, newPairing, 32},  // Pairing
	{0x12, newFPtoG1, 128},  // FP to G1
	{0x13, newFP2toG2, 256}, // FP2 to G2
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

func RandCallBLS() []byte {
	p := program.NewProgram()
	offset := 0
	for _, precompile := range precompilesBLS {
		data := precompile.newData()
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
		p.AddAll(p2)
		// pop the ret value
		p.Op(ops.POP)
		// Store the output in some slot, to make sure the stateroot changes
		p.MemToStorage(0, sizeOut, offset)
		offset += sizeOut
	}
	return p.Bytecode()
}

func newG1Add() []byte {
	a := newG1Point()
	b := newG1Point()
	return append(a, b...)
}

func newG1Mul() []byte {
	a := newG1Point()
	mul := make([]byte, 32)
	_, _ = crand.Read(mul)
	return append(a, mul...)
}

func newG1Exp() []byte {
	i := randInt64()
	var res []byte
	for k := 0; k < int(i); k++ {
		input := newG1Mul()
		res = append(res, input...)
	}
	return res
}

func newG2Add() []byte {
	a := newG2Point()
	b := newG2Point()
	return append(a, b...)
}

func newG2Mul() []byte {
	a := newG2Point()
	mul := make([]byte, 32)
	_, _ = crand.Read(mul)
	return append(a, mul...)
}

func newG2Exp() []byte {
	i := randInt64()
	var res []byte
	for k := 0; k < int(i); k++ {
		input := newG2Mul()
		res = append(res, input...)
	}
	return res
}

func newFPtoG1() []byte {
	return newFieldElement()
}

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
func newPairing() []byte {
	_, _, _, genG2 := bls12381.Generators()
	pairs := randInt64()
	var res []byte
	target := new(big.Int)
	// LHS: sum(x: 1->n: e(aMulx * G1, bMulx * G2))
	for k := 0; k < int(pairs); k++ {
		aMul := randScalar()
		bMul := randScalar()
		g1 := new(bls12381.G1Affine).ScalarMultiplicationBase(aMul)
		g2 := new(bls12381.G2Affine).ScalarMultiplication(&genG2, bMul)
		res = append(res, g1.Marshal()...)
		res = append(res, g2.Marshal()...)
		// Add to s
		target = target.Add(target, aMul.Mul(aMul, bMul))
	}
	// RHS: e(G1, G2) ^ s
	ta := new(bls12381.G1Affine).ScalarMultiplicationBase(target)
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
	buf := make([]byte, 48)
	copy(buf[48-len(bytes):], bytes)
	return buf
}

// newG1Point generates a random G1 and returns it as a 96-byte
// byte slice (without point compression)
func newG1Point() []byte {
	s := randScalar()
	_, _, g1Gen, _ := bls12381.Generators()
	cp := new(bls12381.G1Affine)
	cp.ScalarMultiplication(&g1Gen, s)
	marshalled := cp.Marshal()
	return marshalled[:]
}

// newG2Point generates a random G2 and returns it as a 192-byte
// byte slice (without point compression)
func newG2Point() []byte {
	s := randScalar()
	_, _, _, g2gen := bls12381.Generators()
	cp := new(bls12381.G2Affine)
	cp.ScalarMultiplication(&g2gen, s)
	marshalled := cp.Marshal()
	return marshalled[:]
}
