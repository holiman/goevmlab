package evms

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
)

//go:generate go run github.com/fjl/gencodec -type opLog -field-override opLogMarshaling -out gen_oplog.go

// opLog represent a line of output from an EVM. It is very similar to
// go-ethereum StructLog, but
//   - also has the ability to soup up a stateroot,
//   - and is a bit more lax in parsing (e.g allows negative refund)
type opLog struct {
	Pc            uint64        `json:"pc"`
	Section       uint64        `json:"section,omitempty"`
	Op            vm.OpCode     `json:"op"`
	Gas           uint64        `json:"gas"`
	GasCost       uint64        `json:"gasCost"`
	Memory        []byte        `json:"memory,omitempty"`
	MemorySize    int           `json:"memSize"`
	Stack         []uint256.Int `json:"stack"`
	ReturnData    []byte        `json:"returnData,omitempty"`
	Depth         int           `json:"depth"`
	FunctionDepth int           `json:"functionDepth,omitempty"`
	Err           error         `json:"-"`

	// stateroot as output by geth, reth, eels, nethermind
	StateRoot1 string `json:"stateRoot"`
	// stateroot as output by besu
	StateRoot2 string `json:"postHash"`
}

// overrides for gencodec
type opLogMarshaling struct {
	Gas        math.HexOrDecimal64
	GasCost    math.HexOrDecimal64
	Memory     hexutil.Bytes
	ReturnData hexutil.Bytes
	//RefundCounter HexOrDecimalSigned64
	MemorySize math.HexOrDecimal64
	Stack      []hexutil.U256
	OpName     string `json:"opName"` // adds call to OpName() in MarshalJSON
}

// OpName formats the operand name in a human-readable format.
func (l *opLog) OpName() string {
	return l.Op.String()
}
