// Copyright 2019 Martin Holst Swende
// This file is part of the go-evmlab library.
//
// The go-evmlab library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package program

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"
)

type program struct {
	code []byte
}

func NewProgram() *program {
	p := &program{
		code: make([]byte, 0),
	}
	return p
}

func (p *program) add(op byte) {
	p.code = append(p.code, op)
}

func (p *program) pushBig(val *big.Int) {
	if val == nil {
		val = big.NewInt(0)
	}
	valBytes := val.Bytes()
	if len(valBytes) == 0 {
		valBytes = append(valBytes, 0)
	}
	bLen := len(valBytes)
	if bLen > 32 {
		panic(fmt.Sprintf("Push value too large, %d bytes", bLen))
	}
	p.add(byte(vm.PUSH1) - 1 + byte(bLen))
	p.AddAll(valBytes)

}

// AddAll adds the data to the program
func (p *program) AddAll(data []byte) {
	p.code = append(p.code, data...)
}

// Op appends the given opcode
func (p *program) Op(opCode vm.OpCode) {
	p.add(byte(opCode))
}

// Push creates a PUSHX instruction with the data provided
func (p *program) Push(val interface{}) {
	switch v := val.(type) {
	case int:
		p.pushBig(new(big.Int).SetUint64(uint64(v)))
	case uint64:
		p.pushBig(new(big.Int).SetUint64(v))
	case *big.Int:
		p.pushBig(v)
	case common.Address:
		p.pushBig(new(big.Int).SetBytes(v.Bytes()))
	case *common.Address:
		p.pushBig(new(big.Int).SetBytes(v.Bytes()))
	case nil:
		p.pushBig(nil)
	default:
		panic(fmt.Sprintf("unsupported type %v", v))
	}
}

// Bytecode returns the program bytecode
func (p *program) Bytecode() []byte {
	return p.code
}

// Call is a convenience function to make a call
func (p *program) Call(gas *big.Int, address, value, inOffset, inSize, outOffset, outSize interface{}) {
	p.Push(outSize)
	p.Push(outOffset)
	p.Push(inSize)
	p.Push(inOffset)
	p.Push(value)
	p.Push(address)
	if gas == nil {
		p.Op(vm.GAS)
	} else {
		p.pushBig(gas)
	}
	p.Op(vm.CALL)
}

// Label returns the PC (of the next instruction)
func (p *program) Label() uint64 {
	return uint64(len(p.code))
}

// Jumpdest adds a JUMPDEST op, and returns the PC of that instruction
func (p *program) Jumpdest() uint64 {
	here := p.Label()
	p.Op(vm.JUMPDEST)
	return here
}

// Jump pushes the destination and adds a JUMP
func (p *program) Jump(loc interface{}) {
	p.Push(loc)
	p.Op(vm.JUMP)
}
