// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package fuzzing

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
)

// GenesisAlloc specifies the initial state that is part of the genesis block.
type GenesisAlloc map[common.Address]GenesisAccount

func (ga *GenesisAlloc) UnmarshalJSON(data []byte) error {
	m := make(map[common.UnprefixedAddress]GenesisAccount)
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*ga = make(GenesisAlloc)
	for addr, a := range m {
		(*ga)[common.Address(addr)] = a
	}
	return nil
}

//go:generate gencodec -type GenesisAccount -field-override genesisAccountMarshaling -out gen_genesis_account.go

// GenesisAccount is an account in the state of the genesis block.
// Copied from go-ethereum, with the mod of making Storage mandatory
type GenesisAccount struct {
	Code []byte `json:"code"`
	// N.B: parity demands storage even if it's empty
	Storage    map[common.Hash]common.Hash `json:"storage"`
	Balance    *big.Int                    `json:"balance" gencodec:"required"`
	Nonce      uint64                      `json:"nonce"`
	PrivateKey []byte                      `json:"secretKey,omitempty"` // for tests
}

type genesisAccountMarshaling struct {
	Code       hexutil.Bytes
	Balance    *math.HexOrDecimal256
	Nonce      math.HexOrDecimal64
	Storage    map[storageJSON]storageJSON
	PrivateKey hexutil.Bytes
}

// storageJSON represents a 256 bit byte array, but allows less than 256 bits when
// unmarshaling from hex.
type storageJSON common.Hash

func (h *storageJSON) UnmarshalText(text []byte) error {
	text = bytes.TrimPrefix(text, []byte("0x"))
	if len(text) > 64 {
		return fmt.Errorf("too many hex characters in storage key/value %q", text)
	}
	offset := len(h) - len(text)/2 // pad on the left
	if _, err := hex.Decode(h[offset:], text); err != nil {
		fmt.Println(err)
		return fmt.Errorf("invalid hex storage key/value %q", text)
	}
	return nil
}

func (h storageJSON) MarshalText() ([]byte, error) {
	return hexutil.Bytes(h[:]).MarshalText()
}

type GeneralStateTest map[string]*stJSON

// StateTest checks transaction processing without block context.
// See https://github.com/ethereum/EIPs/issues/176 for the test format specification.
type StateTest struct {
	json stJSON
}

// StateSubtest selects a specific configuration of a General State Test.
type StateSubtest struct {
	Fork  string
	Index int
}

func (t *StateTest) UnmarshalJSON(in []byte) error {
	return json.Unmarshal(in, &t.json)
}

type stJSON struct {
	Env  stEnv                    `json:"env"`
	Pre  GenesisAlloc             `json:"pre"`
	Tx   StTransaction            `json:"transaction"`
	Out  hexutil.Bytes            `json:"out"`
	Post map[string][]stPostState `json:"post"`
}

type stPostState struct {
	Root    common.Hash `json:"hash"`
	Logs    common.Hash `json:"logs"`
	Indexes stIndex     `json:"indexes"`
}

type stIndex struct {
	Data  int `json:"data"`
	Gas   int `json:"gas"`
	Value int `json:"value"`
}

//go:generate gencodec -type stEnv -field-override stEnvMarshaling -out gen_stenv.go

type stEnv struct {
	Coinbase     common.Address `json:"currentCoinbase"   gencodec:"required"`
	Difficulty   *big.Int       `json:"currentDifficulty" gencodec:"optional"`
	Random       *common.Hash   `json:"currentRandom,omitempty"     gencodec:"optional"`
	GasLimit     uint64         `json:"currentGasLimit"   gencodec:"required"`
	Number       uint64         `json:"currentNumber"     gencodec:"required"`
	Timestamp    uint64         `json:"currentTimestamp"  gencodec:"required"`
	PreviousHash common.Hash    `json:"previousHash"`
	BaseFee      *big.Int       `json:"currentBaseFee"`
}

type stEnvMarshaling struct {
	Coinbase   common.UnprefixedAddress
	Difficulty *math.HexOrDecimal256
	Random     *common.Hash
	GasLimit   math.HexOrDecimal64
	Number     math.HexOrDecimal64
	Timestamp  math.HexOrDecimal64
	BaseFee    *math.HexOrDecimal256
}

//go:generate gencodec -type StTransaction -field-override stTransactionMarshaling -out gen_sttransaction.go

type StTransaction struct {
	GasPrice             *big.Int            `json:"gasPrice"`
	MaxFeePerGas         *big.Int            `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas *big.Int            `json:"maxPriorityFeePerGas,omitempty"`
	Nonce                uint64              `json:"nonce"`
	To                   string              `json:"to"`
	Data                 []string            `json:"data"`
	AccessLists          []*types.AccessList `json:"accessLists,omitempty"`
	GasLimit             []uint64            `json:"gasLimit"`
	Value                []string            `json:"value"`
	PrivateKey           []byte              `json:"secretKey"`
	Sender               common.Address      `json:"sender"`
	BlobVersionedHashes  []common.Hash       `json:"blobVersionedHashes,omitempty"`
	BlobGasFeeCap        *big.Int            `json:"maxFeePerBlobGas,omitempty"`
	AuthorizationList    []*stAuthorization  `json:"authorizationList,omitempty"`
}

type stTransactionMarshaling struct {
	GasPrice             *math.HexOrDecimal256
	MaxFeePerGas         *math.HexOrDecimal256
	MaxPriorityFeePerGas *math.HexOrDecimal256
	Nonce                math.HexOrDecimal64
	GasLimit             []math.HexOrDecimal64
	PrivateKey           hexutil.Bytes
	BlobGasFeeCap        *math.HexOrDecimal256
}

//go:generate go run github.com/fjl/gencodec -type stAuthorization -field-override stAuthorizationMarshaling -out gen_stauthorization.go

// Authorization is an authorization from an account to deploy code at it's address.
type stAuthorization struct {
	ChainID *big.Int        `json:"chainId" gencodec:"required"`
	Address common.Address  `json:"address" gencodec:"required"`
	Nonce   uint64          `json:"nonce" gencodec:"required"`
	V       uint8           `json:"v" gencodec:"required"`
	R       *big.Int        `json:"r" gencodec:"required"`
	S       *big.Int        `json:"s" gencodec:"required"`
	Signer  *common.Address `json:"signer"`
}

// field type overrides for gencodec
type stAuthorizationMarshaling struct {
	ChainID *math.HexOrDecimal256
	Nonce   math.HexOrDecimal64
	V       math.HexOrDecimal64
	R       *math.HexOrDecimal256
	S       *math.HexOrDecimal256
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	_ = rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

// CopyAndDropAuth returns a copy of the authorizationlist, with the item at the given
// index dropped
func CopyAndDropAuth(list []*stAuthorization, index int) []*stAuthorization {
	var cpy []*stAuthorization
	for i, auth := range list {
		if i != index {
			cpy = append(cpy, auth)
		}
	}
	return cpy
}
