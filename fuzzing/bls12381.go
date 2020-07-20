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
	"fmt"
	"io/ioutil"
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
	newData func(iv []byte, config MutationConfig) []byte
	outsize int
}

var precompilesBLS = []blsPrec{
	blsPrec{10, NewG1Add, 0}, // G1Add
	blsPrec{11, NewG1Mul, 0}, // G1Mul
	blsPrec{12, NewG1Exp, 0}, // G1Exp
	blsPrec{13, NewG2Add, 0}, // G2Add
	blsPrec{14, NewG2Mul, 0}, // G2Add
}

func GenerateBLS() *GstMaker {
	gst := basicStateTest("Yolo-Ephnet-1")
	// Add a contract which calls BLS
	dest := common.HexToAddress("0x00ca110b15012381")
	gst.AddAccount(dest, GenesisAccount{
		Code:    RandCallBLS(),
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
	return gst
}

func RandCallBLS() []byte {
	p := program.NewProgram()
	offset := 0
	var iv []byte
	config := new(MutationConfig)
	for _, precompile := range precompilesBLS {
		data := precompile.newData(iv, *config)
		sizeIn := len(data)
		if sizeIn == 160 || sizeIn == 288 || sizeIn == 384 {
			sizeIn = sizeIn * rand.Intn(1000)
		}
		p.Mstore(data, 0)
		memInFn := func() (offset, size interface{}) {
			offset, size = 0, sizeIn
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

func randomBLSArgs(inputSize int) []byte {
	if len(BLSCorpus) != 0 {
		for i := 0; i < 100; i++ {
			index := rand.Intn(len(BLSCorpus))
			if len(BLSCorpus[index]) == inputSize {
				return BLSCorpus[index]
			}
		}
		fmt.Println("No suitable corpus element found, falling back to random")
	}
	data := make([]byte, inputSize)
	rand.Read(data)
	return data
}

var BLSCorpus [][]byte

func ReadBLSCorpus() error {
	dir, err := ioutil.ReadDir("corpus")
	if err != nil {
		return err
	}
	BLSCorpus = make([][]byte, len(dir))
	for i, info := range dir {
		name := info.Name()
		input, err := ioutil.ReadFile("corpus/" + name)
		if err != nil {
			return err
		}
		BLSCorpus[i] = input
	}
	return nil
}

func NewG1Add(iv []byte, config MutationConfig) []byte {
	a := NewG1Point(iv, config)
	b := NewG1Point(iv, config)
	return append(a, b...)
}

func NewG1Mul(iv []byte, config MutationConfig) []byte {
	a := NewG1Point(iv, config)
	mul := make([]byte, 32)
	rand.Read(mul)
	return append(a, mul...)
}

func NewG1Exp(iv []byte, config MutationConfig) []byte {
	i := rand.Int31n(140)
	var res []byte
	for k := 0; k < int(i); k++ {
		input := NewG1Mul(iv, config)
		res = append(res, input...)
	}
	return res
}

func NewG2Add(iv []byte, config MutationConfig) []byte {
	a := NewG2Point(iv, config)
	b := NewG2Point(iv, config)
	return append(a, b...)
}

func NewG2Mul(iv []byte, config MutationConfig) []byte {
	a := NewG2Point(iv, config)
	mul := make([]byte, 32)
	rand.Read(mul)
	return append(a, mul...)
}

// sanitizeMutate maps arbitrary input to a valid field element on the curve
func sanitizeMutate(data []byte) []byte {
	fe := big.NewInt(0).SetBytes(data)
	if fe.Cmp(modulo) == 1 {
		fe = fe.Mod(fe, modulo)
	}
	buf := make([]byte, 48)
	copy(buf[48-len(fe.Bytes()):], fe.Bytes())
	return buf
}

func NewFieldElement(iv []byte, config MutationConfig) []byte {
	re := Mutate(iv, config)
	return sanitizeMutate(re)
}

func NewG1Point(iv []byte, config MutationConfig) []byte {
	a := NewFieldElement(iv, config)
	// Initialize G1
	g := bls12381.NewG1()
	b, err := g.MapToCurve(a)
	if err != nil {
		panic(err)
	}
	return g.EncodePoint(b)
}

func NewG2Point(iv []byte, config MutationConfig) []byte {
	a := NewFieldElement(iv, config)
	b := NewFieldElement(a, config)
	x := append(a, b...)
	// Initialize G2
	g := bls12381.NewG2()
	// Compute mapping
	res, err := g.MapToCurve(x)
	if err != nil {
		panic(err)
	}
	return g.EncodePoint(res)
}
