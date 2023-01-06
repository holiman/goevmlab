// Copyright 2023 Martin Holst Swende
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

package fuzzing

import (
	"fmt"
	"math/big"
	"math/rand"

	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func oneOf(cases ...any) any {
	return cases[rand.Intn(len(cases))]
}

func asBig(in string) *big.Int {
	a := new(big.Int)
	a, ok := a.SetString(in, 0)
	if !ok {
		panic(fmt.Sprintf("bad input: %v", in))
	}
	return a
}

func GenerateCallFProgram(maxSections int) ([]byte, int) {

	// The section is comprised of a list of metadata where the metadata index in
	// the type section corresponds to a code section index.
	// Therefore, the type section size MUST be n * 4 bytes, where n is the
	// number of code sections.
	//	Each metadata item has 3 attributes:
	//  	a uint8 inputs, a uint8 outputs,
	//  	and a uint16 max_stack_height.
	//  	Note: This implies that there is a limit of 255 stack for the input and in the output.
	//  	This is further restricted to 127 stack items, because the upper bit of both the input
	//  	and output bytes are reserved for future use. max_stack_height is further defined in EIP-5450.
	// The first code section MUST have 0 inputs and 0 outputs.

	var p = program.NewProgram()
	maxStack := 0
	curStack := 0
	//for {
	switch oneOf(1, 2, 3, 4, 5, 6, 7) {
	case 1:
		p.CallF(uint16(rand.Intn(maxSections)))
		p.Op(ops.STOP)
	case 2:
		p.RetF()
	case 3:
		// jump to minus three
		p.RJump(uint16(0xffff - 2))
	case 4:
		// jump to plus one
		p.RJump(uint16(0))
		p.Op(ops.STOP) // jump location
	case 5:
		// jump to plus one
		p.RJump(uint16(0))
		p.Op(ops.STOP) // jump location
	case 6:
		// we push one and pop one
		p.RJumpI(0, 0)
		if maxStack < curStack+1 {
			maxStack = curStack + 1
		}
		p.Op(ops.STOP)
	default:
		//p.Push0()

		len := rand.Intn(255)
		p.Push(oneOf(
			asBig("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"),
			asBig("0x1000000000000000000000000000000000000000000000000000000000000000"),
			big.NewInt(int64(len)),
			big.NewInt(int64(len-1)),
			big.NewInt(int64(len+1)),
		))
		dests := make([]uint16, len)
		if len > 0 && rand.Intn(4) != 0 {
			dests[len-1] = uint16(0x10000 - 2*len - 2)
		}
		p.RJumpV(dests)
		// we push one and pop one
		if maxStack < curStack+1 {
			maxStack = curStack + 1
		}
		p.Op(ops.STOP)
	}
	//break
	//}
	return p.Bytecode(), maxStack
}
