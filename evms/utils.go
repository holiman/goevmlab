package evms

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/vm"
)

// RemoveUnsupportedElems removes some elements that not all clients support.
// Once the relenvant json-fields have been added, we can remove things from this
// method
func RemoveUnsupportedElems(elem *vm.StructLog) {
	if elem.Stack == nil {
		elem.Stack = make([]*big.Int, 0)
	}
	elem.Memory = make([]byte, 0)
	// Parity is missing gasCost, memSize and refund
	elem.GasCost = 0
	elem.MemorySize = 0
	elem.RefundCounter = 0
	// Nethermind is missing returnData
	elem.ReturnData = nil

}
