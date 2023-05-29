package fuzzing

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func fillSstore(gst *GstMaker, fork string) {
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
	nonGenesisAddresses := []common.Address{
		common.HexToAddress("0x00"),
		common.HexToAddress("0x01"),
		common.HexToAddress("0x02"),
		common.HexToAddress("0x03"),
		common.HexToAddress("0x04"),
		common.HexToAddress("0x05"),
		common.HexToAddress("0x06"),
		common.HexToAddress("0x07"),
		common.HexToAddress("0x08"),
		common.HexToAddress("0x09"),
		common.HexToAddress("0x0A"),
		common.HexToAddress("0x0B"),
		common.HexToAddress("0x0C"),
		common.HexToAddress("0x0D"),
		common.HexToAddress("0x0E"),
	}
	var allAddrs []common.Address
	allAddrs = append(allAddrs, addrs...)
	allAddrs = append(allAddrs, nonGenesisAddresses...)
	// make them exist in the state
	for _, addr := range nonGenesisAddresses {
		gst.AddAccount(addr, GenesisAccount{
			Balance: new(big.Int).SetUint64(1),
			Storage: make(map[common.Hash]common.Hash),
		})
	}
	for _, addr := range addrs {
		gst.AddAccount(addr, GenesisAccount{
			Code:    RandCall2200(allAddrs),
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
			GasPrice:   big.NewInt(0x10),
			To:         addrs[0].Hex(),
			Sender:     sender,
			PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		}
		gst.SetTx(tx)
	}
}
