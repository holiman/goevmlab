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
	"crypto/sha256"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	gokzg4844 "github.com/crate-crypto/go-kzg-4844"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
)

func fillPointEvaluation4844(gst *GstMaker, fork string) {
	// Add a contract which calls the Bn precompiles
	dest := common.HexToAddress("0x00ca11004844")
	code := RandCallPointEval()
	gst.AddAccount(dest, GenesisAccount{
		Code:    code,
		Balance: big.NewInt(10_000_000),
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

func RandCallPointEval() []byte {
	p := program.New()
	offset := 0
	data := makeData()
	mutate(data) // don't always use valid data
	p.Mstore(data, 0)
	memInFn := func() (offset, size interface{}) {
		offset, size = 0, len(data)
		return
	}
	sizeOut := 64
	memOutFn := func() (offset, size interface{}) {
		offset, size = 0, sizeOut
		return
	}
	addrGen := func() interface{} { return []byte{0x0a} }
	p2 := RandCall(GasRandomizer(), addrGen, ValueRandomizer(), memInFn, memOutFn)
	p.Append(p2)
	// pop the ret value
	p.Op(vm.POP)
	// Store the output in some slot, to make sure the stateroot changes
	p.MemToStorage(0, sizeOut, offset)
	offset += sizeOut
	return p.Bytes()
}

// From https://eips.ethereum.org/EIPS/eip-4844:
//
//	The data is encoded as follows:
//	versioned_hash | z | y | commitment | proof |
//	with z and y being padded 32 byte big endian values
//
//	assert len(input) == 192
//	versioned_hash = input[:32]
//	z = input[32:64] (point)
//	y = input[64:96] (claim)
//	commitment = input[96:144]
//	proof = input[144:192]
func makeData() []byte {

	blob := randBlob()
	b2 := (*kzg4844.Blob)(blob)

	commitment, err := kzg4844.BlobToCommitment(b2)
	if err != nil {
		panic(err)
	}
	var (
		versionedHash   = kZGToVersionedHash(commitment)
		point           = randFieldElement()
		proof, claim, _ = kzg4844.ComputeProof(b2, kzg4844.Point(point))
	)
	var data []byte
	data = append(data, versionedHash[:]...)
	data = append(data, point[:]...)
	data = append(data, claim[:]...)
	data = append(data, commitment[:]...)
	data = append(data, proof[:]...)
	return data
}

// kZGToVersionedHash implements kzg_to_versioned_hash from EIP-4844
func kZGToVersionedHash(kzg kzg4844.Commitment) common.Hash {
	h := sha256.Sum256(kzg[:])
	h[0] = 0x01

	return h
}

func randFieldElement() gokzg4844.Scalar {
	var r fr.Element
	r.SetRandom()
	return gokzg4844.SerializeScalar(r)
}

func randBlob() *gokzg4844.Blob {
	var blob gokzg4844.Blob
	for i := 0; i < len(blob); i += gokzg4844.SerializedScalarSize {
		fieldElementBytes := randFieldElement()
		copy(blob[i:i+gokzg4844.SerializedScalarSize], fieldElementBytes[:])
	}
	return &blob
}
