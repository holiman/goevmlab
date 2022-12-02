package evms

import (
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/holiman/uint256"
)

const (
	// Nethermind does not support refundcounter
	ClearRefunds = true

	// Nethermind spits out zero memSize unless it's run with full memory output.
	// https://github.com/NethermindEth/nethermind/issues/4955
	ClearMemSize = true

	// Nethermind is missing returnData
	ClearReturndata = true

	// Besu sometimes reports GasCost of 0x7fffffffffffffff, along with ,"error":"Out of gas"
	ClearGascost = true
)

// RemoveUnsupportedElems removes some elements that not all clients support.
// Once the relevant json-fields have been added, we can remove things from this
// method.
func RemoveUnsupportedElems(elem *logger.StructLog) {
	if elem.Stack == nil {
		elem.Stack = make([]uint256.Int, 0)
	}
	elem.Memory = make([]byte, 0)

	if ClearGascost {
		elem.GasCost = 0
	}
	if ClearMemSize {
		elem.MemorySize = 0
	}
	if ClearRefunds {
		elem.RefundCounter = 0
	}
	if ClearReturndata {
		elem.ReturnData = nil
	}
}
