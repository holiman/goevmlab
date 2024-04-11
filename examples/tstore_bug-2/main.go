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

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func main() {
	if err := makeTest(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func makeTest() error {
	gst := fuzzing.BasicStateTest("Cancun")

	a := common.HexToAddress("0xaa")

	// bb is the initcode of the contract to be created
	{
		bb := program.NewProgram()

		// Calculate code length based on the output of TLOAD
		// If TLOAD results in 0, the code length is 10
		// For anything else, it is too large
		bb.Push(1)
		bb.Op(ops.TLOAD)

		bb.Push(1)
		bb.Push(1)
		bb.Op(ops.TSTORE)

		bb.Op(ops.ISZERO)
		bb.Push(24576) // MAX_CODE_SIZE 0x6000
		bb.Op(ops.MUL)
		bb.Push(10)
		bb.Op(ops.ADD)
		bb.Push(0)

		bb.Op(ops.RETURN)

		// Repeat the whole thing above one more time
		aa := program.NewProgram()
		aa.Mstore(bb.Bytecode(), 0)

		// Call the initcode twice using the CREATE2 opcode
		aa.Push(0).Push(len(bb.Bytecode())).Push(0).Push(0).Op(ops.CREATE2)
		aa.Push(0).Push(len(bb.Bytecode())).Push(0).Push(0).Op(ops.CREATE2)

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
	gst.Fill(traceOut)
	t := gst.ToGeneralStateTest("tstore_test-2")
	output, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("tstore_test-2.json", output, 0777)
}
