package fuzzing

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
	"math/big"
	"math/rand"
	"time"
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
// valid opcodes,
// not all push, only push1, push2 and push20
var usedOpCodes []ops.OpCode

func init() {
	for _, a := range randomAddresses {
		allAddresses = append(allAddresses, a)
	}
	for _, a := range precompiles {
		allAddresses = append(allAddresses, a)
	}
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

func getGenesisAlloc() (*GenesisAlloc, *common.Address) {
	alloc := make(GenesisAlloc)
	alloc[sender] = GenesisAccount{
		Nonce:   0,
		Balance: big.NewInt(0xffffffff),
		Storage: make(map[common.Hash]common.Hash),
		Code:    []byte{},
	}
	// Add at least one that we can invoke
	nAccounts := 1 + rand.Intn(len(randomAddresses)-1)
	var dest *common.Address
	for i := 0; i < nAccounts; i++ {
		code := RandCallBlake()
		address := randomAddresses[i]
		if dest == nil {
			dest = &address
		}
		alloc[address] = GenesisAccount{
			Nonce:   uint64(rand.Intn(500)),
			Balance: big.NewInt(int64(rand.Intn(500000))),
			Code:    code,
			Storage: make(map[common.Hash]common.Hash),
		}
		break
	}

	return &alloc, dest
}

func randAddress() common.Address {
	n := rand.Intn(len(allAddresses))
	return allAddresses[n]
}

func randCall(p *program.Program, op ops.OpCode) {

	p.Push(randSmallMemLength()) //mem out size
	p.Push(randSmallMemLength()) // mem out start
	p.Push(randSmallMemLength()) //mem in size
	p.Push(randSmallMemLength()) // mem in start

	switch op {
	case ops.CALL, ops.CALLCODE:
		p.Push(rand.Intn(256)) //value
	}
	addr := randAddress()
	p.Push(addr)
	p.Op(ops.GAS)
	p.Op(op)
}

func randSmallMemLength() uint64 {
	return uint64(rand.Intn(500))
}

func randProgram(maxSize int) []byte {
	var p = program.NewProgram()
	stack := 0
	var numOps = len(usedOpCodes)
	for {
		op := usedOpCodes[rand.Intn(numOps)]
		if op.IsCall() {
			randCall(p, op)
			stack += 1
		} else {
			if stack-len(op.Pops()) < 0 {
				continue
			}
			stack -= len(op.Pops())
			stack += len(op.Pushes())
			p.Op(op)
		}
		if p.Size() > maxSize {
			break
		}
	}
	return p.Bytecode()
}

func randHex(maxSize int) string {
	size := rand.Intn(maxSize)
	b := make([]byte, size)
	rand.Read(b)
	return hexutil.Encode(b)
}
func init() {
	rand.Seed(time.Now().Unix())
}

func GenerateStateTest(name string) *GeneralStateTest {
	st := &stJSON{}
	// The alloc portion
	alloc, dest := getGenesisAlloc()
	st.Pre = *alloc
	// The ENV portion
	{
		st.Env.Number = 1
		st.Env.GasLimit = 0x26e1f476fe1e22
		st.Env.Difficulty = big.NewInt(0x20000)
		st.Env.Coinbase = common.HexToAddress("b94f5374fce5edbc8e2a8697c15331677e6ebf0b")
		st.Env.Timestamp = 0x03e8
	}
	// The POST portion
	{
		postHash := common.HexToHash("0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134")
		logsHash := common.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347")
		postState := make(map[string][]stPostState)
		postState["Istanbul"] = []stPostState{
			stPostState{
				Logs:    common.UnprefixedHash(logsHash),
				Root:    common.UnprefixedHash(postHash),
				Indexes: stIndex{Gas: 0, Value: 0, Data: 0},
			},
		}
		st.Post = postState
	}
	// The transaction
	{
		st.Tx = stTransaction{
			// 3M gaslimit
			GasLimit:   []uint64{3000000},
			Nonce:      0,
			Value:      []string{randHex(4)},
			Data:       []string{randHex(100)},
			GasPrice:   big.NewInt(0x01),
			PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		}
		if dest != nil {
			st.Tx.To = dest.Hex()
		}

	}
	gst := make(GeneralStateTest)
	gst[name] = st
	return &gst

}
