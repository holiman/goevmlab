package evms

import (
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/uint256"
)

const (
	// Nethermind does not support refundcounter
	ClearRefunds = true

	// Nethermind reports the change in memory on step earlier than others. E.g.
	// MSTORE shows the _new_ memory, besu/geth shows the old memory until the next op.
	// This could be handled differently, e.g. clearing only on mem-expanding ops.
	ClearMemSize = true

	// Nethermind reports the change in memory on step earlier than others. E.g.
	// MSTORE shows the _new_ memory, besu/geth shows the old memory until the next op.
	// Unfortunately, nethermind also "forgets" the memsize when an error occurs, reporting
	// memsize zero (see testdata/traces/stackUnderflow_nonzeroMem.json). So we use
	// ClearMemSize instead
	ClearMemSizeOnExpand = true

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
	} else if ClearMemSizeOnExpand && ops.OpCode(elem.Op).ExpandsMem() {
		elem.MemorySize = 0
	}
	if ClearRefunds {
		elem.RefundCounter = 0
	}
	if ClearReturndata {
		elem.ReturnData = nil
	}
}
