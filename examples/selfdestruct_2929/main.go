// Copyright 2019 Martin Holst Swende, Hubert Ritzdorf
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
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/params"
	common2 "github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	program2 "github.com/holiman/goevmlab/utils"
	"github.com/holiman/uint256"
)

// This program creates a testcase surrounding selfdestruct in the context of
// EIP-2929.
// Selfdestruct(beneficiary) costs:
// If the ETH recipient of a SELFDESTRUCT is not in accessed_addresses
// (regardless of whether or not the amount sent is nonzero), charge an
// additional COLD_ACCOUNT_ACCESS_COST on top of the existing gas costs,
// and add the ETH recipient to the set.
func main() {

	if err := program2.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runit() error {

	// The destructor, caller and sender
	destAddr := common.HexToAddress("0x000000000000000000000000000000000000dead")
	callerAddr := common.HexToAddress("0x000000000000000000000000000000000000c411")
	var sender = common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")

	var existingAddresses = []common.Address{
		common.HexToAddress("0x00000000000000000000000000000000000000AA"),
		common.HexToAddress("0x00000000000000000000000000000000000000CC"),
	}
	var warmAddresses = []common.Address{
		common.HexToAddress("0x00000000000000000000000000000000000000CC"),
		common.HexToAddress("0x00000000000000000000000000000000000000DD"),
		common.HexToAddress("0x0000000000000000000000000000000000000003"),
	}
	var addresses = []common.Address{
		// This one is 'cold', but exists. Load it twice
		common.HexToAddress("0x00000000000000000000000000000000000000AA"),
		common.HexToAddress("0x00000000000000000000000000000000000000AA"),
		// This one is 'cold', and does not exist. Load it twice
		common.HexToAddress("0x00000000000000000000000000000000000000BB"),
		common.HexToAddress("0x00000000000000000000000000000000000000BB"),
		// This one is 'hot', and exists. Load it twice
		common.HexToAddress("0x00000000000000000000000000000000000000CC"),
		common.HexToAddress("0x00000000000000000000000000000000000000CC"),
		// This one is 'hot', and does not exist. Load it twice
		common.HexToAddress("0x00000000000000000000000000000000000000DD"),
		common.HexToAddress("0x00000000000000000000000000000000000000DD"),
		// This one is a precompile, thus 'hot'. Test it twice
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		common.HexToAddress("0x0000000000000000000000000000000000000001"),
		// This one is a precompile, thus 'hot', but also touched. Test it twice
		common.HexToAddress("0x0000000000000000000000000000000000000002"),
		common.HexToAddress("0x0000000000000000000000000000000000000002"),
		// Another precompile, test it once
		common.HexToAddress("0x0000000000000000000000000000000000000003"),
	}

	// The destructor contract selfdestructs to the beneficiary given in the input data
	destructor := program.New()
	destructor.InputAddressToStack(0)
	destructor.Op(vm.SELFDESTRUCT)

	// The caller contract calls the destructor repeatedly
	caller := program.New()

	// First of all, 'touch' the hot ones
	for _, addr := range warmAddresses {
		caller.Call(uint256.NewInt(0), addr, 0, 0, 0, 0, 0) // Touch it
		caller.Op(vm.POP)                                   // Ignore returnvalue
	}

	for _, addr := range addresses {
		paddedAddr := make([]byte, 32)
		copy(paddedAddr[12:], addr.Bytes())
		caller.Mstore(paddedAddr, 0)               // Load address to memory
		caller.Call(nil, destAddr, 0, 0, 32, 0, 0) // Call destructor
		caller.Op(vm.POP)                          // Ignore returnvalue
	}
	// Set up a genesis
	alloc := make(types.GenesisAlloc)
	// Create those that are supposed to exist
	for _, addr := range existingAddresses {
		alloc[addr] = types.Account{
			Nonce:   1,
			Balance: big.NewInt(0),
		}
	}

	alloc[destAddr] = types.Account{
		Nonce:   1,
		Code:    destructor.Bytes(),
		Balance: big.NewInt(0x1),
	}
	alloc[callerAddr] = types.Account{
		Nonce:   1,
		Code:    caller.Bytes(),
		Balance: big.NewInt(0x1),
	}
	alloc[sender] = types.Account{
		Nonce:   0,
		Balance: big.NewInt(1000000000000000000), // 1 eth
	}

	//-------------

	outp, err := json.MarshalIndent(alloc, "", " ")
	if err != nil {
		fmt.Printf("error : %v", err)
		os.Exit(1)
	}
	fmt.Printf("output \n%v\n", string(outp))
	//----------
	var (
		statedb = common2.StateDBWithAlloc(alloc)
	)
	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    10000000,
		Difficulty:  big.NewInt(0x200000),
		BlockNumber: new(big.Int).SetUint64(1),
		ChainConfig: &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			EIP150Block:         new(big.Int),
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
			BerlinBlock:         new(big.Int),
		},
		EVMConfig: vm.Config{
			Tracer: logger.NewMarkdownLogger(nil, os.Stdout).Hooks(),
		},
	}
	// Run with tracing
	_, _, _ = runtime.Call(callerAddr, nil, &runtimeConfig)
	// Diagnose it
	runtimeConfig.EVMConfig = vm.Config{}
	t0 := time.Now()
	_, _, err = runtime.Call(callerAddr, nil, &runtimeConfig)
	t1 := time.Since(t0)
	fmt.Printf("Time elapsed: %v\n", t1)
	// Turn it into a statetest
	mkr := fuzzing.BasicStateTest("Berlin")
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
		// 8M gaslimit
		GasLimit:   []uint64{8000000},
		Nonce:      0,
		Value:      []string{"0x0"},
		Data:       []string{""},
		GasPrice:   big.NewInt(0x01),
		PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		To:         callerAddr.Hex(),
	}
	mkr.SetTx(tx)

	mkr.SetPre(&fuzzGenesisAlloc)
	if err := mkr.Fill(os.Stdout); err != nil {
		return err
	}

	gst := mkr.ToGeneralStateTest("TestSelfdestruct")
	dat, _ := json.MarshalIndent(gst, "", " ")
	fmt.Println(string(dat))
	return err
}
