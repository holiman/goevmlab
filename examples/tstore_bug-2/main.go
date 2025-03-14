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
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/holiman/goevmlab/fuzzing"
)

func main() {
	if err := makeTest(); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func makeTest() error {
	gst := fuzzing.BasicStateTest("Cancun")

	a := common.HexToAddress("0xaa")
	var bbCode []byte
	// bb is the initcode of the contract to be created
	{
		bb := program.New()

		// Use the value of TLOAD(1) as the size of the returned data.
		// As long as TLOAD(1) is zero, we error out by returning too large initcode
		bb.Push(1)
		bb.Op(vm.TLOAD)
		bb.Push(0x600a)
		bb.Op(vm.SUB) // stack: [ 0x600a - tload(1) ]

		bb.Tstore(1, 0x6000) // do the TSTORE(1), which _should_ be rolled back and thus a no-op.

		bb.Push(0)
		bb.Op(vm.RETURN)
		bbUnaligned := bb.Bytes()
		// Make it align to 32 byte, then the trace is simpler using MSTORE
		// instead of MSTORE8. We just zfill with STOP
		bbCode = make([]byte, 32*((len(bbUnaligned)+31)/32))
		copy(bbCode, bbUnaligned)
	}

	// aa is the code which invokes CREATE2 twice, with bbCode as initcode
	{
		aa := program.New()
		aa.Mstore(bbCode, 0)

		// Call the initcode twice using the CREATE2 opcode
		aa.Push0().Push(len(bbCode)).Push0().Push0().Op(vm.CREATE2)
		aa.Push0().Push(len(bbCode)).Push0().Push0().Op(vm.CREATE2)

		gst.AddAccount(a, fuzzing.GenesisAccount{
			Code:    aa.Bytes(),
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
	t := gst.ToGeneralStateTest("tstore_test-2")
	output, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("tstore_test-2.json", output, 0777)
}
