// Copyright 2019 Martin Holst Swende
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

package ops

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/core/vm"
)

// TestSanity checks the npops and npushes against the
// go-ethereum codebase
func TestSanity(t *testing.T) {

	for i := 0; i < 256; i++ {

		// We have the EOF opcodes defined, geth doesn't yet.
		switch i {
		case 0x5c, 0x5d, 0x5e, 0xb0, 0xb1, 0xb3, 0xb4:
			continue

		}
		// Lookup the name via opcode
		gethOp := vm.OpCode(byte(i))
		ourOp := OpCode(byte(i))
		{
			exp, got := gethOp.String(), ourOp.String()
			if exp != got {
				t.Errorf("op 0x%x, got %v expected %v", i, got, exp)
			}
		}
		// Lookup opcode via name
		if name := ourOp.String(); !strings.HasPrefix(name, "opcode") {
			our := byte(StringToOp(name))
			if byte(our) != byte(i) {
				t.Errorf("name %v, got 0x%x expected 0x%x", name, our, byte(i))
			}
			geth := byte(vm.StringToOp(name))
			if byte(geth) != byte(i) {
				t.Errorf("name %v, got 0x%x expected 0x%x", name, geth, byte(i))
			}
		}
		// Check stack pushes and pops
		{
			// This check can only be executed if the go-ethereum codebase
			// is refactored a bit, to make the following public:
			// - vm.LatestInstructionSet pointing to latest instruction set
			// - vm.operation.MinStack
			// - vm.operation.MaxStack
			// Was tested on 2021-12-15, oll korrekt

			/*
				gotPops := len(ourOp.Pops())
				geth_instr := vm.LatestInstructionset[gethOp]
				if gotPops != geth_instr.MinStack {
					t.Errorf("op %v pops wrong, us: %d, geth: %d", ourOp.String(), gotPops, geth_instr.MinStack)
				}
				havePush := len(ourOp.Pushes())
				wantPush := 1024 - geth_instr.MaxStack + geth_instr.MinStack
				if havePush != wantPush {
					t.Errorf("op %v push wrong, us: %d, geth: %d", ourOp.String(), havePush, wantPush)
				}

			*/
		}
	}
}
