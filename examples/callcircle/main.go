// Copyright 2025 Martin Holst Swende
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

package main

import (
	"fmt"
	"math/big"
	"os"
	"time"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
	common2 "github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	program2 "github.com/holiman/goevmlab/utils"
)

func smallCallCode(p *program.Program, address any) *program.Program {
	return p.
		Push0(). //outsize
		Push0(). // outoffset
		Push0(). //insize
		Push0(). //inoffset
		Push0(). //value
		Push(address).
		Op(vm.GAS).
		Op(vm.CALLCODE)
}

func clearCallCode(p *program.Program) *program.Program {
	return p.
		Push0(). //outsize
		Push0(). // outoffset
		Push0(). //insize
		Push0(). //inoffset
		Push0(). //value
		Push0(). // address
		Op(vm.GAS).
		Op(vm.CALLCODE)
}

// Pads the given data to 24576 (0x6000) size
func pad(data []byte) []byte {
	// ah fuhgedaboudit
	return data
	//d := make([]byte, 24_576)
	//copy(d, data)
	//return d
}

func main() {
	if err := program2.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runit() error {
	alloc := make(types.GenesisAlloc)
	aAddr := common.HexToAddress("0xf1")
	bAddr := common.HexToAddress("0xf2")
	cAddr := common.HexToAddress("0xf3")
	sender := common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")

	alloc[sender] = types.Account{
		Balance: big.NewInt(1000000000000000000), // 1 eth
	}

	{
		p := program.New()
		p = smallCallCode(p, 0xf2)
		p = clearCallCode(p)
		p = smallCallCode(p, 0xf2)
		p = smallCallCode(p, 0xf2)
		alloc[aAddr] = types.Account{
			Code:    pad(p.Bytes()),
			Balance: big.NewInt(0xffffffff),
		}
	}
	{
		p := program.New()
		p = smallCallCode(p, 0xf3)
		p = clearCallCode(p)
		alloc[bAddr] = types.Account{
			Code:    pad(p.Bytes()),
			Balance: big.NewInt(0xffffffff),
		}
	}
	{
		p := program.New()
		p = smallCallCode(p, 0xf1)
		p = clearCallCode(p)
		alloc[cAddr] = types.Account{
			Code:    pad(p.Bytes()),
			Balance: big.NewInt(0xffffffff),
		}
	}
	var (
		statedb = common2.StateDBWithAlloc(alloc)
	)
	statedb.CreateAccount(sender)
	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    100_000_000,
		Difficulty:  big.NewInt(0x200000),
		BlockNumber: new(big.Int).SetUint64(1),
		ChainConfig: &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      new(big.Int),
			EIP150Block:         new(big.Int),
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
			MuirGlacierBlock:    new(big.Int),
			BerlinBlock:         new(big.Int),
			LondonBlock:         new(big.Int),
			ArrowGlacierBlock:   new(big.Int),
			GrayGlacierBlock:    new(big.Int),
			MergeNetsplitBlock:  new(big.Int),
			ShanghaiTime:        new(uint64),
			CancunTime:          new(uint64),
			PragueTime:          new(uint64),
		},
		EVMConfig: vm.Config{
			Tracer: new(customTracer).Hooks(),
		},
	}
	// Run with tracing
	_, _, _ = runtime.Call(aAddr, nil, &runtimeConfig)
	// Diagnose it
	runtimeConfig.EVMConfig = vm.Config{}

	for i := 0; i < 5; i++ {
		t0 := time.Now()
		_, _, err := runtime.Call(aAddr, nil, &runtimeConfig)
		t1 := time.Since(t0)
		fmt.Fprintf(os.Stderr, "Time elapsed: %v, err: %v\n", t1, err)
	}

	// Turn it into a statetest
	mkr := fuzzing.BasicStateTest("Prague")
	// convert the genesisAlloc
	var fuzzGenesisAlloc = make(fuzzing.GenesisAlloc)
	for k, v := range alloc {
		fuzzAcc := fuzzing.GenesisAccount{
			Code:       v.Code,
			Storage:    v.Storage,
			Balance:    v.Balance,
			Nonce:      v.Nonce,
			PrivateKey: v.PrivateKey,
		}
		if fuzzAcc.Storage == nil {
			fuzzAcc.Storage = make(map[common.Hash]common.Hash)
		}
		fuzzGenesisAlloc[k] = fuzzAcc
	}

	tx := &fuzzing.StTransaction{
		GasLimit:   []uint64{100_000_000},
		Nonce:      0,
		Value:      []string{"0x0"},
		Data:       []string{""},
		GasPrice:   big.NewInt(0x10),
		PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		To:         aAddr.Hex(),
		Sender:     sender,
	}
	mkr.SetTx(tx)

	mkr.SetPre(&fuzzGenesisAlloc)
	if err := mkr.Fill(nil); err != nil {
		return err
	}
	gst := mkr.ToGeneralStateTest("CallCodeCircle")
	dat, _ := json.MarshalIndent(gst, "", " ")
	fmt.Println(string(dat))
	return nil
}

type customTracer struct{}

func (d *customTracer) Hooks() *tracing.Hooks {
	count := 0
	return &tracing.Hooks{
		OnOpcode: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
			count++
		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			fmt.Fprintf(os.Stderr, "CaptureFault %v\n", err)
		},
		OnTxStart: func(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
		},
		OnTxEnd: func(receipt *types.Receipt, err error) {
			fmt.Fprintf(os.Stderr, "Execution lasted %d steps, gasUsed: %v\n", count, receipt.GasUsed)
		},
	}
}
