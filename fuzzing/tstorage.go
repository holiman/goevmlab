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

package fuzzing

import (
	crand "crypto/rand"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

func randTStoreOps() *program.Program {
	p := program.NewProgram()
	for {
		r := chance(rand.Intn(101))
		switch {
		case r.between(0, 20):
			p.Tstore(rand.Intn(5), rand.Intn(3))
		case r.between(20, 40):
			p.Sstore(rand.Intn(5), rand.Intn(3))
		case r.between(40, 60):
			p.Push(rand.Intn(10))
			p.Op(ops.TLOAD)
			p.Op(ops.POP)
		case r.between(60, 80):
			p.Push(rand.Intn(10))
			p.Op(ops.SLOAD)
			p.Op(ops.POP)
		default:
			return p
		}
	}
}

func RandCallTStore(addresses []common.Address) []byte {
	return randCallTStore(addresses, 0)
}

type chance int

func (c chance) between(a, b int) bool {
	return int(c) >= a && int(c) < b
}

// randCallTStore creates code which does a mix of TSTORE, TLOAD, SSTORE, SLOAD
// and other (mostly well-formed) opcodes, plys a fair bit of calls to other contracts.
func randCallTStore(addresses []common.Address, depth int) []byte {
	if depth > 10 {
		return []byte{}
	}
	addrGen := addressRandomizer(addresses)

	p := program.NewProgram()
	for {
		r := chance(rand.Intn(101))
		switch {
		case r.between(0, 10): // TSTORE 10%
			p.Tstore(rand.Intn(5), rand.Intn(10))
		case r.between(10, 20): // SSTORE 10%
			p.Sstore(rand.Intn(5), rand.Intn(10))
		case r.between(20, 35): // TLOAD 15%
			p.Push(rand.Intn(5))
			p.Op(ops.TLOAD)
			p.Op(ops.POP)
		case r.between(35, 50): // SLOAD 15%
			p.Push(rand.Intn(5))
			p.Op(ops.SLOAD)
			p.Op(ops.POP)
		case r.between(50, 60): // 10% chance of some well-formed opcodes
			b := make([]byte, 10)
			_, _ = crand.Read(b)
			for i := 0; i < len(b); i++ {
				if op := ops.OpCode(b[i]); ops.IsDefined(op) {
					p.Op(op)
				}
			}
		case r.between(60, 70): // 10% chance of some random opcode
			p.Op(ops.OpCode(rand.Uint32()))
		case r.between(70, 80): // 10% zero value call with no data
			p.AddAll(RandCall(nil, addrGen, nil, nil, nil))
			p.Op(ops.POP) // pop returnvalue
		case r.between(80, 90): // 10% create and call
			ctor := randTStoreOps()
			runtimeCode := randCallTStore(addresses, depth+1)
			ctor.ReturnData(runtimeCode)
			p.CreateAndCall(ctor.Bytecode(), r%2 == 0, randCallType())
		case r.between(90, 95):
			p.Push(addrGen())
			p.Op(ops.SELFDESTRUCT)
		default:
			p.Push(32) //len
			p.Push(0)  //offset
			if r%2 == 0 {
				p.Op(ops.RETURN)
			} else {
				p.Op(ops.REVERT)
			}
			return p.Bytecode()
		}
		if p.Size() > 500 {
			return p.Bytecode()
		}
	}
}
