// Package program contains utilities to build small evm programs. It has since been
// upstreamed into go-ethereum, only bits and pieces remain here.
package program

import (
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
)

// CreateAndCall calls create/create2 with the given bytecode
// and then checks if the returnvalue is non-zero. If so, it calls into the
// newly created contract with all available gas
func CreateAndCall(p *program.Program, code []byte, isCreate2 bool, callOp vm.OpCode) {
	var (
		value    = 0
		offset   = 0
		size     = len(code)
		salt     = 0
		createOp = vm.CREATE
	)
	// Load the code into mem
	p.Mstore(code, 0)
	// Create it
	if isCreate2 {
		p.Push(salt)
		createOp = vm.CREATE2
	}
	p.Push(size).Push(offset).Push(value).Op(createOp)
	// If there happen to be a zero on the stack, it doesn't matter, we're
	// not sending any value anyway
	p.Push(0).Push(0) // mem out
	p.Push(0).Push(0) // mem in
	addrOffset := vm.OpCode(vm.DUP5)
	if callOp == vm.CALL || callOp == vm.CALLCODE {
		p.Push(0) // value
		addrOffset = vm.DUP6
	}
	p.Op(addrOffset) // address (from create-op above)
	p.Op(vm.GAS, callOp)
	p.Op(vm.POP, vm.POP) // pop  retval, pop address
}

// RJump implements RJUMP (0x5c) - relative jump
func RJump(p *program.Program, relOffset uint16) {
	panic("Need RJUMP defined")
	//p.Op(ops.RJUMP)
	//p.code = binary.BigEndian.AppendUint16(p.code, relOffset)
}

// RJumpI implements RJUMPI (0x5d) - conditional relative jump
func RJumpI(p *program.Program, relOffset uint16, condition interface{}) {
	panic("Need RJUMPI defined") // unclear what op it is
	//p.Push(condition)
	//p.Op(ops.RJUMPI)
	//p.code = binary.BigEndian.AppendUint16(p.code, relOffset)
}

// RJumpV implements RJUMPV (0x5e) - relative jump via jump table
func RJumpV(p *program.Program, relOffsets []uint16) {
	panic("Need RJUMPV defined") // unclear what op it is
	//p.Op(ops.RJUMPV)
	// Immediate 1: the length
	//p.add(byte(len(relOffsets)))
	// Immediates 2...N, the offsets
	//for _, offset := range relOffsets {
	//	p.code = binary.BigEndian.AppendUint16(p.code, offset)
	//}
}

// CallF implements CALLF (0xb0) - call a function
func CallF(p *program.Program, i uint16) {
	panic("Need CALLF defined") // unclear what op it is
	//p.Op(ops.CALLF)
	// Has one immediate argument,code_section_index,
	// encoded as a 16-bit unsigned big-endian value.
	//p.code = binary.BigEndian.AppendUint16(p.code, i)
}

// RetF implements RETF (0xb1) - return from a function
func RetF(p *program.Program) {
	panic("Need RETF defined") // unclear what op it is
	//p.Op(ops.RETF)
}
