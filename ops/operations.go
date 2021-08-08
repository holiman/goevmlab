// Copyright 2019 Martin Holst Swende
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the goevmlab library. If not, see <http://www.gnu.org/licenses/>.

package ops

import (
	"fmt"
)

// OpCode is an EVM opcode
type OpCode byte

// IsPush specifies if an opcode is a PUSH opcode.
func (op OpCode) IsPush() bool {
	return op >= PUSH1 && op <= PUSH32
}
func (op OpCode) IsCall() bool {
	return op == CALL ||
		op == DELEGATECALL ||
		op == CALLCODE ||
		op == STATICCALL

}

func (op OpCode) PushSize() int {
	if op.IsPush() {
		return (int(op) - int(PUSH1) + 1)
	}
	return 0
}

// 0x0 range - arithmetic ops.
const (
	STOP       = OpCode(0x00)
	ADD        = OpCode(0x01)
	MUL        = OpCode(0x02)
	SUB        = OpCode(0x03)
	DIV        = OpCode(0x04)
	SDIV       = OpCode(0x05)
	MOD        = OpCode(0x06)
	SMOD       = OpCode(0x07)
	ADDMOD     = OpCode(0x08)
	MULMOD     = OpCode(0x09)
	EXP        = OpCode(0x0A)
	SIGNEXTEND = OpCode(0x0B)
)

// 0x10 range - comparison ops.
const (
	LT     = OpCode(0x10)
	GT     = OpCode(0x11)
	SLT    = OpCode(0x12)
	SGT    = OpCode(0x13)
	EQ     = OpCode(0x14)
	ISZERO = OpCode(0x15)
	AND    = OpCode(0x16)
	OR     = OpCode(0x17)
	XOR    = OpCode(0x18)
	NOT    = OpCode(0x19)
	BYTE   = OpCode(0x1A)
	SHL    = OpCode(0x1B)
	SHR    = OpCode(0x1C)
	SAR    = OpCode(0x1D)

	SHA3 = OpCode(0x20)
)

// 0x30 range - closure state.
const (
	ADDRESS        = OpCode(0x30)
	BALANCE        = OpCode(0x31)
	ORIGIN         = OpCode(0x32)
	CALLER         = OpCode(0x33)
	CALLVALUE      = OpCode(0x34)
	CALLDATALOAD   = OpCode(0x35)
	CALLDATASIZE   = OpCode(0x36)
	CALLDATACOPY   = OpCode(0x37)
	CODESIZE       = OpCode(0x38)
	CODECOPY       = OpCode(0x39)
	GASPRICE       = OpCode(0x3A)
	EXTCODESIZE    = OpCode(0x3B)
	EXTCODECOPY    = OpCode(0x3C)
	RETURNDATASIZE = OpCode(0x3D)
	RETURNDATACOPY = OpCode(0x3E)
	EXTCODEHASH    = OpCode(0x3F)
)

// 0x40 range - block operations.
const (
	BLOCKHASH   = OpCode(0x40)
	COINBASE    = OpCode(0x41)
	TIMESTAMP   = OpCode(0x42)
	NUMBER      = OpCode(0x43)
	DIFFICULTY  = OpCode(0x44)
	GASLIMIT    = OpCode(0x45)
	CHAINID     = OpCode(0x46)
	SELFBALANCE = OpCode(0x47)
	BASEFEE     = OpCode(0x48)
)

// 0x50 range - 'storage' and execution.
const (
	POP       = OpCode(0x50)
	MLOAD     = OpCode(0x51)
	MSTORE    = OpCode(0x52)
	MSTORE8   = OpCode(0x53)
	SLOAD     = OpCode(0x54)
	SSTORE    = OpCode(0x55)
	JUMP      = OpCode(0x56)
	JUMPI     = OpCode(0x57)
	PC        = OpCode(0x58)
	MSIZE     = OpCode(0x59)
	GAS       = OpCode(0x5A)
	JUMPDEST  = OpCode(0x5B)
	BEGINSUB  = OpCode(0x5c) // Not live, used in a test
	RETURNSUB = OpCode(0x5d) // Not live, used in a test
	JUMPSUB   = OpCode(0x5e) // Not live, used in a test
)

// 0x60 through 0x7F range.
const (
	PUSH1 OpCode = 0x60 + iota
	PUSH2
	PUSH3
	PUSH4
	PUSH5
	PUSH6
	PUSH7
	PUSH8
	PUSH9
	PUSH10
	PUSH11
	PUSH12
	PUSH13
	PUSH14
	PUSH15
	PUSH16
	PUSH17
	PUSH18
	PUSH19
	PUSH20
	PUSH21
	PUSH22
	PUSH23
	PUSH24
	PUSH25
	PUSH26
	PUSH27
	PUSH28
	PUSH29
	PUSH30
	PUSH31
	PUSH32
)

// 0x80 range
const (
	DUP1 OpCode = 0x80 + iota
	DUP2
	DUP3
	DUP4
	DUP5
	DUP6
	DUP7
	DUP8
	DUP9
	DUP10
	DUP11
	DUP12
	DUP13
	DUP14
	DUP15
	DUP16
)

// 0x90 range
const (
	SWAP1 OpCode = 0x90 + iota
	SWAP2
	SWAP3
	SWAP4
	SWAP5
	SWAP6
	SWAP7
	SWAP8
	SWAP9
	SWAP10
	SWAP11
	SWAP12
	SWAP13
	SWAP14
	SWAP15
	SWAP16
)

// 0xa0 range - logging ops.
const (
	LOG0 OpCode = 0xa0 + iota
	LOG1
	LOG2
	LOG3
	LOG4
)

// unofficial opcodes used for parsing.
const (
	PUSH OpCode = 0xb0 + iota
	DUP
	SWAP
)

// 0xf0 range - closures.
const (
	CREATE       = OpCode(0xf0)
	CALL         = OpCode(0xf1)
	CALLCODE     = OpCode(0xf2)
	RETURN       = OpCode(0xf3)
	DELEGATECALL = OpCode(0xf4)
	CREATE2      = OpCode(0xf5)

	STATICCALL = OpCode(0xfa)

	REVERT       = OpCode(0xfd)
	SELFDESTRUCT = OpCode(0xff)
)

func (op OpCode) String() string {
	if info, ok := opCodeInfo[op]; ok {
		return info.name
	}
	return fmt.Sprintf("opcode 0x%x not defined", int(op))
}

// stringToOp is a mapping from strings to OpCode
var stringToOp map[string]OpCode

// ValidOpcodes is the set of valid opcodes
var ValidOpcodes []OpCode

func init() {
	stringToOp = make(map[string]OpCode)

	for k, elem := range opCodeInfo {
		stringToOp[elem.name] = k
		ValidOpcodes = append(ValidOpcodes, k)
	}
}

// StringToOp finds the opcode whose name is stored in `str`.
func StringToOp(str string) OpCode {
	return stringToOp[str]
}

type opInfo struct {
	name   string
	pops   []string
	pushes []string
}

var opCodeInfo = map[OpCode]opInfo{

	STOP:       {"STOP", nil, nil},
	ADD:        {"ADD", []string{"a", "b"}, []string{"a + b"}},
	MUL:        {"MUL", []string{"a", "b"}, []string{"a * b"}},
	SUB:        {"SUB", []string{"a", "b"}, []string{"a - b"}},
	DIV:        {"DIV", []string{"a", "b"}, []string{"a / b"}},
	SDIV:       {"SDIV", []string{"a", "b"}, []string{"a / b (signed)"}},
	MOD:        {"MOD", []string{"a", "b"}, []string{"a % b"}},
	SMOD:       {"SMOD", []string{"a", "b"}, []string{"a mod b (signed)"}},
	EXP:        {"EXP", []string{"base", "exp"}, []string{"base^exp"}},
	NOT:        {"NOT", []string{"a"}, []string{"not(a)"}},
	LT:         {"LT", []string{"a", "b"}, []string{"a < b"}},
	GT:         {"GT", []string{"a", "b"}, []string{"a > b"}},
	SLT:        {"SLT", []string{"a", "b"}, []string{"a < b (signed)"}},
	SGT:        {"SGT", []string{"a", "b"}, []string{"a > b (signed)"}},
	EQ:         {"EQ", []string{"a", "b"}, []string{"a == b"}},
	ISZERO:     {"ISZERO", []string{"a"}, []string{"a == 0"}},
	SIGNEXTEND: {"SIGNEXTEND", []string{"bitlen", "a"}, []string{"signextend(a, bitlen)"}},

	AND:    {"AND", []string{"a", "b"}, []string{"a && b"}},
	OR:     {"OR", []string{"a", "b"}, []string{"a || b"}},
	XOR:    {"XOR", []string{"a", "b"}, []string{"a xor b"}},
	BYTE:   {"BYTE", []string{"index", "val"}, []string{"byte at val[index]"}},
	SHL:    {"SHL", []string{"shift", "x"}, []string{"x << shift"}},
	SHR:    {"SHR", []string{"shift", "x"}, []string{"x >> shift"}},
	SAR:    {"SAR", []string{"shift", "x"}, []string{"x >>> shift"}},
	ADDMOD: {"ADDMOD", []string{"a", "b", "x"}, []string{"(a + b) mod x"}},
	MULMOD: {"MULMOD", []string{"a", "b", "x"}, []string{"(a * b) mod x"}},

	// 0x20 range - crypto.
	SHA3: {"SHA3", []string{"offset", "size"}, []string{"keccak256(mem[offset:offset+size])"}},
	// 0x30 range - closure state.
	ADDRESS:      {"ADDRESS", nil, []string{"address of current context"}},
	BALANCE:      {"BALANCE", []string{"address"}, []string{"balance of address"}},
	ORIGIN:       {"ORIGIN", nil, []string{"transaction origin"}},
	CALLER:       {"CALLER", nil, []string{"sender"}},
	CALLVALUE:    {"CALLVALUE", nil, []string{"call value"}},
	CALLDATALOAD: {"CALLDATALOAD", []string{"offset"}, []string{"calldata[offset:offset+32]"}},
	CALLDATASIZE: {"CALLDATASIZE", nil, []string{"size of calldata"}},
	CALLDATACOPY: {"CALLDATACOPY", []string{"memOffset", "dataOffset", "length"}, nil},
	CODESIZE:     {"CODESIZE", nil, []string{"size of code in this context"}},
	CODECOPY:     {"CODECOPY", []string{"memOffset", "codeOffset", "length"}, nil},
	GASPRICE:     {"GASPRICE", nil, []string{"transaction gasprice"}},

	EXTCODESIZE: {"EXTCODESIZE", []string{"address"}, []string{"code size at 'address'"}},
	EXTCODECOPY: {"EXTCODECOPY", []string{"address", "memOffset", "codeOffset", "length"}, nil},

	RETURNDATASIZE: {"RETURNDATASIZE", nil, []string{"size of returndata"}},
	RETURNDATACOPY: {"RETURNDATACOPY", []string{"memOffset", "dataOffset", "length"}, nil},
	EXTCODEHASH:    {"EXTCODEHASH", []string{"address"}, []string{"codehash at 'address'"}},

	// 0x40 range - block operations.
	BLOCKHASH:   {"BLOCKHASH", []string{"blocknum"}, []string{"hash of block at blocknum"}},
	COINBASE:    {"COINBASE", nil, []string{"block miner address"}},
	TIMESTAMP:   {"TIMESTAMP", nil, []string{"unix time of current block"}},
	NUMBER:      {"NUMBER", nil, []string{"current block number"}},
	DIFFICULTY:  {"DIFFICULTY", nil, []string{"current block difficulty"}},
	GASLIMIT:    {"GASLIMIT", nil, []string{"block gas limit"}},
	CHAINID:     {"CHAINID", nil, []string{"chain id"}},
	SELFBALANCE: {"SELFBALANCE", nil, []string{"balance at current context"}},
	BASEFEE:     {"BASEFEE", nil, []string{"basefee in current block"}},

	POP:      {"POP", nil, []string{"value to pop"}},
	MLOAD:    {"MLOAD", []string{"offset"}, nil},
	MSTORE:   {"MSTORE", []string{"offset", "value"}, nil},
	MSTORE8:  {"MSTORE8", []string{"offset", "value"}, nil},
	SLOAD:    {"SLOAD", []string{"slot"}, nil},
	SSTORE:   {"SSTORE", []string{"slot", "value"}, nil},
	JUMP:     {"JUMP", []string{"loc"}, nil},
	JUMPI:    {"JUMPI", []string{"loc", "cond"}, nil},
	PC:       {"PC", nil, []string{"current PC"}},
	MSIZE:    {"MSIZE", nil, []string{"size of memory"}},
	GAS:      {"GAS", nil, []string{"current gas remaining"}},
	JUMPDEST: {"JUMPDEST", nil, nil},
	//BEGINSUB:  {"BEGINSUB", nil, nil},
	//RETURNSUB: {"RETURNSUB", nil, nil},
	//JUMPSUB:   {"JUMPSUB", []string{"subroutine destination"}, nil},
	// 0x60 through 0x7F range - push.
	PUSH1:  {"PUSH1", nil, []string{"1 byte pushed value"}},
	PUSH2:  {"PUSH2", nil, []string{"2 bytes pushed value"}},
	PUSH3:  {"PUSH3", nil, []string{"3 bytes pushed value"}},
	PUSH4:  {"PUSH4", nil, []string{"4 bytes pushed value"}},
	PUSH5:  {"PUSH5", nil, []string{"5 bytes pushed value"}},
	PUSH6:  {"PUSH6", nil, []string{"6 bytes pushed value"}},
	PUSH7:  {"PUSH7", nil, []string{"7 bytes pushed value"}},
	PUSH8:  {"PUSH8", nil, []string{"8 bytes pushed value"}},
	PUSH9:  {"PUSH9", nil, []string{"9 bytes pushed value"}},
	PUSH10: {"PUSH10", nil, []string{"10 bytes pushed value"}},
	PUSH11: {"PUSH11", nil, []string{"11 bytes pushed value"}},
	PUSH12: {"PUSH12", nil, []string{"12 bytes pushed value"}},
	PUSH13: {"PUSH13", nil, []string{"13 bytes pushed value"}},
	PUSH14: {"PUSH14", nil, []string{"14 bytes pushed value"}},
	PUSH15: {"PUSH15", nil, []string{"15 bytes pushed value"}},
	PUSH16: {"PUSH16", nil, []string{"16 bytes pushed value"}},
	PUSH17: {"PUSH17", nil, []string{"17 bytes pushed value"}},
	PUSH18: {"PUSH18", nil, []string{"18 bytes pushed value"}},
	PUSH19: {"PUSH19", nil, []string{"19 bytes pushed value"}},
	PUSH20: {"PUSH20", nil, []string{"19 bytes pushed value"}},
	PUSH21: {"PUSH21", nil, []string{"21 bytes pushed value"}},
	PUSH22: {"PUSH22", nil, []string{"22 bytes pushed value"}},
	PUSH23: {"PUSH23", nil, []string{"23 bytes pushed value"}},
	PUSH24: {"PUSH24", nil, []string{"24 bytes pushed value"}},
	PUSH25: {"PUSH25", nil, []string{"25 bytes pushed value"}},
	PUSH26: {"PUSH26", nil, []string{"26 bytes pushed value"}},
	PUSH27: {"PUSH27", nil, []string{"27 bytes pushed value"}},
	PUSH28: {"PUSH28", nil, []string{"28 bytes pushed value"}},
	PUSH29: {"PUSH29", nil, []string{"29 bytes pushed value"}},
	PUSH30: {"PUSH30", nil, []string{"30 bytes pushed value"}},
	PUSH31: {"PUSH31", nil, []string{"31 bytes pushed value"}},
	PUSH32: {"PUSH32", nil, []string{"32 bytes pushed value"}},

	// cover your eyes, here comes ugly
	DUP1:  {"DUP1", []string{"x"}, []string{"x", "x"}},
	DUP2:  {"DUP2", []string{"-", "x"}, []string{"x", "-", "x"}},
	DUP3:  {"DUP3", []string{"-", "-", "x"}, []string{"x", "-", "x"}},
	DUP4:  {"DUP4", []string{"-", "-", "-", "x"}, []string{"x", "-", "-", "x"}},
	DUP5:  {"DUP5", []string{"-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "x"}},
	DUP6:  {"DUP6", []string{"-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "x"}},
	DUP7:  {"DUP7", []string{"-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "x"}},
	DUP8:  {"DUP8", []string{"-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "x"}},
	DUP9:  {"DUP9", []string{"-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP10: {"DUP10", []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP11: {"DUP11", []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP12: {"DUP12", []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP13: {"DUP13", []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP14: {"DUP14", []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP15: {"DUP15", []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP16: {"DUP16", []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},

	SWAP1:  {"SWAP1", []string{"a", "b"}, []string{"b", "a"}},
	SWAP2:  {"SWAP2", []string{"a", "", "b"}, []string{"b", "", "a"}},
	SWAP3:  {"SWAP3", []string{"a", "", "", "b"}, []string{"b", "", "", "a"}},
	SWAP4:  {"SWAP4", []string{"a", "", "", "", "b"}, []string{"b", "", "", "", "a"}},
	SWAP5:  {"SWAP5", []string{"a", "", "", "", "", "b"}, []string{"b", "", "", "", "", "a"}},
	SWAP6:  {"SWAP6", []string{"a", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "a"}},
	SWAP7:  {"SWAP7", []string{"a", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "a"}},
	SWAP8:  {"SWAP8", []string{"a", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "a"}},
	SWAP9:  {"SWAP9", []string{"a", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "a"}},
	SWAP10: {"SWAP10", []string{"a", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP11: {"SWAP11", []string{"a", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP12: {"SWAP12", []string{"a", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP13: {"SWAP13", []string{"a", "", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP14: {"SWAP14", []string{"a", "", "", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP15: {"SWAP15", []string{"a", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP16: {"SWAP16", []string{"a", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "a"}},

	LOG0: {"LOG0", []string{"mStart", "mSize"}, nil},
	LOG1: {"LOG1", []string{"mStart", "mSize", "topic"}, nil},
	LOG2: {"LOG2", []string{"mStart", "mSize", "topic", "topic"}, nil},
	LOG3: {"LOG3", []string{"mStart", "mSize", "topic", "topic", "topic"}, nil},
	LOG4: {"LOG4", []string{"mStart", "mSize", "topic", "topic", "topic", "topic"}, nil},

	// 0xf0 range.
	CREATE:       {"CREATE", []string{"value", "mem offset", "mem size"}, []string{"address or zero"}},
	CALL:         {"CALL", []string{"gas", "address", "value", "in offset", "in size", "out offset", "out size"}, []string{"exitcode (1 for success)"}},
	RETURN:       {"RETURN", []string{"offset", "size"}, nil},
	CALLCODE:     {"CALLCODE", []string{"gas", "address", "value", "in offset", "in size", "out offset", "out size"}, []string{"exitcode (1 for success)"}},
	DELEGATECALL: {"DELEGATECALL", []string{"gas", "address", "in offset", "in size", "out offset", "out size"}, []string{"exitcode (1 for success)"}},
	CREATE2:      {"CREATE2", []string{"value", "mem offset", "mem size", "salt"}, []string{"address or zero"}},
	STATICCALL:   {"STATICCALL", []string{"gas", "address", "in offset", "in size", "out offset", "out size"}, []string{"exitcode (1 for success)"}},
	REVERT:       {"REVERT", []string{"offset", "size"}, nil},
	SELFDESTRUCT: {"SELFDESTRUCT", []string{"beneficiary address"}, nil},
}

func (op OpCode) Pops() []string {
	info, exist := opCodeInfo[op]
	if !exist {
		return nil
	}
	return info.pops
}

func (op OpCode) Pushes() []string {
	info, exist := opCodeInfo[op]
	if !exist {
		return nil
	}
	return info.pushes
}

func (op OpCode) Stackdelta() int {
	return len(op.Pushes()) - len(op.Pops())
}
