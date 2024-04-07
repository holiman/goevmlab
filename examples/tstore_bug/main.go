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
	b := common.HexToAddress("0xbb")

	// 0xaa calls 0xbb, then checks TLOAD(4) and puts it into state
	{
		aa := program.NewProgram()
		aa.CallCode(nil, b, nil, 0, 0, 0, 0)

		aa.Push(4)
		aa.Op(ops.TLOAD)

		aa.Push(1)
		aa.Op(ops.SSTORE)

		gst.AddAccount(a, fuzzing.GenesisAccount{
			Code:    aa.Bytecode(),
			Balance: big.NewInt(0),
			Storage: make(map[common.Hash]common.Hash),
		})
	}

	// 0xbb does a TSTORE, then exits on revert
	{

		bb := program.NewProgram()
		bb.Tstore(4, 1)

		// Now call a precompile

		bb.Call(nil, common.HexToAddress("0x6"), nil, 0, 0, 0, 0)

		bb.Push0()
		bb.Push0()
		bb.Op(ops.REVERT) // Now exit with error
		gst.AddAccount(b, fuzzing.GenesisAccount{
			Code:    bb.Bytecode(),
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
	gst.Fill(traceOut)
	t := gst.ToGeneralStateTest("tstore_test-1")
	output, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("tstore_test-1.json", output, 0777)
}
