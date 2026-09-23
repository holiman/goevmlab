// Copyright 2026 Martin Holst Swende
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
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/types/bal"
	"github.com/ethereum/go-ethereum/params"
)

// This file contains the JSON schema for blockchain tests, in the two flavours
// produced by execution-spec-tests (EEST):
//
//   - blockchain_test: blocks are given as RLP, the client has to construct the
//     block access list itself and verify it against the header.
//   - blockchain_test_engine: blocks are given as engine_newPayload parameters,
//     including the block access list, so the client can take the BAL-driven
//     (parallel) execution path.
//
// go-ethereum's tests.BlockTest only has an unmarshaller, so the types are
// redefined here with the field names used by EEST. Both go-ethereum and
// Nethermind accept these names.

// BlockchainTest is a set of named blockchain tests (the blockchain_test flavour).
type BlockchainTest map[string]*BlockTestJSON

// EngineTest is a set of named blockchain tests in engine flavour.
type EngineTest map[string]*EngineTestJSON

// BlockTestJSON is a single test in the blockchain_test fixture format.
type BlockTestJSON struct {
	Blocks     []BtBlockJSON `json:"blocks"`
	Genesis    *BtHeaderJSON `json:"genesisBlockHeader"`
	GenesisRLP hexutil.Bytes `json:"genesisRLP"`
	Pre        GenesisAlloc  `json:"pre"`
	Post       GenesisAlloc  `json:"postState"`
	LastBlock  string        `json:"lastblockhash"`
	Network    string        `json:"network"`
	SealEngine string        `json:"sealEngine"`
	Config     *BtConfigJSON `json:"config,omitempty"`
}

// EngineTestJSON is a single test in the blockchain_test_engine fixture format.
type EngineTestJSON struct {
	Network   string            `json:"network"`
	Genesis   *BtHeaderJSON     `json:"genesisBlockHeader"`
	Pre       GenesisAlloc      `json:"pre"`
	Post      GenesisAlloc      `json:"postState"`
	LastBlock string            `json:"lastblockhash"`
	Payloads  []BtEnginePayload `json:"engineNewPayloads"`
	Config    *BtConfigJSON     `json:"config,omitempty"`
}

// BtEnginePayload is one engine_newPayload invocation.
type BtEnginePayload struct {
	// Params are, for newPayloadV4 and later:
	// [executionPayload, blobVersionedHashes, parentBeaconBlockRoot, executionRequests]
	Params                   []any  `json:"params"`
	NewPayloadVersion        string `json:"newPayloadVersion"`
	ForkchoiceUpdatedVersion string `json:"forkchoiceUpdatedVersion"`
	ValidationError          string `json:"validationError,omitempty"`
}

// BtConfigJSON mirrors the EEST fixture 'config' object.
type BtConfigJSON struct {
	Network      string                    `json:"network"`
	ChainID      math.HexOrDecimal64       `json:"chainid"`
	BlobSchedule map[string]stBlobSchedule `json:"blobSchedule,omitempty"`
}

// BtBlockJSON is a single block in a blockchain_test fixture.
type BtBlockJSON struct {
	BlockHeader     *BtHeaderJSON   `json:"blockHeader"`
	RLP             hexutil.Bytes   `json:"rlp"`
	BlockNumber     string          `json:"blocknumber"`
	UncleHeaders    []*BtHeaderJSON `json:"uncleHeaders"`
	ExpectException string          `json:"expectException,omitempty"`
	// BlockAccessList is not part of the block RLP; it is included for
	// debugging (and for clients that can consume it directly).
	BlockAccessList *bal.BlockAccessList `json:"blockAccessList,omitempty"`
}

// BtHeaderJSON is a block header with the field names used by EEST fixtures.
type BtHeaderJSON struct {
	Bloom                 types.Bloom           `json:"bloom"`
	Coinbase              common.Address        `json:"coinbase"`
	MixHash               common.Hash           `json:"mixHash"`
	Nonce                 types.BlockNonce      `json:"nonce"`
	Number                *math.HexOrDecimal256 `json:"number"`
	Hash                  common.Hash           `json:"hash"`
	ParentHash            common.Hash           `json:"parentHash"`
	ReceiptTrie           common.Hash           `json:"receiptTrie"`
	StateRoot             common.Hash           `json:"stateRoot"`
	TransactionsTrie      common.Hash           `json:"transactionsTrie"`
	UncleHash             common.Hash           `json:"uncleHash"`
	ExtraData             hexutil.Bytes         `json:"extraData"`
	Difficulty            *math.HexOrDecimal256 `json:"difficulty"`
	GasLimit              math.HexOrDecimal64   `json:"gasLimit"`
	GasUsed               math.HexOrDecimal64   `json:"gasUsed"`
	Timestamp             math.HexOrDecimal64   `json:"timestamp"`
	BaseFeePerGas         *math.HexOrDecimal256 `json:"baseFeePerGas,omitempty"`
	WithdrawalsRoot       *common.Hash          `json:"withdrawalsRoot,omitempty"`
	BlobGasUsed           *math.HexOrDecimal64  `json:"blobGasUsed,omitempty"`
	ExcessBlobGas         *math.HexOrDecimal64  `json:"excessBlobGas,omitempty"`
	ParentBeaconBlockRoot *common.Hash          `json:"parentBeaconBlockRoot,omitempty"`
	RequestsHash          *common.Hash          `json:"requestsHash,omitempty"`
	BlockAccessListHash   *common.Hash          `json:"blockAccessListHash,omitempty"`
	SlotNumber            *math.HexOrDecimal64  `json:"slotNumber,omitempty"`
}

// headerToBt converts a go-ethereum header into the fixture representation.
func headerToBt(h *types.Header) *BtHeaderJSON {
	o := &BtHeaderJSON{
		Bloom:                 h.Bloom,
		Coinbase:              h.Coinbase,
		MixHash:               h.MixDigest,
		Nonce:                 h.Nonce,
		Number:                (*math.HexOrDecimal256)(h.Number),
		Hash:                  h.Hash(),
		ParentHash:            h.ParentHash,
		ReceiptTrie:           h.ReceiptHash,
		StateRoot:             h.Root,
		TransactionsTrie:      h.TxHash,
		UncleHash:             h.UncleHash,
		ExtraData:             h.Extra,
		Difficulty:            (*math.HexOrDecimal256)(h.Difficulty),
		GasLimit:              math.HexOrDecimal64(h.GasLimit),
		GasUsed:               math.HexOrDecimal64(h.GasUsed),
		Timestamp:             math.HexOrDecimal64(h.Time),
		BaseFeePerGas:         (*math.HexOrDecimal256)(h.BaseFee),
		WithdrawalsRoot:       h.WithdrawalsHash,
		ParentBeaconBlockRoot: h.ParentBeaconRoot,
		RequestsHash:          h.RequestsHash,
		BlockAccessListHash:   h.BlockAccessListHash,
	}
	if h.Extra == nil {
		o.ExtraData = hexutil.Bytes{}
	}
	if h.BlobGasUsed != nil {
		o.BlobGasUsed = (*math.HexOrDecimal64)(h.BlobGasUsed)
	}
	if h.ExcessBlobGas != nil {
		o.ExcessBlobGas = (*math.HexOrDecimal64)(h.ExcessBlobGas)
	}
	if h.SlotNumber != nil {
		o.SlotNumber = (*math.HexOrDecimal64)(h.SlotNumber)
	}
	return o
}

// btConfig builds the fixture 'config' object for the given chain config.
func btConfig(network string, config *params.ChainConfig) *BtConfigJSON {
	c := &BtConfigJSON{
		Network: network,
		ChainID: math.HexOrDecimal64(config.ChainID.Uint64()),
	}
	if bs := config.BlobScheduleConfig; bs != nil {
		c.BlobSchedule = make(map[string]stBlobSchedule)
		add := func(name string, bc *params.BlobConfig) {
			if bc != nil {
				c.BlobSchedule[name] = stBlobSchedule{
					Target:                math.HexOrDecimal64(bc.Target),
					Max:                   math.HexOrDecimal64(bc.Max),
					BaseFeeUpdateFraction: math.HexOrDecimal64(bc.UpdateFraction),
				}
			}
		}
		add("Cancun", bs.Cancun)
		add("Prague", bs.Prague)
		add("BPO1", bs.BPO1)
		add("BPO2", bs.BPO2)
		add("BPO3", bs.BPO3)
		add("BPO4", bs.BPO4)
		add("BPO5", bs.BPO5)
	}
	return c
}

// engineVersions returns the engine_newPayload and engine_forkchoiceUpdated
// versions to use for a block at the given time, following the version bumps
// defined by EEST.
func engineVersions(config *params.ChainConfig, num *big.Int, time uint64) (newPayload, fcu string) {
	switch {
	case config.IsAmsterdam(num, time):
		return "5", "4"
	case config.IsPrague(num, time):
		return "4", "3"
	case config.IsCancun(num, time):
		return "3", "3"
	case config.IsShanghai(num, time):
		return "2", "2"
	default:
		return "1", "1"
	}
}

// enginePayloadParams builds the newPayload parameter list for a block.
func enginePayloadParams(block *types.Block, requests [][]byte) ([]any, error) {
	envelope := engine.BlockToExecutableData(block, common.Big0, nil, requests)
	payload, err := json.Marshal(envelope.ExecutionPayload)
	if err != nil {
		return nil, err
	}
	var versionedHashes []common.Hash
	for _, tx := range block.Transactions() {
		versionedHashes = append(versionedHashes, tx.BlobHashes()...)
	}
	if versionedHashes == nil {
		versionedHashes = []common.Hash{}
	}
	reqs := make([]hexutil.Bytes, 0, len(requests))
	for _, r := range requests {
		reqs = append(reqs, r)
	}
	params := []any{json.RawMessage(payload)}
	if block.Header().ParentBeaconRoot != nil { // Cancun+
		params = append(params, versionedHashes, block.Header().ParentBeaconRoot)
	}
	if block.Header().RequestsHash != nil { // Prague+
		params = append(params, reqs)
	}
	return params, nil
}
