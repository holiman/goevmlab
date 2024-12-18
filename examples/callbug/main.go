package main

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/uint256"
)

func main() {
	t := makeTest()
	out, _ := json.MarshalIndent(t, "", "  ")
	fmt.Println(string(out))
}

func makeTest() *fuzzing.GeneralStateTest {
	gst := fuzzing.BasicStateTest("Berlin")

	a := common.HexToAddress("0xaa")
	b := common.HexToAddress("0xbb")
	// 0xaa calls 0xbb, with exactly 0x2cef gas
	{
		aa := program.New().Call(uint256.NewInt(0x2cef+21+36), b, 0, 0, 0, 0, 0)
		gst.AddAccount(a, fuzzing.GenesisAccount{
			Code:    aa.Bytes(),
			Balance: big.NewInt(5),
			Storage: make(map[common.Hash]common.Hash),
		})
	}
	// 0xbb does a call
	{
		// expand memory
		bb := program.New().
			Push(1).     // push the value
			Push(0x100). // push the memory index
			Op(vm.MSTORE)
		gas := new(uint256.Int)
		if err := gas.SetFromHex("0x7ef0367e633852132a0ebbf70eb714015dd44bc82e1e55a96ef1389c999c1bca"); err != nil {
			panic(err)
		}
		// {"pc":21090,"op":241,"gas":"0x2cef","gasCost":"0x2cd3","memSize":0,"stack":["0x100","0x0","0x60","0x0","0x4b","0x11","0x7ef0367e633852132a0ebbf70eb714015dd44bc82e1e55a96ef1389c999c1bca"],
		bb.Call(gas, common.HexToAddress("0x11"), 0x4b, 0x0, 0x60, 0x0, 0x100)
		bb.Op(vm.POP)
		bb.Return(0, 0)
		gst.AddAccount(b, fuzzing.GenesisAccount{
			Code:    bb.Bytes(),
			Balance: big.NewInt(5),
			Storage: make(map[common.Hash]common.Hash),
		})
	}

	// The transaction from sender to a
	{
		fuzzing.AddTransaction(&a, gst)
	}
	return gst.ToGeneralStateTest("nethermind-caller")

}
