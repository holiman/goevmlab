package evms

import (
	"encoding/json"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
)

func JSONMarshal(log *logger.StructLog) []byte {
	data, _ := json.Marshal(log)
	return data
}

// CustomMarshal writes a logger.Structlog element into a concise json format.
// OBS! This output format will omit all stack element except the last 6 items.
func CustomMarshal(log *opLog) []byte {
	b := make([]byte, 0, 200)

	// Depth : PC
	b = append(b, `{"depth":`...)
	b = strconv.AppendUint(b, uint64(log.Depth), 10)

	b = append(b, []byte(`,"pc":`)...)
	b = strconv.AppendUint(b, uint64(log.Pc), 10)

	if !IgnoreEOF {
		// code section, if not zero
		if log.Section != 0 {
			b = append(b, []byte(`,"section":`)...)
			b = strconv.AppendUint(b, uint64(log.Section), 10)
		}

		// function call depth, if not zero
		if log.FunctionDepth != 0 {
			b = append(b, []byte(`,"functionDepth":`)...)
			b = strconv.AppendUint(b, uint64(log.FunctionDepth), 10)
		}
	}

	// Gas remaining
	b = append(b, []byte(`,"gas":`)...)
	b = strconv.AppendUint(b, uint64(log.Gas), 10)

	// Op
	b = append(b, []byte(`,"op":"0x`)...)
	opcode := uint64(log.Op)
	if opcode < 0x10 {
		b = append(b, []byte(`0`)...)
	}
	b = strconv.AppendUint(b, opcode, 16)
	b = append(b, []byte(`","opName":"`)...)
	b = append(b, []byte(log.Op.String())...)
	b = append(b, '"')

	// Gascost of operation
	if !ClearGascost {
		b = append(b, []byte(`,"gasCost":`)...)
		b = strconv.AppendUint(b, uint64(log.GasCost), 10)
	}
	// Memory size
	if !ClearMemSize {
		b = append(b, []byte(`,"memorySize":`)...)
		b = strconv.AppendUint(b, uint64(log.MemorySize), 10)
	}
	// Refunds
	if !ClearRefunds {
		panic("refunds parsing is not implemented now")
		//b = append(b, []byte(`,"refund":`)...)
		//b = strconv.AppendUint(b, uint64(log.RefundCounter), 10)
	}
	// Returndata
	if !ClearReturndata {
		b = append(b, []byte(`,"returnData":"0x`)...)
		b = append(b, hexutil.Encode(log.ReturnData)...)
		b = append(b, '"')
	}
	// Stack
	// At most 6 stack items, top item last
	b = append(b, []byte(`,"stack":[`)...)
	start := len(log.Stack) - 6
	if start < 0 {
		start = 0
	}
	for i := start; i < len(log.Stack); i++ {
		if i != start {
			b = append(b, ',')
		}
		b = append(b, '"')
		b = append(b, []byte(log.Stack[i].Hex())...)
		b = append(b, '"')
	}
	b = append(b, ']')
	// Error, if any
	if log.Err != nil {
		b = append(b, []byte(`,"error":"`)...)
		b = append(b, []byte(log.Err.Error())...)
		b = append(b, '"')
	}
	b = append(b, '}')
	return b
}
