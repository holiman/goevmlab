package fuzzing

import (
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func fillNaive(gst *GstMaker) {
	// The accounts which we want to be able to invoke
	addrs := []common.Address{
		common.HexToAddress("0xF1"),
	}
	for _, addr := range addrs {
		gst.AddAccount(addr, GenesisAccount{
			Code:    randomBytecode(),
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
		PrivateKey: pKey,
	})
}

// randomBytecode returns a pretty simplistic bytecode, 1024 ops.
func randomBytecode() []byte {
	b := make([]byte, 1024)
	rand.Read(b)
	i := 0
	var next = func() byte {
		x := b[i]
		i++
		if i >= len(b) {
			rand.Read(b)
			i = 0
		}
		return x
	}
	p := program.NewProgram()
	p.Push(0)
	p.Push(1)
	p.Push(1)
	p.Push(2)
	p.Push(2)
	p.Push(500)
	p.Push(0xffff)
	for {
		op := ops.OpCode(next())
		if !ops.IsDefined(op) {
			continue
		}
		p.Op(op)
		if len(p.Bytecode()) > 1024 {
			break
		}
	}
	return p.Bytecode()
}
