package evms

import (
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/holiman/uint256"
)

// RemoveUnsupportedElems removes some elements that not all clients support.
// Once the relevant json-fields have been added, we can remove things from this
// method.
func RemoveUnsupportedElems(elem *logger.StructLog) {
	if elem.Stack == nil {
		elem.Stack = make([]uint256.Int, 0)
	}
	elem.Memory = make([]byte, 0)
	// Besu sometimes reports GasCost of 0x7fffffffffffffff, along with ,"error":"Out of gas"
	elem.GasCost = 0

	// Nethermind spits out zero memSize unless it's run with full memory output.
	// https://github.com/NethermindEth/nethermind/issues/4955
	elem.MemorySize = 0

	// Nethermind does not support refundcounter
	elem.RefundCounter = 0

	// Nethermind is missing returnData
	elem.ReturnData = nil
}
