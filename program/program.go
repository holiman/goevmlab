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

package program

import (
	"encoding/binary"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/uint256"
	"math/big"
)

type Program struct {
	code []byte
}

func NewProgram() *Program {
	p := &Program{
		code: make([]byte, 0),
	}
	return p
}

func (p *Program) add(op byte) {
	p.code = append(p.code, op)
}

func (p *Program) pushBig(val *big.Int) {
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

// AddAll adds the data to the Program
func (p *Program) AddAll(data []byte) {
	p.code = append(p.code, data...)
}

// Op appends the given opcode
func (p *Program) Op(op ops.OpCode) {
	p.add(byte(op))
}

// Push creates a PUSHX instruction with the data provided
func (p *Program) Push(val interface{}) *Program {
	switch v := val.(type) {
	case int:
		p.pushBig(new(big.Int).SetUint64(uint64(v)))
	case uint64:
		p.pushBig(new(big.Int).SetUint64(v))
	case uint32:
		p.pushBig(new(big.Int).SetUint64(uint64(v)))
	case *big.Int:
		p.pushBig(v)
	case *uint256.Int:
		p.pushBig(v.ToBig())
	case uint256.Int:
		p.pushBig(v.ToBig())
	case common.Address:
		p.pushBig(new(big.Int).SetBytes(v.Bytes()))
	case *common.Address:
		p.pushBig(new(big.Int).SetBytes(v.Bytes()))
	case []byte:
		p.pushBig(new(big.Int).SetBytes(v))
	case byte:
		p.pushBig(new(big.Int).SetUint64(uint64(v)))
	case nil:
		p.pushBig(nil)
	default:
		panic(fmt.Sprintf("unsupported type %v", v))
	}
	return p
}

// Bytecode returns the Program bytecode
func (p *Program) Bytecode() []byte {
	return p.code
}

// Bytecode returns the Program bytecode
func (p *Program) Hex() string {
	return fmt.Sprintf("%02x", p.Bytecode())
}

func (p *Program) ExtcodeCopy(address, memOffset, codeOffset, length interface{}) {
	p.Push(length)
	p.Push(codeOffset)
	p.Push(memOffset)
	p.Push(address)
	p.Op(ops.EXTCODECOPY)
}

// Call is a convenience function to make a call
func (p *Program) Call(gas *big.Int, address, value, inOffset, inSize, outOffset, outSize interface{}) {
	p.Push(outSize)
	p.Push(outOffset)
	p.Push(inSize)
	p.Push(inOffset)
	p.Push(value)
	p.Push(address)
	if gas == nil {
		p.Op(ops.GAS)
	} else {
		p.pushBig(gas)
	}
	p.Op(ops.CALL)
}

// DelegateCall is a convenience function to make a delegatecall
func (p *Program) DelegateCall(gas *big.Int, address, inOffset, inSize, outOffset, outSize interface{}) {
	p.Push(outSize)
	p.Push(outOffset)
	p.Push(inSize)
	p.Push(inOffset)
	p.Push(address)
	if gas == nil {
		p.Op(ops.GAS)
	} else {
		p.pushBig(gas)
	}
	p.Op(ops.DELEGATECALL)
}

// StaticCall is a convenience function to make a staticcall
func (p *Program) StaticCall(gas *big.Int, address, inOffset, inSize, outOffset, outSize interface{}) {
	p.Push(outSize)
	p.Push(outOffset)
	p.Push(inSize)
	p.Push(inOffset)
	p.Push(address)
	if gas == nil {
		p.Op(ops.GAS)
	} else {
		p.pushBig(gas)
	}
	p.Op(ops.STATICCALL)
}

func (p *Program) CallCode(gas *big.Int, address, value, inOffset, inSize, outOffset, outSize interface{}) {
	p.Push(outSize)
	p.Push(outOffset)
	p.Push(inSize)
	p.Push(inOffset)
	p.Push(value)
	p.Push(address)
	if gas == nil {
		p.Op(ops.GAS)
	} else {
		p.pushBig(gas)
	}
	p.Op(ops.CALLCODE)
}

// Label returns the PC (of the next instruction)
func (p *Program) Label() uint64 {
	return uint64(len(p.code))
}

// Jumpdest adds a JUMPDEST op, and returns the PC of that instruction
func (p *Program) Jumpdest() uint64 {
	here := p.Label()
	p.Op(ops.JUMPDEST)
	return here
}

// Jump pushes the destination and adds a JUMP
func (p *Program) Jump(loc interface{}) {
	p.Push(loc)
	p.Op(ops.JUMP)
}

// Jump pushes the destination and adds a JUMP
func (p *Program) JumpIf(loc interface{}, condition interface{}) {
	p.Push(condition)
	p.Push(loc)
	p.Op(ops.JUMPI)
}

func (p *Program) Size() int {
	return len(p.code)
}

// InputToMemory stores the input (calldata) to memory as address (20 bytes).
func (p *Program) InputAddressToStack(inputOffset uint32) {
	p.Push(inputOffset)
	p.Op(ops.CALLDATALOAD) // Loads [n -> n + 32] of input data to stack top
	mask, ok := big.NewInt(0).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", 16)
	if !ok {
		panic("whoa")
	}
	p.Push(mask) // turn into address
	p.Op(ops.AND)
}

// MStore stores the provided data (into the memory area starting at memStart)
func (p *Program) Mstore(data []byte, memStart uint32) {
	var idx = 0
	// We need to store it in chunks of 32 bytes
	for ; idx+32 <= len(data); idx += 32 {
		chunk := data[idx : idx+32]
		// push the value
		p.Push(chunk)
		// push the memory index
		p.Push(uint32(idx) + memStart)
		p.Op(ops.MSTORE)
	}
	// Remainders become stored using MSTORE8
	for ; idx < len(data); idx++ {
		b := data[idx]
		// push the byte
		p.Push(b)
		p.Push(uint32(idx) + memStart)
		p.Op(ops.MSTORE8)
	}
}

// MemToStorage copies the given memory area into SSTORE slots,
// It expects data to be aligned to 32 byte, and does not zero out
// remainders if some data is not
// I.e, if given a 1-byte area, it will still copy the full 32 bytes to storage
func (p *Program) MemToStorage(memStart, memSize, startSlot int) {
	// We need to store it in chunks of 32 bytes
	for idx := memStart; idx < (memStart + memSize); idx += 32 {
		dataStart := idx
		// Mload the chunk
		p.Push(dataStart)
		p.Op(ops.MLOAD)
		// Value is now on stack,
		p.Push(startSlot)
		p.Op(ops.SSTORE)
		startSlot++
	}
}

// Sstore stores the given byte array to the given slot.
// OBS! Does not verify that the value indeed fits into 32 bytes
// If it does not, it will panic later on via pushBig
func (p *Program) Sstore(slot interface{}, value interface{}) {
	p.Push(value)
	p.Push(slot)
	p.Op(ops.SSTORE)
}

func (p *Program) Return(offset, len uint32) {
	p.Push(len)
	p.Push(offset)
	p.Op(ops.RETURN)
}

// ReturnData loads the given data into memory, and does a return with it
func (p *Program) ReturnData(data []byte) {
	p.Mstore(data, 0)
	p.Return(0, uint32(len(data)))
}

// CreateAndCall calls create/create2 with the given bytecode
// and then checks if the returnvalue is non-zero. If so, it calls into the
// newly created contract with all available gas
func (p *Program) CreateAndCall(code []byte, isCreate2 bool, callOp ops.OpCode) {
	var (
		value    = 0
		offset   = 0
		size     = len(code)
		salt     = 0
		createOp = ops.CREATE
	)
	// Load the code into mem
	p.Mstore(code, 0)
	// Create it
	if isCreate2 {
		p.Push(salt)
		createOp = ops.CREATE2
	}
	p.Push(size).Push(offset).Push(value).Op(createOp)
	// If there happen to be a zero on the stack, it doesn't matter, we're
	// not sending any value anyway
	p.Push(0).Push(0) // mem out
	p.Push(0).Push(0) // mem in
	addrOffset := ops.DUP5
	if callOp == ops.CALL || callOp == ops.CALLCODE {
		p.Push(0) // value
		addrOffset = ops.DUP6
	}
	p.Op(addrOffset) // address (from create-op above)
	p.Op(ops.GAS)
	p.Op(callOp)
	p.Op(ops.POP) // pop the retval
	p.Op(ops.POP) // pop the address
}

// Push0 implements PUSH0 (0x5f)
func (p *Program) Push0() {
	p.Op(ops.PUSH0)
}

// RJump implements RJUMP (0x5c) - relative jump
func (p *Program) RJump(relOffset uint16) {
	p.Op(ops.RJUMP)
	p.code = binary.BigEndian.AppendUint16(p.code, relOffset)
}

// RJumpI implements RJUMPI (0x5d) - conditional relative jump
func (p *Program) RJumpI(relOffset uint16, condition interface{}) {
	p.Push(condition)
	p.Op(ops.RJUMPI)
	p.code = binary.BigEndian.AppendUint16(p.code, relOffset)
}

// RJumpV implements RJUMPV (0x5e) - relative jump via jump table
func (p *Program) RJumpV(relOffsets []uint16) {
	p.Op(ops.RJUMPV)
	// Immediate 1: the length
	p.add(byte(len(relOffsets)))
	// Immediates 2...N, the offsets
	for _, offset := range relOffsets {
		p.code = binary.BigEndian.AppendUint16(p.code, offset)
	}
}

// CallF implements CALLF (0xb0) - call a function
func (p *Program) CallF(i uint16) {
	p.Op(ops.CALLF)
	// Has one immediate argument,code_section_index,
	// encoded as a 16-bit unsigned big-endian value.
	p.code = binary.BigEndian.AppendUint16(p.code, i)
}

// RetF implements RETF (0xb1) - return from a function
func (p *Program) RetF() {
	p.Op(ops.RETF)
}
