package ops

import (
	"github.com/ethereum/go-ethereum/core/vm"
	"strings"
	"testing"
)

// TestSanity checks the npops and npushes against the
// go-ethereum codebase
func TestSanity(t *testing.T) {

	for i := 0; i < 256; i++ {

		// Geth has some 'virtual' opcodes: POP, PUSH, DUP and SWAP
		switch i {
		case 0x50, 0xb0, 0xb1, 0xb2:
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
		if name := ourOp.String(); !strings.HasPrefix(name, "Missing"){
			our := byte(StringToOp(name))
			if byte(our) != byte(i){
				t.Errorf("name %v, got 0x%x expected 0x%x", name, our, byte(i))
			}
			geth := byte(vm.StringToOp(name))
			if byte(geth) != byte(i){
				t.Errorf("name %v, got 0x%x expected 0x%x", name, geth, byte(i))
			}
		}
		// Check stack pushes and pops
		{
			// This check can only be executed if the go-ethereum codebase
			// is refactored a bit, to make the following public:
			// - vm.IstanbulInstructionSet
			// - vm.operation.MinStack
			// - vm.operation.MaxStack
			// Was tested on 2019-08-29, oll korrekt
			/*
			gotPops := len(ourOp.Pops())
			geth_instr := vm.IstanbulInstructionSet[gethOp]
			if gotPops != geth_instr.MinStack{
				t.Errorf("op %v pops wrong, us: %d, geth: %d", ourOp.String(), gotPops, geth_instr.MinStack)
			}
			*/
		}
	}
}
