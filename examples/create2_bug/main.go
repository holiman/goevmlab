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

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
	"os"
)

func main() {
	if err := makeTest(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func makeTest() error {
	gst := fuzzing.BasicStateTest("Cancun")

	a := common.HexToAddress("0xaa")
	// This program does a CREATE2 which fails. The CREATE2 can fail for two reasons:
	// 1: it is way to large initcode. This failure exits the current scope.
	// 2: it tries to use too large endowment. This failure fails the create2-op, but
	// does not exit the current scope.
	//
	// The consensus-correct way to fail is 1).
	{
		aa := program.NewProgram()
		aa.Push0()          // gas
		aa.Push0()          // input
		aa.Push0()          //salt
		aa.Push(0x20000)    // size
		aa.Push0()          // offset
		aa.Push(1123123123) // endowment
		aa.Op(ops.CREATE2)
		// Make a mark on the state
		aa.Sstore(1, 1)
		gst.AddAccount(a, fuzzing.GenesisAccount{
			Code:    aa.Bytecode(),
			Balance: big.NewInt(0),
			Storage: make(map[common.Hash]common.Hash),
		})
	}

	// The transaction from sender to a
	{
		fuzzing.AddTransaction(&a, gst)
	}
	traceOut, err := os.Create("expected.trace.jsonl")
	if err != nil {
		return err
	}
	defer traceOut.Close()
	if err := gst.Fill(traceOut); err != nil {
		return err
	}
	t := gst.ToGeneralStateTest("tstore_test-1")
	output, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("create2_test.json", output, 0777)
}
