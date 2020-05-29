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
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

var precompilesBLS = []int{
	10, 11, 12, 13, 14, 15, 16, 17, 18,
}

var inputSizes = []int{
	256, 256, 160, 512, 512, 288, 384, 64, 128,
}

var ouputSizes = []int{
	128, 128, 128, 256, 256, 256, 32, 128, 256,
}

func GenerateBLS() *GstMaker {
	gst := basicStateTest("Yolo-Ephnet-1")
	// Add a contract which calls BLS
	dest := common.HexToAddress("0x0000ca1100b1a7e")
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
	precompile := rand.Int31n(int32(len(precompilesBLS)))
	sizeIn := inputSizes[precompile]
	if sizeIn == 160 || sizeIn == 288 || sizeIn == 384 {
		sizeIn = sizeIn * rand.Intn(1000)
	}
	data := randomBLSArgs(sizeIn)
	p.Mstore(data, 0)
	memInFn := func() (offset, size interface{}) {
		offset, size = 0, sizeIn
		return
	}
	sizeOut := ouputSizes[precompile]
	memOutFn := func() (offset, size interface{}) {
		offset, size = 0, sizeOut
		return
	}
	addrGen := func() interface{} {
		return precompilesBLS[precompile]
	}
	p2 := RandCall(GasRandomizer(), addrGen, ValueRandomizer(), memInFn, memOutFn)
	p.AddAll(p2)
	// pop the ret value
	p.Op(ops.POP)
	// Store the output in some slot, to make sure the stateroot changes
	p.MemToStorage(0, sizeOut, 0)
	return p.Bytecode()
}

func randomBLSArgs(inputSize int) []byte {
	data := make([]byte, inputSize)
	rand.Read(data)
	return data
}
