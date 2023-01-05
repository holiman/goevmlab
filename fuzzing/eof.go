package fuzzing

import (
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
	"math/rand"
)

func OneOf(cases ...int) int {
	a := rand.Intn(len(cases))
	return cases[a]
}

//
//func generateEofContainer(rnd RandSource) {
//	var c vm.Container
//	numCodes := 1024
//
//	for i := 0; i < numCodes; i++ {
//		code, maxStack := genCallFProgram()
//		c.Code = append(c.Code, code)
//		var metadata = &vm.FunctionMetadata{
//			Input:          uint8(0),
//			Output:         uint8(0),
//			MaxStackHeight: uint16(maxStack),
//		}
//		if i == 0 {
//			metadata.Input = 0
//			metadata.Output = 0
//		}
//		c.Types = append(c.Types, metadata)
//	}
//	fmt.Printf("%x\n", c.MarshalBinary())
//}

func GenerateCallFProgram() ([]byte, int) {

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
	for {
		switch OneOf(1, 2, 3, 4, 5) {
		case 1:
			p.CallF(uint16(rand.Intn(1024)))
		case 2:
			p.RetF()
		case 3:
			p.RJump(0)
		case 4:
			// we push one and pop one
			p.RJumpI(0, 0)
			if maxStack < curStack+1 {
				maxStack = curStack + 1
			}
			p.Op(ops.STOP)
		default:
			p.Push0()
			p.RJumpV(0, []uint16{0})
			// we push one and pop one
			if maxStack < curStack+1 {
				maxStack = curStack + 1
			}
			p.Op(ops.STOP)
		}
		break
	}
	return p.Bytecode(), maxStack
}
