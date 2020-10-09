package evms

import "github.com/ethereum/go-ethereum/core/vm"

// RemoveUnsupportedElems removes some elements that not all clients support.
// Once the relenvant json-fields have been added, we can remove things from this
// method
func RemoveUnsupportedElems(elem *vm.StructLog) {
	// Parity is missing gasCost, memSize and refund
	elem.GasCost = 0
	elem.MemorySize = 0
	elem.RefundCounter = 0
	// Nethermind is missing returnStack and returnData
	elem.ReturnStack = make([]uint32, 0)
	elem.ReturnData = nil

}
