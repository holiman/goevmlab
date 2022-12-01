package fuzzing

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var pKey = hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8")

// FactoryBlake returns a 'factory' which creates blake-precompile tests on the given fork.
func FactoryBlake(fork string) func() *GstMaker {
	return func() *GstMaker {
		gst := BasicStateTest(fork)
		fillBlake(gst)
		return gst
	}
}

func fillBlake(gst *GstMaker) {
	// Add a contract which calls blake
	dest := common.HexToAddress("0x0000ca1100b1a7e")
	gst.AddAccount(dest, GenesisAccount{
		Code:    RandCallBlake(),
		Balance: new(big.Int),
		Storage: make(map[common.Hash]common.Hash),
	})
	// The transaction
	gst.SetTx(&StTransaction{
		// 8M gaslimit
		GasLimit:   []uint64{8000000},
		Value:      []string{randHex(4)},
		Data:       []string{randHex(100)},
		GasPrice:   big.NewInt(0x10),
		To:         dest.Hex(),
		PrivateKey: pKey,
	})
}
