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
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/tests"
)

// buildTestBt builds a small Amsterdam blocktest from the given fillers.
func buildTestBt(t *testing.T, fork string, fillerNames []string, blocks, txs int) *BtMaker {
	t.Helper()
	factory, err := BtFactory(fork, BtOptions{Fillers: fillerNames, Blocks: blocks, TxsPerBlock: txs})
	if err != nil {
		t.Fatal(err)
	}
	// Generation may legitimately fail (e.g. a filler producing a tx that
	// does not fit); retry a few times.
	for i := 0; i < 5; i++ {
		bt, err := factory()
		if err != nil {
			t.Fatal(err)
		}
		if err := bt.Build(); err != nil {
			t.Logf("build attempt %d failed: %v", i, err)
			continue
		}
		return bt
	}
	t.Fatal("could not build blocktest")
	return nil
}

// runGethBlockTest runs the blockchain_test fixture through go-ethereum's
// blocktest runner, in-process.
func runGethBlockTest(t *testing.T, bt *BlockchainTest) error {
	t.Helper()
	data, err := json.Marshal(bt)
	if err != nil {
		t.Fatal(err)
	}
	var gethTests map[string]*tests.BlockTest
	if err := json.Unmarshal(data, &gethTests); err != nil {
		t.Fatal(err)
	}
	if len(gethTests) != 1 {
		t.Fatalf("expected 1 test, got %d", len(gethTests))
	}
	for _, test := range gethTests {
		return test.Run(false, rawdb.HashScheme, false, nil, nil)
	}
	return nil
}

func TestBlockTestRoundTrip(t *testing.T) {
	for _, fork := range []string{"Amsterdam", "Osaka"} {
		t.Run(fork, func(t *testing.T) {
			bt := buildTestBt(t, fork, []string{"naive", "sstore_sload", "simpleops"}, 3, 3)
			blocks := bt.Blocks()
			if len(blocks) != 3 {
				t.Fatalf("expected 3 blocks, got %d", len(blocks))
			}
			for _, block := range blocks {
				if fork == "Amsterdam" {
					if block.Header().BlockAccessListHash == nil {
						t.Fatalf("block %d has no access list hash", block.NumberU64())
					}
					if block.AccessList() == nil {
						t.Fatalf("block %d has no access list", block.NumberU64())
					}
				} else if block.Header().BlockAccessListHash != nil {
					t.Fatalf("block %d has access list hash before Amsterdam", block.NumberU64())
				}
			}
			test, err := bt.ToBlockchainTest("test")
			if err != nil {
				t.Fatal(err)
			}
			if err := runGethBlockTest(t, test); err != nil {
				t.Fatalf("blocktest failed: %v", err)
			}
			// Also check that the JSON survives a re-encode (i.e. that unmarshalling
			// our own format works too).
			data, _ := json.Marshal(test)
			var again BlockchainTest
			if err := json.Unmarshal(data, &again); err != nil {
				t.Fatal(err)
			}
			if len(again["test"].Post) == 0 {
				t.Fatal("empty post state")
			}
		})
	}
}

func TestBlockTestBadBAL(t *testing.T) {
	bt := buildTestBt(t, "Amsterdam", []string{"sstore_sload"}, 1, 2)
	test, err := bt.ToBlockchainTest("test")
	if err != nil {
		t.Fatal(err)
	}
	// Tamper with the access list hash of the first block
	tj := (*test)["test"]
	var block types.Block
	if err := rlp.DecodeBytes(tj.Blocks[0].RLP, &block); err != nil {
		t.Fatal(err)
	}
	header := block.Header()
	bad := common.HexToHash("0xdead")
	header.BlockAccessListHash = &bad
	tampered := types.NewBlockWithHeader(header).WithBody(*block.Body())
	enc, _ := rlp.EncodeToBytes(tampered)
	tj.Blocks[0].RLP = enc
	tj.Blocks[0].BlockHeader = headerToBt(tampered.Header())
	tj.Blocks = tj.Blocks[:1]
	tj.LastBlock = tampered.Hash().Hex()

	err = runGethBlockTest(t, test)
	if err == nil {
		t.Fatal("expected tampered blocktest to fail")
	}
	if !strings.Contains(err.Error(), "access list hash mismatch") {
		t.Fatalf("expected access list hash mismatch, got: %v", err)
	}
}

func TestEngineTestFormat(t *testing.T) {
	bt := buildTestBt(t, "Amsterdam", []string{"naive"}, 2, 2)
	test, err := bt.ToEngineTest("test")
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(test)
	if err != nil {
		t.Fatal(err)
	}
	// Decode it back and reconstruct the blocks the way an engine API
	// consumer would.
	var decoded map[string]struct {
		Payloads []struct {
			Params                   []json.RawMessage `json:"params"`
			NewPayloadVersion        string            `json:"newPayloadVersion"`
			ForkchoiceUpdatedVersion string            `json:"forkchoiceUpdatedVersion"`
		} `json:"engineNewPayloads"`
		LastBlock string `json:"lastblockhash"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	payloads := decoded["test"].Payloads
	if len(payloads) != 2 {
		t.Fatalf("expected 2 payloads, got %d", len(payloads))
	}
	for i, p := range payloads {
		if p.NewPayloadVersion != "5" || p.ForkchoiceUpdatedVersion != "4" {
			t.Fatalf("unexpected versions %v/%v", p.NewPayloadVersion, p.ForkchoiceUpdatedVersion)
		}
		if len(p.Params) != 4 {
			t.Fatalf("expected 4 params, got %d", len(p.Params))
		}
		var (
			ed         engine.ExecutableData
			hashes     []common.Hash
			beaconRoot common.Hash
			requests   []hexutil.Bytes
		)
		if err := json.Unmarshal(p.Params[0], &ed); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(p.Params[1], &hashes); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(p.Params[2], &beaconRoot); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(p.Params[3], &requests); err != nil {
			t.Fatal(err)
		}
		if len(ed.BlockAccessList) == 0 {
			t.Fatalf("payload %d has no block access list", i)
		}
		reqs := make([][]byte, 0, len(requests))
		for _, r := range requests {
			reqs = append(reqs, r)
		}
		block, err := engine.ExecutableDataToBlock(ed, hashes, &beaconRoot, reqs)
		if err != nil {
			t.Fatalf("payload %d: %v", i, err)
		}
		if block.Hash() != bt.Blocks()[i].Hash() {
			t.Fatalf("payload %d: hash mismatch: %x != %x", i, block.Hash(), bt.Blocks()[i].Hash())
		}
		if block.AccessList() == nil {
			t.Fatalf("payload %d: reconstructed block has no access list", i)
		}
	}
	if decoded["test"].LastBlock != bt.Blocks()[1].Hash().Hex() {
		t.Fatal("lastblockhash mismatch")
	}
}
