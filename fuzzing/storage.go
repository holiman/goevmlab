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
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
	"math/rand"
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
			rand.Read(b)
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
			p.Push(0)
			p.Push(0)
			if r%2 == 0 {
				p.Op(ops.RETURN)
			} else {
				p.Op(ops.REVERT)
			}
			return p.Bytecode()
		}
		if len(p.Bytecode()) > 500 {
			return p.Bytecode()
		}
	}
}

func RandCallSubroutine(addresses []common.Address) []byte {

	addrGen := addressRandomizer(addresses)

	// We want to do a program which contains a mixture of
	// - BEGINSUB
	// - JUMPSUB
	// - JUMP / JUMPDEST
	// - RETURNSUB

	// Mixed with some calls to other programs

	// In order to not just run into illegal jumps all the time, we will place BEGINSUB
	// on every 10th pc, and JUMPEST on every 11:th

	p := program.NewProgram()
	legitSubs := make(map[int]bool)
	legitJumps := make(map[int]bool)

	randElem := func(m map[int]bool) int {
		for k := range m {
			return k
		}
		return 0
	}

	for {
		r := rand.Intn(16)
		switch r {
		case 0, 1, 2:
			// beginsub
			pc := int(p.Label())
			p.Op(ops.BEGINSUB)
			legitSubs[pc] = true
		case 3, 5, 6:
			// JUMPSUB
			pc := int(p.Label())
			p.Op(ops.JUMPDEST)
			legitJumps[pc] = true
		case 7, 8:
			// JUMP (legit)
			p.Jump(randElem(legitJumps))
		case 9:
			// random jump
			p.Jump(rand.Intn(1 + len(p.Bytecode())*3/2))
		case 10, 11:
			// JUMPSUB (legit)
			p.JumpSub(randElem(legitSubs))
		case 12:
			// JUMPSUB (random)
			p.JumpSub(rand.Intn(1 + len(p.Bytecode())*3/2))
		case 13, 14:
			// CALL
			// zero value call with no data
			p2 := RandCall(nil, addrGen, ValueRandomizer(), nil, nil)
			p.AddAll(p2)
			// pop the ret value
			p.Op(ops.POP)
		case 15:
			p.Push(0)
			p.Push(0)
			p.Op(ops.RETURN)
			return p.Bytecode()
		case 16:
			p.Push(0)
			p.Push(0)
			p.Op(ops.REVERT)
			return p.Bytecode()
		}
		if len(p.Bytecode()) > 1024 {
			return p.Bytecode()
		}
	}
}
