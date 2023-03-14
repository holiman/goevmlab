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

// RandStorage sets some slots
func RandStorage(maxSlots, maxVal int) map[common.Hash]common.Hash {
	storage := make(map[common.Hash]common.Hash)
	numSlots := rand.Intn(maxSlots)
	for i := 0; i < numSlots; i++ {
		v, slot := byte(rand.Intn(maxVal)), byte(rand.Intn(numSlots))
		storage[common.BytesToHash([]byte{slot})] = common.BytesToHash([]byte{v})
	}
	return storage
}

func RandStorageOps() *program.Program {
	p := program.NewProgram()
	for {
		r := rand.Intn(100)
		switch {
		case r < 40:
			slot, val := rand.Intn(5), rand.Intn(3)
			p.Sstore(slot, val)
		case r < 80:
			slot := rand.Intn(10)
			p.Push(slot)
			p.Op(ops.SLOAD)
			p.Op(ops.POP)
		default:
			return p
		}
	}
}

func RandCall2200(addresses []common.Address) []byte {
	return randCall2200(addresses, 0)
}
func randCall2200(addresses []common.Address, depth int) []byte {
	if depth > 10 {
		return []byte{}
	}
	addrGen := addressRandomizer(addresses)

	// 10% sstore,
	// 10% sload,
	// 30% valid ops
	// 10% random op
	// 25% call of some type
	// 5% create, 5% create2,
	// 5% return, 5% revert
	p := program.NewProgram()
	for {
		r := rand.Intn(101)
		switch {
		case r < 10:
			p.Sstore(rand.Intn(5), rand.Intn(3))
		case r < 20:
			slot := rand.Intn(5)
			p.Push(slot)
			p.Op(ops.SLOAD)
			p.Op(ops.POP)
		case r < 50: // 30% chance of well-formed opcode
			b := make([]byte, 10)
			_, _ = crand.Read(b)
			for i := 0; i < len(b); i++ {
				if op := ops.OpCode(b[i]); ops.IsDefined(op) {
					p.Op(op)
				}
			}
		case r < 60: // 10% chance of some random opcode
			p.Op(ops.OpCode(rand.Uint32()))
		case r < 80:
			// zero value call with no data
			p2 := RandCall(nil, addrGen, nil, nil, nil)
			p.AddAll(p2)
			// pop the ret value
			p.Op(ops.POP)
		case r < 90:
			ctor := RandStorageOps()
			runtimeCode := randCall2200(addresses, depth+1)
			ctor.ReturnData(runtimeCode)
			p.CreateAndCall(ctor.Bytecode(), r%2 == 0, randCallType())
		case r < 95:
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
