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
	crypto "crypto/rand"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto/bls12381"
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
	{10, NewG1Add, 128},   // G1Add
	{11, NewG1Mul, 128},   // G1Mul
	{12, NewG1Exp, 128},   // G1MultiExp
	{13, NewG2Add, 256},   // G2Add
	{14, NewG2Mul, 256},   // G2Mul
	{15, NewG2Exp, 256},   // G2MultiExp
	{16, NewPairing, 32},  // Pairing
	{17, NewFPtoG1, 128},  // FP to G1
	{17, NewFP2toG2, 256}, // FP2 to G2
}

func GenerateBLS() (*GstMaker, []byte) {
	gst := basicStateTest("Yolo-Ephnet-1")
	// Add a contract which calls BLS
	dest := common.HexToAddress("0x00ca110b15012381")
	code := RandCallBLS()
	gst.AddAccount(dest, GenesisAccount{
		Code:    code,
		Balance: new(big.Int),
		Storage: make(map[common.Hash]common.Hash),
	})
	// The transaction
	{
		tx := &stTransaction{
			// 8M gaslimit
			GasLimit:   []uint64{8000000},
			Nonce:      0,
			Value:      []string{randHex(4)},
			Data:       []string{randHex(100)},
			GasPrice:   big.NewInt(0x01),
			To:         dest.Hex(),
			PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		}
		gst.SetTx(tx)
	}
	return gst, code
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

func NewG1Add() []byte {
	a := NewG1Point()
	b := NewG1Point()
	return append(a, b...)
}

func NewG1Mul() []byte {
	a := NewG1Point()
	mul := make([]byte, 32)
	rand.Read(mul)
	return append(a, mul...)
}

func NewG1Exp() []byte {
	i := rand.Int31n(140)
	var res []byte
	for k := 0; k < int(i); k++ {
		input := NewG1Mul()
		res = append(res, input...)
	}
	return res
}

func NewG2Add() []byte {
	a := NewG2Point()
	b := NewG2Point()
	return append(a, b...)
}

func NewG2Mul() []byte {
	a := NewG2Point()
	mul := make([]byte, 32)
	rand.Read(mul)
	return append(a, mul...)
}

func NewG2Exp() []byte {
	i := rand.Int31n(140)
	var res []byte
	for k := 0; k < int(i); k++ {
		input := NewG2Mul()
		res = append(res, input...)
	}
	return res
}

func NewFPtoG1() []byte {
	return NewFieldElement()
}

func NewFP2toG2() []byte {
	a := NewFieldElement()
	b := NewFieldElement()
	return append(a, b...)
}

var (
	// Initialize G1
	g1 = bls12381.NewG1()
	// Initialize G2
	g2 = bls12381.NewG2()
	// Initialize pairing engine
	bls = bls12381.NewPairingEngine()
	// Initialize rand reader
	reader = rand.New(rand.NewSource(1234))
)

// NewPairing creates a new valid pairing.
// We create the following pairing:
// e(aMul1 * G1, bMul1 * G2) * e(aMul2 * G1, bMul2 * G2) * ... * e(aMuln * G1, bMuln * G2) == e(G1, G2) ^ s
// with s = sum(x: 1 -> n: (aMulx * bMulx))
func NewPairing() []byte {
	pairs := rand.Int31n(150)
	var res []byte
	target := new(big.Int)
	// LHS: sum(x: 1->n: e(aMulx * G1, bMulx * G2))
	for k := 0; k < int(pairs); k++ {
		a, b := g1.One(), g2.One()
		aMul := new(big.Int).SetBytes(NewFieldElement())
		bMul := new(big.Int).SetBytes(NewFieldElement())
		a = g1.MulScalar(a, a, aMul)
		b = g2.MulScalar(b, b, bMul)
		res = append(res, g1.EncodePoint(a)...)
		res = append(res, g2.EncodePoint(b)...)
		// Add to s
		target = target.Add(target, aMul.Mul(aMul, bMul))
	}
	// RHS: e(G1, G2) ^ s
	ta, tb := g1.One(), g2.One()
	g1.MulScalar(ta, ta, target)
	g1.Neg(ta, ta)
	res = append(res, g1.EncodePoint(ta)...)
	res = append(res, g2.EncodePoint(tb)...)
	return res
}

func NewFieldElement() []byte {
	ret, err := crypto.Int(reader, modulo)
	if err != nil {
		panic(err)
	}
	bytes := ret.Bytes()
	buf := make([]byte, 48)
	copy(buf[48-len(bytes):], bytes)
	return buf
}

func NewG1Point() []byte {
	a := NewFieldElement()
	b, err := g1.MapToCurve(a)
	if err != nil {
		panic(err)
	}
	return g1.EncodePoint(b)
}

func NewG2Point() []byte {
	a := NewFieldElement()
	b := NewFieldElement()
	x := append(a, b...)
	// Compute mapping
	res, err := g2.MapToCurve(x)
	if err != nil {
		panic(err)
	}
	return g2.EncodePoint(res)
}
