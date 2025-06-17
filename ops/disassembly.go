package ops

import "fmt"

// instructionIterator is an iterator for disassembled EVM instructions.
type instructionIterator struct {
	code         []byte
	pc           uint64
	arg          []byte
	op           OpCode
	error        error
	started      bool
	stackBalance int
}

// NewInstructionIterator creates a new instruction iterator.
func NewInstructionIterator(code []byte) *instructionIterator {
	it := new(instructionIterator)
	it.code = code
	return it
}

// InstructionCount counts the number of instructions
func InstructionCount(code []byte) int {
	it := NewInstructionIterator(code)
	c := 0
	for it.Next() {
		c++
	}
	return c
}

// Skips num instructions.
func (it *instructionIterator) Skip(num int) {
	c := 0
	for it.Next() && c < num {
		c++
	}
}

// Next returns true if there is a next instruction and moves on.
func (it *instructionIterator) Next() bool {
	if it.error != nil || uint64(len(it.code)) <= it.pc {
		// We previously reached an error or the end.
		return false
	}

	if it.started {
		// Since the iteration has been already started we move to the next instruction.
		if it.arg != nil {
			it.pc += uint64(len(it.arg))
		}
		it.pc++
	} else {
		// We start the iteration from the first instruction.
		it.started = true
	}

	if uint64(len(it.code)) <= it.pc {
		// We reached the end.
		return false
	}
	it.op = OpCode(it.code[it.pc])
	if it.op.HasImmediate() {
		switch {
		case it.op >= PUSH1 && it.op <= PUSH32:
			a := uint64(it.op) - uint64(PUSH1) + 1
			u := it.pc + 1 + a
			if uint64(len(it.code)) < u {
				it.error = fmt.Errorf("incomplete push instruction at %v", it.pc)
				return false
			}
			it.arg = it.code[it.pc+1 : u]
		//case it.op == RJUMP || it.op == RJUMPI:
		//	u := it.pc + 1 + 2
		//	if uint64(len(it.code)) < u {
		//		it.error = fmt.Errorf("incomplete RJUMP/RJUMPI instruction at %v", it.pc)
		//		return false
		//	}
		//	it.arg = it.code[it.pc+1 : u]
		//case it.op == RJUMPV:
		//	// First we need to peek at the next byte, to see the length
		//	if uint64(len(it.code)) <= it.pc+1 {
		//		it.error = fmt.Errorf("incomplete RJUMPV instruction at %v", it.pc)
		//		return false
		//	}
		//	count := uint64(it.code[it.pc+1])
		//	// The rumpv table is count x uint16 bytes large
		//	a := 1 + 2*count
		//	u := it.pc + 1 + a
		//	if uint64(len(it.code)) < u {
		//		it.error = fmt.Errorf("incomplete RJUMPV instruction at %v", it.pc)
		//		return false
		//	}
		//	it.arg = it.code[it.pc+1 : u]
		default:
			panic("Unkown op")
		}
	} else {
		it.arg = nil
	}
	it.stackBalance += it.op.Stackdelta()
	return true
}

// Error returns any error that may have been encountered.
func (it *instructionIterator) Error() error {
	return it.error
}

// PC returns the PC of the current instruction.
func (it *instructionIterator) PC() uint64 {
	return it.pc
}

// Op returns the opcode of the current instruction.
func (it *instructionIterator) Op() OpCode {
	return it.op
}

// Arg returns the argument of the current instruction.
func (it *instructionIterator) Arg() []byte {
	return it.arg
}
