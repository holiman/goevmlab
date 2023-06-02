// Copyright 2022 Martin Holst Swende
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
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func fillEcRecover(gst *GstMaker, fork string) {
	// Add a contract which calls BLS
	dest := common.HexToAddress("0x00ca11ec5ec04e5")
	gst.AddAccount(dest, GenesisAccount{
		Code:    randCallECRecover(),
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

func randCallECRecover() []byte {
	p := program.NewProgram()
	offset := 0
	rounds := rand.Int31n(10000)
	for i := int32(0); i < rounds; i++ {
		data := make([]byte, 128)
		_, _ = crand.Read(data)
		p.Mstore(data, 0)
		memInFn := func() (offset, size interface{}) {
			offset, size = 0, 128
			return
		}
		// ecrecover outputs 32 bytes
		memOutFn := func() (offset, size interface{}) {
			offset, size = 0, 32
			return
		}
		addrGen := func() interface{} {
			return 7
		}
		gasRand := func() interface{} {
			return big.NewInt(rand.Int63n(100000))
		}
		p2 := RandCall(gasRand, addrGen, ValueRandomizer(), memInFn, memOutFn)
		p.AddAll(p2)
		// pop the ret value
		p.Op(ops.POP)
		// Store the output in some slot, to make sure the stateroot changes
		p.MemToStorage(0, 32, offset)
		offset += 32
	}
	return p.Bytecode()
}
