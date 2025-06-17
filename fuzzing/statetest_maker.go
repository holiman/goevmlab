// Copyright 2019 Martin Holst Swende
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
	"io"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/tests"
)

// DisallowEOF makes it so that any statetest that are created never
// contain 0xEF-prefixed code.
// See https://github.com/holiman/goevmlab/issues/127
const DisallowEOF = true

// The sender
var sender = common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")
var pKey = hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8")

// GstMaker is a construct to generate General State Tests
type GstMaker struct {
	pre   *GenesisAlloc
	env   *stEnv
	tx    StTransaction
	forks []string
	root  common.Hash
	logs  common.Hash
}

func NewGstMaker() *GstMaker {
	alloc := make(GenesisAlloc)
	rnd := common.HexToHash("0x200000")
	gst := &GstMaker{
		env: &stEnv{
			// The ENV portion
			Number:     1,
			GasLimit:   0x26e1f476fe1e22,
			Difficulty: big.NewInt(0x200000),
			Random:     &rnd,
			Coinbase:   common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
			Timestamp:  0x03e8,
			BaseFee:    big.NewInt(0x10),
			// Keccak256([]byte{'0'}) see https://github.com/hyperledger/besu/issues/5122
			PreviousHash: common.HexToHash("0x044852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d"),
		},
		pre: &alloc,
	}
	return gst
}

func (g *GstMaker) SetPre(genesis *GenesisAlloc) {
	g.pre = genesis
}

func (g *GstMaker) AddAccount(address common.Address, a GenesisAccount) {
	// See https://github.com/holiman/goevmlab/issues/127
	// We must not allow any code in genesis to start with `0xEF`, otherwise
	// evmone will reject the test.
	// If the constant 'DisallowEof' is set, then we change `0xEF...` into `0xEE...`
	if DisallowEOF && len(a.Code) > 0 && a.Code[0] == 0xEF {
		a.Code[0] = 0xEE
	}
	alloc := *g.pre
	alloc[address] = a
}

// GetDestination returns the to- address from the tx
func (g *GstMaker) GetDestination() common.Address {
	return common.HexToAddress(g.tx.To)
}

// SetCode sets the code at the given address (creating the account
// if it did not previously exist)
func (g *GstMaker) SetCode(address common.Address, code []byte) {
	// See https://github.com/holiman/goevmlab/issues/127
	// We must not allow any code in genesis to start with `0xEF`, otherwise
	// evmone will reject the test.
	// If the constant 'DisallowEof' is set, then we change `0xEF...` into `0xEE...`
	if DisallowEOF && len(code) > 0 && code[0] == 0xEF {
		code[0] = 0xEE
	}
	alloc := *g.pre
	account, exist := alloc[address]
	if !exist {
		account = GenesisAccount{
			Code:    code,
			Storage: make(map[common.Hash]common.Hash),
			Nonce:   0,
			Balance: new(big.Int),
		}
	} else {
		account.Code = code
	}
	alloc[address] = account
}

func (g *GstMaker) SetResult(root, logs common.Hash) {
	g.root = root
	g.logs = logs
}

func (g *GstMaker) SetTx(tx *StTransaction) {
	g.tx = *tx
}

func (g *GstMaker) ToSubTest() *stJSON {
	st := &stJSON{}
	st.Pre = *g.pre
	st.Env = *g.env
	st.Tx = g.tx
	for _, fork := range g.forks {
		postState := make(map[string][]stPostState)
		postState[fork] = []stPostState{
			stPostState{
				Logs:    g.logs,
				Root:    g.root,
				Indexes: stIndex{Gas: 0, Value: 0, Data: 0},
			},
		}
		st.Post = postState
	}
	return st
}

func (g *GstMaker) ToGeneralStateTest(name string) *GeneralStateTest {
	gst := make(GeneralStateTest)
	gst[name] = g.ToSubTest()
	return &gst
}

func FromGeneralStateTest(name string) (*GeneralStateTest, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	gst := make(GeneralStateTest)
	err = json.Unmarshal(data, &gst)
	return &gst, err
}

func (g *GstMaker) ToStateTest() (tests.StateTest, error) {

	stjson := g.ToSubTest()
	var gethStateTest tests.StateTest
	data, err := json.Marshal(stjson)
	if err != nil {
		return gethStateTest, err
	}
	if err := json.Unmarshal(data, &gethStateTest); err != nil {
		return gethStateTest, err
	}
	return gethStateTest, nil
}

func (g *GstMaker) EnableFork(fork string) {
	g.forks = append(g.forks, fork)
}

// Fill uses go-ethereum internally to determine the state root and logs, and optionally
// outputs the trace to the given writer (if non-nil)
func (g *GstMaker) Fill(traceOutput io.Writer) error {

	test, err := g.ToStateTest()
	if err != nil {
		return err
	}
	subtest := test.Subtests()[0]
	cfg := vm.Config{}
	if traceOutput != nil {
		cfg.Tracer = logger.NewJSONLogger(&logger.Config{}, traceOutput)
	}
	state, root, _, err := test.RunNoVerify(subtest, cfg, false, rawdb.HashScheme)
	if err != nil {
		return err
	}
	defer state.Close()
	logs := rlpHash(state.StateDB.Logs())
	g.SetResult(root, logs)
	return nil
}

func BasicStateTest(fork string) *GstMaker {
	gst := NewGstMaker()
	// Add sender
	gst.AddAccount(sender, GenesisAccount{
		Nonce:   0,
		Balance: big.NewInt(0xffffffffff),
		Storage: make(map[common.Hash]common.Hash),
		Code:    []byte{},
	})
	gst.EnableFork(fork)
	return gst
}

func AddTransaction(dest *common.Address, gst *GstMaker) {
	tx := &StTransaction{
		// 8M gaslimit
		To:         dest.Hex(),
		GasLimit:   []uint64{8000000},
		Nonce:      0,
		Value:      []string{"0x01"},
		Data:       []string{"0x"},
		GasPrice:   big.NewInt(0x16),
		Sender:     sender,
		PrivateKey: pKey,
	}
	gst.SetTx(tx)
}
