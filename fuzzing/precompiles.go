package fuzzing

import (
	crand "crypto/rand"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
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

// randSize returns a new random int64
// With 3% probability it outputs 0
// With 92% probability it outputs a number [0..256)
// With 5% probability it outputs a number [0..1024)
func randSize() int64 {
	b := rand.Int31n(100)
	// Zero or not?
	if b < 3 {
		return 0
	}
	if b < 95 {
		return rand.Int63n(257)
	}
	return rand.Int63n(1024)
}

func randCallPrecompile() []byte {
	// fill the memory
	p := program.New()
	size := randSize()
	data := make([]byte, size)
	_, _ = crand.Read(data)
	p.Mstore(data, 0)
	memInFn := func() (offset, size interface{}) {
		return 0, len(data)
	}
	memOutFn := func() (offset, size interface{}) {
		offset, size = 0, 64
		return
	}
	addrGen := func() interface{} {
		return rand.Uint32() % 18
	}
	p2 := RandCall(GasRandomizer(), addrGen, ValueRandomizer(), memInFn, memOutFn)
	p.Append(p2)
	// store the returnvalue ot slot 1337
	p.Push(0x1337)
	p.Op(vm.SSTORE)
	// Store the output in some slot, to make sure the stateroot changes
	p.MemToStorage(0, 64, 0)
	return p.Bytes()
}
