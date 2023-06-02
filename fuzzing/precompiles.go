package fuzzing

import (
	crand "crypto/rand"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func fillPrecompileTest(gst *GstMaker, fork string) {
	// Add a contract which calls a precompile
	dest := common.HexToAddress("0x0000ca1100b1a7e")
	gst.AddAccount(dest, GenesisAccount{
		Code:    randCallPrecompile(),
		Balance: new(big.Int),
		Storage: make(map[common.Hash]common.Hash),
	})
	// The transaction
	gst.SetTx(&StTransaction{
		// 8M gaslimit
		GasLimit:   []uint64{8000000},
		Nonce:      0,
		Value:      []string{randHex(4)},
		Data:       []string{""},
		GasPrice:   big.NewInt(0x20),
		To:         dest.Hex(),
		Sender:     sender,
		PrivateKey: pKey,
	})
}

func randCallPrecompile() []byte {
	// fill the memory
	p := program.NewProgram()
	data := make([]byte, 1024)
	_, _ = crand.Read(data)
	p.Mstore(data, 0)
	memInFn := func() (offset, size interface{}) {
		offset, size = 0, rand.Uint32()%uint32(len(data))
		return
	}
	memOutFn := func() (offset, size interface{}) {
		offset, size = 0, 64
		return
	}
	addrGen := func() interface{} {
		return rand.Uint32() % 18
	}
	p2 := RandCall(GasRandomizer(), addrGen, ValueRandomizer(), memInFn, memOutFn)
	p.AddAll(p2)
	// store the returnvalue ot slot 1337
	p.Push(0x1337)
	p.Op(ops.SSTORE)
	// Store the output in some slot, to make sure the stateroot changes
	p.MemToStorage(0, 64, 0)
	return p.Bytecode()
}
