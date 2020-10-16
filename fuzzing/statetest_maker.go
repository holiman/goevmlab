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
	"io/ioutil"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/core/state"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/tests"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

// The sender
var sender = common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")

var randomAddresses = []common.Address{
	// Some random accounts
	common.HexToAddress("ffffffffffffffffffffffffffffffffffffffff"),
	common.HexToAddress("1000000000000000000000000000000000000000"),
	common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
	common.HexToAddress("c94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
	common.HexToAddress("d94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
}

var precompiles = []common.Address{
	// Some precompiles
	common.HexToAddress("0000000000000000000000000000000000000001"),
	common.HexToAddress("0000000000000000000000000000000000000002"),
	common.HexToAddress("0000000000000000000000000000000000000003"),
	common.HexToAddress("0000000000000000000000000000000000000004"),
	common.HexToAddress("0000000000000000000000000000000000000005"),
	common.HexToAddress("0000000000000000000000000000000000000006"),
	common.HexToAddress("0000000000000000000000000000000000000007"),
	common.HexToAddress("0000000000000000000000000000000000000008"),
	common.HexToAddress("0000000000000000000000000000000000000005"),
	common.HexToAddress("0000000000000000000000000000000000000006"),
	common.HexToAddress("0000000000000000000000000000000000000007"),
	common.HexToAddress("0000000000000000000000000000000000000008"),
	common.HexToAddress("0000000000000000000000000000000000000009"),
}

var allAddresses []common.Address

// We don't use all opcodes, only
// - valid opcodes,
// - not all push,
//   - only push1, push2 and push20
var usedOpCodes []ops.OpCode

func init() {
	allAddresses = append(allAddresses, randomAddresses...)
	allAddresses = append(allAddresses, precompiles...)
	usedOpCodes = ops.ValidOpcodes
	for _, op := range ops.ValidOpcodes {
		if op > ops.PUSH2 || op <= ops.PUSH19 {
			continue
		}
		if op >= ops.PUSH21 || op <= ops.PUSH32 {
			continue
		}
		usedOpCodes = append(usedOpCodes, op)
	}
}

//// randCall creates a random call-type
//func randCall(p *program.Program, op ops.OpCode) {
//	addrGen := addressRandomizer(allAddresses)
//	memGen := randInt(0x70, 0xef)
//
//	p.Push(memGen()) //mem out size
//	p.Push(memGen()) // mem out start
//	p.Push(memGen()) //mem in size
//	p.Push(memGen()) // mem in start
//	switch op {
//	case ops.CALL, ops.CALLCODE:
//		p.Push(rand.Intn(256)) //value
//	}
//	addr := addrGen()
//	p.Push(addr)
//	p.Op(ops.GAS)
//	p.Op(op)
//}
//// randProgram creates a random program
//func randProgram(maxSize int) []byte {
//	var p = program.NewProgram()
//	stack := 0
//	var numOps = len(usedOpCodes)
//	for {
//		op := usedOpCodes[rand.Intn(numOps)]
//		if op.IsCall() {
//			randCall(p, op)
//			stack += 1
//		} else {
//			if stack-len(op.Pops()) < 0 {
//				continue
//			}
//			stack -= len(op.Pops())
//			stack += len(op.Pushes())
//			p.Op(op)
//		}
//		if p.Size() > maxSize {
//			break
//		}
//	}
//	return p.Bytecode()
//}

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
	gst := &GstMaker{
		env: &stEnv{
			// The ENV portion
			Number:     1,
			GasLimit:   0x26e1f476fe1e22,
			Difficulty: big.NewInt(0x20000),
			Coinbase:   common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b"),
			Timestamp:  0x03e8,
		},
		pre: &alloc,
	}
	return gst
}

func (g *GstMaker) AddAccount(address common.Address, a GenesisAccount) {
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

// randomFillGenesisAlloc fills the state with some random data
// and returns a destination account which has code
func (g *GstMaker) randomFillGenesisAlloc() *common.Address {
	// Add at least one that we can invoke
	nAccounts := 1 + rand.Intn(len(randomAddresses)-1)
	var dest *common.Address
	for i := 0; i < nAccounts; i++ {
		code := RandCallBlake()
		address := randomAddresses[i]
		if dest == nil {
			dest = &address
		}
		g.AddAccount(address, GenesisAccount{
			Nonce:   uint64(rand.Intn(500)),
			Balance: big.NewInt(int64(rand.Intn(500000))),
			Code:    code,
			Storage: make(map[common.Hash]common.Hash),
		})
	}
	return dest
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
		//postHash := common.HexToHash("0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134")
		//logsHash := common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")
		postState := make(map[string][]stPostState)
		postState[fork] = []stPostState{
			stPostState{
				Logs:    common.UnprefixedHash(g.logs),
				Root:    common.UnprefixedHash(g.root),
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

	data, err := ioutil.ReadFile(name)
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

// FillTest uses go-ethereum internally to determine the state root and logs, and optionally
// outputs the trace to the given writer (if non-nil)
func (g *GstMaker) Fill(traceOutput io.Writer) error {

	test, err := g.ToStateTest()
	if err != nil {
		return err
	}
	subtest := test.Subtests()[0]
	cfg := vm.Config{}
	if traceOutput != nil {
		cfg.Debug = true
		cfg.Tracer = vm.NewJSONLogger(&vm.LogConfig{}, traceOutput)
	}
	_, statedb, root, err := test.RunNoVerify(subtest, cfg, false)
	if err != nil {
		return err
	}

	logs := rlpHash(statedb.Logs())
	g.SetResult(root, logs)
	return nil
}

func BasicStateTest(fork string) *GstMaker {
	gst := NewGstMaker()
	// Add sender
	gst.AddAccount(sender, GenesisAccount{
		Nonce:   0,
		Balance: big.NewInt(0xffffffff),
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
		GasPrice:   big.NewInt(0x01),
		PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
	}
	gst.SetTx(tx)
}

// GenerateStateTest generates a random state tests
func GenerateStateTest(name string) *GeneralStateTest {
	gst := BasicStateTest("Istanbul")
	// add some random accounts
	dest := gst.randomFillGenesisAlloc()
	// The transaction
	{
		tx := &StTransaction{
			// 8M gaslimit
			GasLimit:   []uint64{8000000},
			Nonce:      0,
			Value:      []string{randHex(4)},
			Data:       []string{randHex(100)},
			GasPrice:   big.NewInt(0x01),
			PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		}
		if dest != nil {
			tx.To = dest.Hex()
		}
		gst.SetTx(tx)
	}
	return gst.ToGeneralStateTest(name)
}

func GenerateBlake() *GstMaker {
	gst := BasicStateTest("Istanbul")
	// Add a contract which calls blake
	dest := common.HexToAddress("0x0000ca1100b1a7e")
	gst.AddAccount(dest, GenesisAccount{
		Code:    RandCallBlake(),
		Balance: new(big.Int),
		Storage: make(map[common.Hash]common.Hash),
	})
	// The transaction
	{
		tx := &StTransaction{
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

// GenerateBlakeTest generates a random test of the blake F precompile
func GenerateBlakeTest(name string) *GeneralStateTest {
	gst := GenerateBlake()
	return gst.ToGeneralStateTest(name)
}

func GenerateSubroutineTest() *GstMaker {
	gst := BasicStateTest("Berlin")
	// The accounts which we want to be able to invoke
	addrs := []common.Address{
		common.HexToAddress("0xF1"),
		common.HexToAddress("0xF2"),
		common.HexToAddress("0xF3"),
		common.HexToAddress("0xF4"),
		common.HexToAddress("0xF5"),
		common.HexToAddress("0xF6"),
		common.HexToAddress("0xF7"),
		common.HexToAddress("0xF8"),
		common.HexToAddress("0xF9"),
		common.HexToAddress("0xFA"),
	}
	//addrGen := addressRandomizer(addrs)
	for _, addr := range addrs {
		gst.AddAccount(addr, GenesisAccount{
			Code:    RandCallSubroutine(addrs),
			Balance: new(big.Int),
			Storage: make(state.Storage),
		})
	}
	// The transaction
	{
		tx := &StTransaction{
			// 8M gaslimit
			GasLimit:   []uint64{8000000},
			Nonce:      0,
			Value:      []string{randHex(4)},
			Data:       []string{randHex(100)},
			GasPrice:   big.NewInt(0x01),
			To:         addrs[0].Hex(),
			PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		}
		gst.SetTx(tx)
	}
	return gst
}

func Generate2200BerlinTest() *GstMaker {
	gst := BasicStateTest("Berlin")
	create2200Test(gst)
	return gst
}

func Generate2200Test() *GstMaker {
	gst := BasicStateTest("Istanbul")
	create2200Test(gst)
	return gst
}

func create2200Test(gst *GstMaker) {
	// The accounts which we want to be able to invoke
	addrs := []common.Address{
		common.HexToAddress("0xF1"),
		common.HexToAddress("0xF2"),
		common.HexToAddress("0xF3"),
		common.HexToAddress("0xF4"),
		common.HexToAddress("0xF5"),
		common.HexToAddress("0xF6"),
		common.HexToAddress("0xF7"),
		common.HexToAddress("0xF8"),
		common.HexToAddress("0xF9"),
		common.HexToAddress("0xFA"),
	}
	//addrGen := addressRandomizer(addrs)
	for _, addr := range addrs {
		gst.AddAccount(addr, GenesisAccount{
			Code:    RandCall2200(addrs),
			Balance: new(big.Int),
			Storage: RandStorage(15, 20),
		})
	}
	// The transaction
	{
		tx := &StTransaction{
			// 8M gaslimit
			GasLimit:   []uint64{8000000},
			Nonce:      0,
			Value:      []string{randHex(4)},
			Data:       []string{randHex(100)},
			GasPrice:   big.NewInt(0x01),
			To:         addrs[0].Hex(),
			PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		}
		gst.SetTx(tx)
	}
}

func GenerateECRecover() (*GstMaker, []byte) {
	gst := BasicStateTest("Istanbul")
	// Add a contract which calls BLS
	dest := common.HexToAddress("0x00ca11ec5ec04e5")
	code := RandCallECRecover()
	gst.AddAccount(dest, GenesisAccount{
		Code:    code,
		Balance: new(big.Int),
		Storage: make(map[common.Hash]common.Hash),
	})
	// The transaction
	{
		tx := &StTransaction{
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

func RandCallECRecover() []byte {
	p := program.NewProgram()
	offset := 0
	rounds := rand.Int31n(10000)
	for i := int32(0); i < rounds; i++ {
		data := make([]byte, 128)
		rand.Read(data)
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
