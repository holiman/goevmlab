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
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/tests"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

// The coinbase used for generated blocks.
var btCoinbase = common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")

// btSenderBalance is the balance given to the sender account in blocktests.
// It needs to be large enough to fund any number of transactions with the
// gas prices and values produced by the fillers.
var btSenderBalance = new(big.Int).Lsh(big.NewInt(1), 200)

// btBlock is the plan for one block: transactions and withdrawals.
type btBlock struct {
	txs         []*StTransaction
	withdrawals []*types.Withdrawal
}

// btBuiltBlock is a block after generation, with its side products.
type btBuiltBlock struct {
	block    *types.Block
	receipts types.Receipts
	requests [][]byte
}

// BtMaker is a construct to generate blockchain tests. The pre-state and
// transactions are typically produced by the state test fillers, and then
// executed in-process via go-ethereum's chain generator, which computes the
// block headers, the block access lists (EIP-7928) and the post state.
type BtMaker struct {
	fork   string
	config *params.ChainConfig
	pre    GenesisAlloc
	plan   []*btBlock

	// Filled by Build
	genesis *core.Genesis
	built   []btBuiltBlock
	post    GenesisAlloc
}

// NewBtMaker creates a new blocktest maker for the given fork. The pre-state
// contains the sender account and the system contracts (EIP-4788, EIP-2935,
// EIP-7002, EIP-7251, EIP-8282, EIP-7997), which post-Prague block processing
// requires to be present.
func NewBtMaker(fork string) (*BtMaker, error) {
	src, ok := tests.Forks[fork]
	if !ok {
		return nil, fmt.Errorf("unknown fork %q", fork)
	}
	config := *src
	if config.TerminalTotalDifficulty == nil {
		config.TerminalTotalDifficulty = big.NewInt(0)
	}
	b := &BtMaker{
		fork:   fork,
		config: &config,
		pre:    make(GenesisAlloc),
	}
	for addr, acc := range core.SystemContractAllocs() {
		b.pre[addr] = GenesisAccount{
			Code:    acc.Code,
			Storage: make(map[common.Hash]common.Hash),
			Balance: new(big.Int),
			Nonce:   acc.Nonce,
		}
	}
	b.pre[sender] = GenesisAccount{
		Balance: new(big.Int).Set(btSenderBalance),
		Storage: make(map[common.Hash]common.Hash),
		Code:    []byte{},
	}
	return b, nil
}

// Fork returns the fork name of the test.
func (b *BtMaker) Fork() string {
	return b.fork
}

// ChainConfig returns the chain configuration used for the test.
func (b *BtMaker) ChainConfig() *params.ChainConfig {
	return b.config
}

// Pre returns the pre-state, which may be modified by the caller before Build.
func (b *BtMaker) Pre() GenesisAlloc {
	return b.pre
}

// AddAccount adds an account to the pre-state. Existing accounts (other than
// the sender, which is always kept as-is) are overwritten.
func (b *BtMaker) AddAccount(address common.Address, a GenesisAccount) {
	if address == sender {
		return
	}
	if DisallowEOF && len(a.Code) > 0 && a.Code[0] == 0xEF {
		a.Code[0] = 0xEE
	}
	if a.Storage == nil {
		a.Storage = make(map[common.Hash]common.Hash)
	}
	if a.Balance == nil {
		a.Balance = new(big.Int)
	}
	b.pre[address] = a
}

// MergePre adds all accounts of the given alloc which do not yet exist in the
// pre-state (first one wins). The sender account is never overwritten.
func (b *BtMaker) MergePre(alloc GenesisAlloc) {
	for addr, acc := range alloc {
		if _, exist := b.pre[addr]; exist {
			continue
		}
		b.AddAccount(addr, acc)
	}
}

// AddBlock starts a new block. Transactions added afterwards go into it.
func (b *BtMaker) AddBlock() {
	b.plan = append(b.plan, &btBlock{})
}

// NumBlocks returns the number of planned blocks.
func (b *BtMaker) NumBlocks() int {
	return len(b.plan)
}

func (b *BtMaker) current() *btBlock {
	if len(b.plan) == 0 {
		b.AddBlock()
	}
	return b.plan[len(b.plan)-1]
}

// AddTx adds a transaction to the current block. The transaction is sent by
// the sender account: the nonce is assigned at build time, and the gas price
// is raised to the base fee if needed.
func (b *BtMaker) AddTx(tx *StTransaction) {
	cpy := *tx
	blk := b.current()
	blk.txs = append(blk.txs, &cpy)
}

// AddTxFromGst merges the pre-state of a filled state test maker into this
// blocktest, and adds its transaction to the current block.
func (b *BtMaker) AddTxFromGst(gst *GstMaker) {
	b.MergePre(*gst.pre)
	b.AddTx(&gst.tx)
}

// AddWithdrawal adds a withdrawal to the current block.
func (b *BtMaker) AddWithdrawal(validator uint64, address common.Address, amount uint64) {
	blk := b.current()
	blk.withdrawals = append(blk.withdrawals, &types.Withdrawal{
		Validator: validator,
		Address:   address,
		Amount:    amount,
	})
}

// Genesis returns the genesis spec of the test. It is available after Build.
func (b *BtMaker) Genesis() *core.Genesis {
	return b.genesis
}

// Blocks returns the generated blocks. Only available after Build.
func (b *BtMaker) Blocks() []*types.Block {
	var blocks []*types.Block
	for _, bb := range b.built {
		blocks = append(blocks, bb.block)
	}
	return blocks
}

// Requests returns the execution requests (EIP-7685) of each generated block.
func (b *BtMaker) Requests() [][][]byte {
	var reqs [][][]byte
	for _, bb := range b.built {
		reqs = append(reqs, bb.requests)
	}
	return reqs
}

func (b *BtMaker) makeGenesis() *core.Genesis {
	alloc := make(types.GenesisAlloc)
	for addr, acc := range b.pre {
		alloc[addr] = types.Account{
			Code:    acc.Code,
			Storage: acc.Storage,
			Balance: acc.Balance,
			Nonce:   acc.Nonce,
		}
	}
	g := &core.Genesis{
		Config:     b.config,
		Nonce:      0,
		Timestamp:  0,
		ExtraData:  []byte{},
		GasLimit:   100_000_000,
		Difficulty: big.NewInt(0),
		Mixhash:    common.Hash{},
		Coinbase:   common.Address{},
		Alloc:      alloc,
		Number:     0,
		BaseFee:    big.NewInt(0x10),
	}
	if b.config.IsCancun(common.Big0, 0) {
		g.BlobGasUsed = new(uint64)
		g.ExcessBlobGas = new(uint64)
	}
	if b.config.IsAmsterdam(common.Big0, 0) {
		g.SlotNumber = new(uint64)
	}
	return g
}

// Build executes the planned blocks. Afterwards, the blocks, requests, and
// post state are available. An error is returned if any transaction cannot be
// applied (e.g. insufficient gas); the caller should then generate a new test.
func (b *BtMaker) Build() (err error) {
	if len(b.plan) == 0 {
		return errors.New("no blocks planned")
	}
	key, err := crypto.ToECDSA(pKey)
	if err != nil {
		return err
	}
	gspec := b.makeGenesis()
	engine := beacon.New(ethash.NewFaker())
	signer := types.LatestSigner(b.config)

	// go-ethereum's chain generator panics if a transaction cannot be applied
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("block generation failed: %v", r)
		}
	}()
	db, blocks, receipts := core.GenerateChainWithGenesis(gspec, engine, len(b.plan), func(i int, gen *core.BlockGen) {
		gen.SetCoinbase(btCoinbase)
		plan := b.plan[i]
		for _, st := range plan.txs {
			tx, err := b.toTransaction(st, gen, signer, key)
			if err != nil {
				panic(err)
			}
			gen.AddTx(tx)
		}
		for _, w := range plan.withdrawals {
			gen.AddWithdrawal(w)
		}
	})
	if err := b.verify(gspec, engine, blocks); err != nil {
		return err
	}
	b.genesis = gspec
	b.built = b.built[:0]
	for i, block := range blocks {
		b.built = append(b.built, btBuiltBlock{block: block, receipts: receipts[i]})
	}
	if err := b.collectRequests(gspec, engine); err != nil {
		return err
	}
	return b.collectPost(db)
}

// verify re-imports the generated blocks, stripped of their access lists, into
// a fresh chain. This ensures the fixture is self-consistent: the importer
// must reconstruct the same block access list hash, requests hash etc.
func (b *BtMaker) verify(gspec *core.Genesis, engine *beacon.Beacon, blocks []*types.Block) error {
	chain, err := core.NewBlockChain(rawdb.NewMemoryDatabase(), gspec, engine, &core.BlockChainConfig{
		StateScheme:        rawdb.HashScheme,
		TxLookupLimit:      -1,
		SlowBlockThreshold: time.Hour, // silence the slow-block logging
	})
	if err != nil {
		return err
	}
	defer chain.Stop()
	var stripped []*types.Block
	for _, block := range blocks {
		enc, err := rlp.EncodeToBytes(block)
		if err != nil {
			return err
		}
		var dec types.Block
		if err := rlp.DecodeBytes(enc, &dec); err != nil {
			return err
		}
		stripped = append(stripped, &dec)
	}
	if n, err := chain.InsertChain(stripped); err != nil {
		return fmt.Errorf("re-import of block %d failed: %w", n+1, err)
	}
	return nil
}

// collectRequests re-executes the blocks (with access lists attached) in order
// to obtain the execution requests, which are needed for the engine flavour
// of the test, and to double-check the block access lists.
func (b *BtMaker) collectRequests(gspec *core.Genesis, engine *beacon.Beacon) error {
	chain, err := core.NewBlockChain(rawdb.NewMemoryDatabase(), gspec, engine, &core.BlockChainConfig{
		StateScheme:        rawdb.HashScheme,
		TxLookupLimit:      -1,
		SlowBlockThreshold: time.Hour, // silence the slow-block logging
	})
	if err != nil {
		return err
	}
	defer chain.Stop()
	processor := core.NewStateProcessor(chain)
	parent := chain.Genesis()
	for i := range b.built {
		block := b.built[i].block
		statedb, err := chain.StateAt(parent.Header())
		if err != nil {
			return err
		}
		res, err := processor.Process(context.Background(), block, statedb, core.NewJumpDestCache(), vm.NewPrecompileCache(), vm.Config{}, nil)
		if err != nil {
			return fmt.Errorf("re-execution of block %d failed: %w", block.NumberU64(), err)
		}
		if res.Bal != nil && b.config.IsAmsterdam(block.Number(), block.Time()) {
			have := res.Bal.ToEncodingObj().Hash()
			if want := block.Header().BlockAccessListHash; want == nil || *want != have {
				return fmt.Errorf("block %d: access list hash mismatch, header %v, re-executed %v", block.NumberU64(), want, have)
			}
		}
		b.built[i].requests = res.Requests
		if b.built[i].requests == nil {
			b.built[i].requests = [][]byte{}
		}
		if _, err := chain.InsertChain([]*types.Block{block}); err != nil {
			return err
		}
		parent = block
	}
	return nil
}

// collectPost reads the post state of the accounts that were touched. The
// accounts and slots are those in the pre-state plus those in the block
// access lists.
func (b *BtMaker) collectPost(db ethdb.Database) error {
	if len(b.built) == 0 {
		return errors.New("no blocks built")
	}
	root := b.built[len(b.built)-1].block.Root()
	tdb := triedb.NewDatabase(db, triedb.HashDefaults)
	defer tdb.Close()
	statedb, err := state.New(root, state.NewDatabase(tdb, nil))
	if err != nil {
		return err
	}
	b.post = b.readPost(statedb)
	return nil
}

// toTransaction converts a state test transaction into a signed transaction
// for the block being generated.
func (b *BtMaker) toTransaction(st *StTransaction, gen *core.BlockGen, signer types.Signer, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	if len(st.BlobVersionedHashes) > 0 {
		return nil, errors.New("blob transactions are not supported in blocktests")
	}
	var (
		to       *common.Address
		data     []byte
		value    = new(big.Int)
		gas      uint64
		nonce    = gen.TxNonce(sender)
		baseFee  = gen.BaseFee()
		chainID  = b.config.ChainID
		accessLs types.AccessList
	)
	if st.To != "" {
		addr := common.HexToAddress(st.To)
		to = &addr
	}
	if len(st.Data) > 0 {
		data = common.FromHex(st.Data[0])
	}
	if len(st.Value) > 0 && st.Value[0] != "" && st.Value[0] != "0x" {
		v, ok := math.ParseBig256(st.Value[0])
		if !ok {
			return nil, fmt.Errorf("invalid value %q", st.Value[0])
		}
		value = v
	}
	if len(st.GasLimit) > 0 {
		gas = st.GasLimit[0]
	}
	if gas > params.MaxTxGas {
		gas = params.MaxTxGas
	}
	if len(st.AccessLists) > 0 && st.AccessLists[0] != nil {
		accessLs = *st.AccessLists[0]
	}
	// Fee handling: raise the price to the base fee if needed
	atLeast := func(x *big.Int) *big.Int {
		if x == nil || x.Cmp(baseFee) < 0 {
			return new(big.Int).Set(baseFee)
		}
		return x
	}
	var txdata types.TxData
	switch {
	case len(st.AuthorizationList) > 0:
		if to == nil {
			return nil, errors.New("setcode tx without recipient")
		}
		var auths []types.SetCodeAuthorization
		for _, a := range st.AuthorizationList {
			auths = append(auths, types.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(a.ChainID),
				Address: a.Address,
				Nonce:   a.Nonce,
				V:       a.V,
				R:       *uint256.MustFromBig(a.R),
				S:       *uint256.MustFromBig(a.S),
			})
		}
		feeCap := atLeast(st.MaxFeePerGas)
		tipCap := st.MaxPriorityFeePerGas
		if tipCap == nil {
			tipCap = new(big.Int)
		}
		if tipCap.Cmp(feeCap) > 0 {
			tipCap = feeCap
		}
		txdata = &types.SetCodeTx{
			ChainID:    uint256.MustFromBig(chainID),
			Nonce:      nonce,
			GasTipCap:  uint256.MustFromBig(tipCap),
			GasFeeCap:  uint256.MustFromBig(feeCap),
			Gas:        gas,
			To:         *to,
			Value:      uint256.MustFromBig(value),
			Data:       data,
			AccessList: accessLs,
			AuthList:   auths,
		}
	case st.MaxFeePerGas != nil || st.MaxPriorityFeePerGas != nil:
		feeCap := atLeast(st.MaxFeePerGas)
		tipCap := st.MaxPriorityFeePerGas
		if tipCap == nil {
			tipCap = new(big.Int)
		}
		if tipCap.Cmp(feeCap) > 0 {
			tipCap = feeCap
		}
		txdata = &types.DynamicFeeTx{
			ChainID:    chainID,
			Nonce:      nonce,
			GasTipCap:  tipCap,
			GasFeeCap:  feeCap,
			Gas:        gas,
			To:         to,
			Value:      value,
			Data:       data,
			AccessList: accessLs,
		}
	case len(st.AccessLists) > 0:
		txdata = &types.AccessListTx{
			ChainID:    chainID,
			Nonce:      nonce,
			GasPrice:   atLeast(st.GasPrice),
			Gas:        gas,
			To:         to,
			Value:      value,
			Data:       data,
			AccessList: accessLs,
		}
	default:
		txdata = &types.LegacyTx{
			Nonce:    nonce,
			GasPrice: atLeast(st.GasPrice),
			Gas:      gas,
			To:       to,
			Value:    value,
			Data:     data,
		}
	}
	return types.SignNewTx(key, signer, txdata)
}

// ToBlockchainTest returns the test in blockchain_test format.
func (b *BtMaker) ToBlockchainTest(name string) (*BlockchainTest, error) {
	if b.genesis == nil {
		return nil, errors.New("not built")
	}
	gblock := b.genesis.ToBlock()
	genesisRLP, err := rlp.EncodeToBytes(gblock)
	if err != nil {
		return nil, err
	}
	bt := &BlockTestJSON{
		Genesis:    headerToBt(gblock.Header()),
		GenesisRLP: genesisRLP,
		Pre:        b.pre,
		Post:       b.post,
		Network:    b.fork,
		SealEngine: "NoProof",
		Config:     btConfig(b.fork, b.config),
		Blocks:     []BtBlockJSON{},
	}
	for _, bb := range b.built {
		enc, err := rlp.EncodeToBytes(bb.block)
		if err != nil {
			return nil, err
		}
		bt.Blocks = append(bt.Blocks, BtBlockJSON{
			BlockHeader:     headerToBt(bb.block.Header()),
			RLP:             enc,
			BlockNumber:     bb.block.Number().String(),
			UncleHeaders:    []*BtHeaderJSON{},
			BlockAccessList: bb.block.AccessList(),
		})
		bt.LastBlock = bb.block.Hash().Hex()
	}
	res := make(BlockchainTest)
	res[name] = bt
	return &res, nil
}

// ToEngineTest returns the test in blockchain_test_engine format, where each
// block is delivered as an engine_newPayload call which includes the block
// access list.
func (b *BtMaker) ToEngineTest(name string) (*EngineTest, error) {
	if b.genesis == nil {
		return nil, errors.New("not built")
	}
	gblock := b.genesis.ToBlock()
	et := &EngineTestJSON{
		Network:  b.fork,
		Genesis:  headerToBt(gblock.Header()),
		Pre:      b.pre,
		Post:     b.post,
		Config:   btConfig(b.fork, b.config),
		Payloads: []BtEnginePayload{},
	}
	for _, bb := range b.built {
		params, err := enginePayloadParams(bb.block, bb.requests)
		if err != nil {
			return nil, err
		}
		np, fcu := engineVersions(b.config, bb.block.Number(), bb.block.Time())
		et.Payloads = append(et.Payloads, BtEnginePayload{
			Params:                   params,
			NewPayloadVersion:        np,
			ForkchoiceUpdatedVersion: fcu,
		})
		et.LastBlock = bb.block.Hash().Hex()
	}
	res := make(EngineTest)
	res[name] = et
	return &res, nil
}

// postStateAccounts returns the set of accounts and slots to include in the
// post state: everything in the pre-state, plus everything touched according
// to the block access lists.
func (b *BtMaker) postStateAccounts() map[common.Address]map[common.Hash]struct{} {
	accounts := make(map[common.Address]map[common.Hash]struct{})
	for addr, acc := range b.pre {
		slots := make(map[common.Hash]struct{})
		for k := range acc.Storage {
			slots[k] = struct{}{}
		}
		accounts[addr] = slots
	}
	for _, bb := range b.built {
		al := bb.block.AccessList()
		if al == nil {
			continue
		}
		for _, acc := range *al {
			slots, ok := accounts[acc.Address]
			if !ok {
				slots = make(map[common.Hash]struct{})
				accounts[acc.Address] = slots
			}
			for _, s := range acc.StorageReads {
				slots[s.Bytes32()] = struct{}{}
			}
			for _, s := range acc.StorageChanges {
				slots[s.Slot.Bytes32()] = struct{}{}
			}
		}
	}
	return accounts
}

// readPost reads the post state from the given state database.
func (b *BtMaker) readPost(statedb *state.StateDB) GenesisAlloc {
	accounts := b.postStateAccounts()
	addrs := make([]common.Address, 0, len(accounts))
	for addr := range accounts {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool { return addrs[i].Cmp(addrs[j]) < 0 })
	post := make(GenesisAlloc)
	for _, addr := range addrs {
		if !statedb.Exist(addr) {
			continue
		}
		acc := GenesisAccount{
			Code:    statedb.GetCode(addr),
			Storage: make(map[common.Hash]common.Hash),
			Balance: statedb.GetBalance(addr).ToBig(),
			Nonce:   statedb.GetNonce(addr),
		}
		if acc.Code == nil {
			acc.Code = []byte{}
		}
		for slot := range accounts[addr] {
			if v := statedb.GetState(addr, slot); v != (common.Hash{}) {
				acc.Storage[slot] = v
			}
		}
		post[addr] = acc
	}
	return post
}
