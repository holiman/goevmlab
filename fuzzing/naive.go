package fuzzing

import (
	"crypto/rand"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func fillNaive(gst *GstMaker, fork string) {
	// The accounts which we want to be able to invoke
	addrs := []common.Address{
		common.HexToAddress("0xF1"),
	}
	forkDef := ops.LookupFork(fork)
	if forkDef == nil {
		panic("bad fork")
	}

	for _, addr := range addrs {
		gst.AddAccount(addr, GenesisAccount{
			Code:    randomBytecode(forkDef),
			Balance: new(big.Int),
			Storage: RandStorage(15, 20),
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
		Sender:     sender,
		PrivateKey: pKey,
	})
}

// randomBytecode returns a pretty simplistic bytecode, 1024 ops.
func randomBytecode(f *ops.Fork) []byte {
	b := make([]byte, 1024)
	_, _ = rand.Read(b)
	i := 0
	var next = func() byte {
		x := b[i]
		i++
		if i >= len(b) {
			_, _ = rand.Read(b)
			i = 0
		}
		return x
	}
	p := program.NewProgram()
	p.Push(next())
	p.Push(next())
	p.Push(next())
	p.Push(next())
	p.Push(next())
	p.Push(next())
	p.Push(next())
	for p.Size() < 1024 {
		p.Op(f.RandomOp(next()))
	}
	return p.Bytecode()
}
