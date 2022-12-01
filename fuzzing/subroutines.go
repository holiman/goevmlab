package fuzzing

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
)

// FactorySubroutine returns a 'factory' which creates EIP-2315-tests on the given fork.
func FactorySubroutine(fork string) func() *GstMaker {
	return func() *GstMaker {
		gst := BasicStateTest(fork)
		fillSubroutineTest(gst)
		return gst
	}
}

func fillSubroutineTest(gst *GstMaker) {
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
	for _, addr := range addrs {
		gst.AddAccount(addr, GenesisAccount{
			Code:    RandCallSubroutine(addrs),
			Balance: new(big.Int),
			Storage: make(state.Storage),
		})
	}
	// The transaction
	gst.SetTx(&StTransaction{
		// 8M gaslimit
		GasLimit:   []uint64{8000000},
		Nonce:      0,
		Value:      []string{randHex(4)},
		Data:       []string{randHex(100)},
		GasPrice:   big.NewInt(0x10),
		To:         addrs[0].Hex(),
		PrivateKey: pKey,
	})
}
