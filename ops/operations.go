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

// Package ops contain definitions of the ethereum opcodes and forks.
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

// HasImmediate returns true if the op has immediate after the op.
func (op OpCode) HasImmediate() bool {
	return opCodeInfo[op].immediates > 0
}

func (op OpCode) IsCall() bool {
	return op == CALL ||
		op == DELEGATECALL ||
		op == CALLCODE ||
		op == STATICCALL ||
		op == EXTCALL ||
		op == EXTSTATICCALL ||
		op == EXTDELEGATECALL

}

func (op OpCode) ImmediateSize() int {
	return opCodeInfo[op].immediates // FIXME for RJUMPV it is variable based on byte 1
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

	KECCAK256 = OpCode(0x20)
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
	BLOBHASH    = OpCode(0x49)
	BLOBBASEFEE = OpCode(0x4a)
)

// 0x50 range - 'storage' and execution.
const (
	POP      = OpCode(0x50)
	MLOAD    = OpCode(0x51)
	MSTORE   = OpCode(0x52)
	MSTORE8  = OpCode(0x53)
	SLOAD    = OpCode(0x54)
	SSTORE   = OpCode(0x55)
	JUMP     = OpCode(0x56)
	JUMPI    = OpCode(0x57)
	PC       = OpCode(0x58)
	MSIZE    = OpCode(0x59)
	GAS      = OpCode(0x5A)
	JUMPDEST = OpCode(0x5B)

	TLOAD  = OpCode(0x5c) // Cancun
	TSTORE = OpCode(0x5d) // Cancun
	MCOPY  = OpCode(0x5e) // Cancun
	PUSH0  = OpCode(0x5f) // Shanghai

)

// 0x60 through 0x7F range.
const (
	PUSH1  = OpCode(0x60)
	PUSH2  = OpCode(0x61)
	PUSH3  = OpCode(0x62)
	PUSH4  = OpCode(0x63)
	PUSH5  = OpCode(0x64)
	PUSH6  = OpCode(0x65)
	PUSH7  = OpCode(0x66)
	PUSH8  = OpCode(0x67)
	PUSH9  = OpCode(0x68)
	PUSH10 = OpCode(0x69)
	PUSH11 = OpCode(0x6a)
	PUSH12 = OpCode(0x6b)
	PUSH13 = OpCode(0x6c)
	PUSH14 = OpCode(0x6d)
	PUSH15 = OpCode(0x6e)
	PUSH16 = OpCode(0x6f)
	PUSH17 = OpCode(0x70)
	PUSH18 = OpCode(0x71)
	PUSH19 = OpCode(0x72)
	PUSH20 = OpCode(0x73)
	PUSH21 = OpCode(0x74)
	PUSH22 = OpCode(0x75)
	PUSH23 = OpCode(0x76)
	PUSH24 = OpCode(0x77)
	PUSH25 = OpCode(0x78)
	PUSH26 = OpCode(0x79)
	PUSH27 = OpCode(0x7a)
	PUSH28 = OpCode(0x7b)
	PUSH29 = OpCode(0x7c)
	PUSH30 = OpCode(0x7d)
	PUSH31 = OpCode(0x7e)
	PUSH32 = OpCode(0x7f)
)

// 0x80 range
const (
	DUP1  = OpCode(0x80)
	DUP2  = OpCode(0x81)
	DUP3  = OpCode(0x82)
	DUP4  = OpCode(0x83)
	DUP5  = OpCode(0x84)
	DUP6  = OpCode(0x85)
	DUP7  = OpCode(0x86)
	DUP8  = OpCode(0x87)
	DUP9  = OpCode(0x88)
	DUP10 = OpCode(0x89)
	DUP11 = OpCode(0x8a)
	DUP12 = OpCode(0x8b)
	DUP13 = OpCode(0x8c)
	DUP14 = OpCode(0x8d)
	DUP15 = OpCode(0x8e)
	DUP16 = OpCode(0x8f)
)

// 0x90 range
const (
	SWAP1  = OpCode(0x90)
	SWAP2  = OpCode(0x91)
	SWAP3  = OpCode(0x92)
	SWAP4  = OpCode(0x93)
	SWAP5  = OpCode(0x94)
	SWAP6  = OpCode(0x95)
	SWAP7  = OpCode(0x96)
	SWAP8  = OpCode(0x97)
	SWAP9  = OpCode(0x98)
	SWAP10 = OpCode(0x99)
	SWAP11 = OpCode(0x9a)
	SWAP12 = OpCode(0x9b)
	SWAP13 = OpCode(0x9c)
	SWAP14 = OpCode(0x9d)
	SWAP15 = OpCode(0x9e)
	SWAP16 = OpCode(0x9f)
)

// 0xa0 range - logging ops.
const (
	LOG0 = OpCode(0xa0)
	LOG1 = OpCode(0xa1)
	LOG2 = OpCode(0xa2)
	LOG3 = OpCode(0xa3)
	LOG4 = OpCode(0xa4)
)

// 0xd0 range - EOF data operations
const (
	DATALOAD  = OpCode(0xd0) // osaka
	DATALOADN = OpCode(0xd1) // osaka
	DATASIZE  = OpCode(0xd2) // osaka
	DATACOPY  = OpCode(0xd3) // osaka
)

// 0xe0 range - EOF opcodes with immediates
const (
	RJUMP          = OpCode(0xe0) // osaka
	RJUMPI         = OpCode(0xe1) // osaka
	RJUMPV         = OpCode(0xe2) // osaka
	CALLF          = OpCode(0xe3) // osaka
	RETF           = OpCode(0xe4) // osaka
	JUMPF          = OpCode(0xe5) // osaka
	DUPN           = OpCode(0xe6) // osaka
	SWAPN          = OpCode(0xe7) // osaka
	EXCHANGE       = OpCode(0xe8) // osaka
	EOFCREATE      = OpCode(0xec) // osaka
	RETURNCONTRACT = OpCode(0xee) // osaka
)

// 0xf0 range - closures.
const (
	CREATE       = OpCode(0xf0)
	CALL         = OpCode(0xf1)
	CALLCODE     = OpCode(0xf2)
	RETURN       = OpCode(0xf3)
	DELEGATECALL = OpCode(0xf4)
	CREATE2      = OpCode(0xf5)

	RETURNDATALOAD  = OpCode(0xf7) // osaka
	EXTCALL         = OpCode(0xf8) // osaka
	EXTDELEGATECALL = OpCode(0xf9) // osaka
	STATICCALL      = OpCode(0xfa)
	EXTSTATICCALL   = OpCode(0xfb) // osaka

	REVERT       = OpCode(0xfd)
	INVALID      = OpCode(0xfe)
	SELFDESTRUCT = OpCode(0xff)
)

func (op OpCode) String() string {
	if info, ok := opCodeInfo[op]; ok {
		return info.name
	}
	return fmt.Sprintf("opcode 0x%x not defined", int(op))
}

func IsDefined(op OpCode) bool {
	_, ok := opCodeInfo[op]
	return ok
}

// stringToOp is a mapping from strings to OpCode
var stringToOp map[string]OpCode

func init() {
	stringToOp = make(map[string]OpCode)
	for k, elem := range opCodeInfo {
		stringToOp[elem.name] = k
	}
	// Add mapping for legacy opcode names
	stringToOp["SHA3"] = KECCAK256
	stringToOp["SUICIDE"] = SELFDESTRUCT
}

// StringToOp finds the opcode whose name is stored in `str`.
func StringToOp(str string) OpCode {
	return stringToOp[str]
}

type opInfo struct {
	name       string
	immediates int
	pops       []string
	pushes     []string
}

var opCodeInfo = map[OpCode]opInfo{

	STOP:       {"STOP", 0, nil, nil},
	ADD:        {"ADD", 0, []string{"a", "b"}, []string{"a + b"}},
	MUL:        {"MUL", 0, []string{"a", "b"}, []string{"a * b"}},
	SUB:        {"SUB", 0, []string{"a", "b"}, []string{"a - b"}},
	DIV:        {"DIV", 0, []string{"a", "b"}, []string{"a / b"}},
	SDIV:       {"SDIV", 0, []string{"a", "b"}, []string{"a / b (signed)"}},
	MOD:        {"MOD", 0, []string{"a", "b"}, []string{"a % b"}},
	SMOD:       {"SMOD", 0, []string{"a", "b"}, []string{"a mod b (signed)"}},
	EXP:        {"EXP", 0, []string{"base", "exp"}, []string{"base^exp"}},
	NOT:        {"NOT", 0, []string{"a"}, []string{"not(a)"}},
	LT:         {"LT", 0, []string{"a", "b"}, []string{"a < b"}},
	GT:         {"GT", 0, []string{"a", "b"}, []string{"a > b"}},
	SLT:        {"SLT", 0, []string{"a", "b"}, []string{"a < b (signed)"}},
	SGT:        {"SGT", 0, []string{"a", "b"}, []string{"a > b (signed)"}},
	EQ:         {"EQ", 0, []string{"a", "b"}, []string{"a == b"}},
	ISZERO:     {"ISZERO", 0, []string{"a"}, []string{"a == 0"}},
	SIGNEXTEND: {"SIGNEXTEND", 0, []string{"bitlen", "a"}, []string{"signextend(a, bitlen)"}},

	AND:    {"AND", 0, []string{"a", "b"}, []string{"a && b"}},
	OR:     {"OR", 0, []string{"a", "b"}, []string{"a || b"}},
	XOR:    {"XOR", 0, []string{"a", "b"}, []string{"a xor b"}},
	BYTE:   {"BYTE", 0, []string{"index", "val"}, []string{"byte at val[index]"}},
	SHL:    {"SHL", 0, []string{"shift", "x"}, []string{"x << shift"}},
	SHR:    {"SHR", 0, []string{"shift", "x"}, []string{"x >> shift"}},
	SAR:    {"SAR", 0, []string{"shift", "x"}, []string{"x >>> shift"}},
	ADDMOD: {"ADDMOD", 0, []string{"a", "b", "x"}, []string{"(a + b) mod x"}},
	MULMOD: {"MULMOD", 0, []string{"a", "b", "x"}, []string{"(a * b) mod x"}},

	// 0x20 range - crypto.
	KECCAK256: {"KECCAK256", 0, []string{"offset", "size"}, []string{"keccak256(mem[offset:offset+size])"}},
	// 0x30 range - closure state.
	ADDRESS:      {"ADDRESS", 0, nil, []string{"address of current context"}},
	BALANCE:      {"BALANCE", 0, []string{"address"}, []string{"balance of address"}},
	ORIGIN:       {"ORIGIN", 0, nil, []string{"transaction origin"}},
	CALLER:       {"CALLER", 0, nil, []string{"sender"}},
	CALLVALUE:    {"CALLVALUE", 0, nil, []string{"call value"}},
	CALLDATALOAD: {"CALLDATALOAD", 0, []string{"offset"}, []string{"calldata[offset:offset+32]"}},
	CALLDATASIZE: {"CALLDATASIZE", 0, nil, []string{"size of calldata"}},
	CALLDATACOPY: {"CALLDATACOPY", 0, []string{"memOffset", "dataOffset", "length"}, nil},
	CODESIZE:     {"CODESIZE", 0, nil, []string{"size of code in this context"}},
	CODECOPY:     {"CODECOPY", 0, []string{"memOffset", "codeOffset", "length"}, nil},
	GASPRICE:     {"GASPRICE", 0, nil, []string{"transaction gasprice"}},

	EXTCODESIZE: {"EXTCODESIZE", 0, []string{"address"}, []string{"code size at 'address'"}},
	EXTCODECOPY: {"EXTCODECOPY", 0, []string{"address", "memOffset", "codeOffset", "length"}, nil},

	RETURNDATASIZE: {"RETURNDATASIZE", 0, nil, []string{"size of returndata"}},
	RETURNDATACOPY: {"RETURNDATACOPY", 0, []string{"memOffset", "dataOffset", "length"}, nil},
	EXTCODEHASH:    {"EXTCODEHASH", 0, []string{"address"}, []string{"codehash at 'address'"}},

	// 0x40 range - block operations.
	BLOCKHASH:   {"BLOCKHASH", 0, []string{"blocknum"}, []string{"hash of block at blocknum"}},
	COINBASE:    {"COINBASE", 0, nil, []string{"block miner address"}},
	TIMESTAMP:   {"TIMESTAMP", 0, nil, []string{"unix time of current block"}},
	NUMBER:      {"NUMBER", 0, nil, []string{"current block number"}},
	DIFFICULTY:  {"DIFFICULTY", 0, nil, []string{"current block difficulty"}},
	GASLIMIT:    {"GASLIMIT", 0, nil, []string{"block gas limit"}},
	CHAINID:     {"CHAINID", 0, nil, []string{"chain id"}},
	SELFBALANCE: {"SELFBALANCE", 0, nil, []string{"balance at current context"}},
	BASEFEE:     {"BASEFEE", 0, nil, []string{"basefee in current block"}},
	BLOBHASH:    {"BLOBHASH", 0, []string{"index"}, []string{"blobhash at index"}},
	BLOBBASEFEE: {"BLOBBASEFEE", 0, nil, []string{"blob basefee in current block"}},

	POP:      {"POP", 0, []string{"value to pop"}, nil},
	MLOAD:    {"MLOAD", 0, []string{"offset"}, []string{"value"}},
	MSTORE:   {"MSTORE", 0, []string{"offset", "value"}, nil},
	MSTORE8:  {"MSTORE8", 0, []string{"offset", "value"}, nil},
	SLOAD:    {"SLOAD", 0, []string{"slot"}, []string{"value"}},
	SSTORE:   {"SSTORE", 0, []string{"slot", "value"}, nil},
	JUMP:     {"JUMP", 0, []string{"loc"}, nil},
	JUMPI:    {"JUMPI", 0, []string{"loc", "cond"}, nil},
	PC:       {"PC", 0, nil, []string{"current PC"}},
	MSIZE:    {"MSIZE", 0, nil, []string{"size of memory"}},
	GAS:      {"GAS", 0, nil, []string{"current gas remaining"}},
	JUMPDEST: {"JUMPDEST", 0, nil, nil},
	MCOPY:    {"MCOPY", 0, []string{"dest", "source", "length"}, nil},
	TLOAD:    {"TLOAD", 0, []string{"t-slot"}, []string{"value"}},
	TSTORE:   {"TSTORE", 0, []string{"t-slot", "value"}, nil},
	PUSH0:    {"PUSH0", 0, nil, []string{"zero"}},

	// 0x60 through 0x7F range - push.
	PUSH1:  {"PUSH1", 1, nil, []string{"1 byte pushed value"}},
	PUSH2:  {"PUSH2", 2, nil, []string{"2 bytes pushed value"}},
	PUSH3:  {"PUSH3", 3, nil, []string{"3 bytes pushed value"}},
	PUSH4:  {"PUSH4", 4, nil, []string{"4 bytes pushed value"}},
	PUSH5:  {"PUSH5", 5, nil, []string{"5 bytes pushed value"}},
	PUSH6:  {"PUSH6", 6, nil, []string{"6 bytes pushed value"}},
	PUSH7:  {"PUSH7", 7, nil, []string{"7 bytes pushed value"}},
	PUSH8:  {"PUSH8", 8, nil, []string{"8 bytes pushed value"}},
	PUSH9:  {"PUSH9", 9, nil, []string{"9 bytes pushed value"}},
	PUSH10: {"PUSH10", 10, nil, []string{"10 bytes pushed value"}},
	PUSH11: {"PUSH11", 11, nil, []string{"11 bytes pushed value"}},
	PUSH12: {"PUSH12", 12, nil, []string{"12 bytes pushed value"}},
	PUSH13: {"PUSH13", 13, nil, []string{"13 bytes pushed value"}},
	PUSH14: {"PUSH14", 14, nil, []string{"14 bytes pushed value"}},
	PUSH15: {"PUSH15", 15, nil, []string{"15 bytes pushed value"}},
	PUSH16: {"PUSH16", 16, nil, []string{"16 bytes pushed value"}},
	PUSH17: {"PUSH17", 17, nil, []string{"17 bytes pushed value"}},
	PUSH18: {"PUSH18", 18, nil, []string{"18 bytes pushed value"}},
	PUSH19: {"PUSH19", 19, nil, []string{"19 bytes pushed value"}},
	PUSH20: {"PUSH20", 20, nil, []string{"19 bytes pushed value"}},
	PUSH21: {"PUSH21", 21, nil, []string{"21 bytes pushed value"}},
	PUSH22: {"PUSH22", 22, nil, []string{"22 bytes pushed value"}},
	PUSH23: {"PUSH23", 23, nil, []string{"23 bytes pushed value"}},
	PUSH24: {"PUSH24", 24, nil, []string{"24 bytes pushed value"}},
	PUSH25: {"PUSH25", 25, nil, []string{"25 bytes pushed value"}},
	PUSH26: {"PUSH26", 26, nil, []string{"26 bytes pushed value"}},
	PUSH27: {"PUSH27", 27, nil, []string{"27 bytes pushed value"}},
	PUSH28: {"PUSH28", 28, nil, []string{"28 bytes pushed value"}},
	PUSH29: {"PUSH29", 29, nil, []string{"29 bytes pushed value"}},
	PUSH30: {"PUSH30", 30, nil, []string{"30 bytes pushed value"}},
	PUSH31: {"PUSH31", 31, nil, []string{"31 bytes pushed value"}},
	PUSH32: {"PUSH32", 32, nil, []string{"32 bytes pushed value"}},

	// cover your eyes, here comes ugly
	DUP1:  {"DUP1", 0, []string{"x"}, []string{"x", "x"}},
	DUP2:  {"DUP2", 0, []string{"-", "x"}, []string{"x", "-", "x"}},
	DUP3:  {"DUP3", 0, []string{"-", "-", "x"}, []string{"x", "-", "-", "x"}},
	DUP4:  {"DUP4", 0, []string{"-", "-", "-", "x"}, []string{"x", "-", "-", "-", "x"}},
	DUP5:  {"DUP5", 0, []string{"-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "x"}},
	DUP6:  {"DUP6", 0, []string{"-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "x"}},
	DUP7:  {"DUP7", 0, []string{"-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "x"}},
	DUP8:  {"DUP8", 0, []string{"-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP9:  {"DUP9", 0, []string{"-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP10: {"DUP10", 0, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP11: {"DUP11", 0, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP12: {"DUP12", 0, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP13: {"DUP13", 0, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP14: {"DUP14", 0, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP15: {"DUP15", 0, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},
	DUP16: {"DUP16", 0, []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}, []string{"x", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "x"}},

	SWAP1:  {"SWAP1", 0, []string{"a", "b"}, []string{"b", "a"}},
	SWAP2:  {"SWAP2", 0, []string{"a", "", "b"}, []string{"b", "", "a"}},
	SWAP3:  {"SWAP3", 0, []string{"a", "", "", "b"}, []string{"b", "", "", "a"}},
	SWAP4:  {"SWAP4", 0, []string{"a", "", "", "", "b"}, []string{"b", "", "", "", "a"}},
	SWAP5:  {"SWAP5", 0, []string{"a", "", "", "", "", "b"}, []string{"b", "", "", "", "", "a"}},
	SWAP6:  {"SWAP6", 0, []string{"a", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "a"}},
	SWAP7:  {"SWAP7", 0, []string{"a", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "a"}},
	SWAP8:  {"SWAP8", 0, []string{"a", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "a"}},
	SWAP9:  {"SWAP9", 0, []string{"a", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "a"}},
	SWAP10: {"SWAP10", 0, []string{"a", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP11: {"SWAP11", 0, []string{"a", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP12: {"SWAP12", 0, []string{"a", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP13: {"SWAP13", 0, []string{"a", "", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP14: {"SWAP14", 0, []string{"a", "", "", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP15: {"SWAP15", 0, []string{"a", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "a"}},
	SWAP16: {"SWAP16", 0, []string{"a", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "b"}, []string{"b", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "a"}},

	LOG0: {"LOG0", 0, []string{"mStart", "mSize"}, nil},
	LOG1: {"LOG1", 0, []string{"mStart", "mSize", "topic"}, nil},
	LOG2: {"LOG2", 0, []string{"mStart", "mSize", "topic", "topic"}, nil},
	LOG3: {"LOG3", 0, []string{"mStart", "mSize", "topic", "topic", "topic"}, nil},
	LOG4: {"LOG4", 0, []string{"mStart", "mSize", "topic", "topic", "topic", "topic"}, nil},

	DATALOAD:       {"DATALOAD", 0, []string{"offset"}, []string{"data"}},
	DATALOADN:      {"DATALOADN", 1, nil, []string{"data"}},
	DATASIZE:       {"DATASIZE", 0, nil, []string{"data"}},
	DATACOPY:       {"DATACOPY", 0, []string{"mem_offset", "offset", "size"}, nil},
	RJUMP:          {"RJUMP", 2, nil, nil},
	RJUMPI:         {"RJUMPI", 2, []string{"condition"}, nil},
	RJUMPV:         {"RJUMPV", 3, []string{"case"}, nil},
	CALLF:          {"CALLF", 2, nil, []string{"0 or more items, depends on type headers"}},
	RETF:           {"RETF", 0, nil, nil},
	JUMPF:          {"JUMPF", 2, nil, nil},
	DUPN:           {"DUPN", 1, nil, nil},
	SWAPN:          {"SWAPN", 1, nil, nil},
	EXCHANGE:       {"EXCHANGE", 1, nil, nil},
	EOFCREATE:      {"EOFCREATE", 2, []string{"value", "salt", "input_offset", "input_size"}, []string{"address"}},
	RETURNCONTRACT: {"RETURNCONTRACT", 2, []string{"aux_data_offset", "aux_data_size"}, nil},

	// 0xf0 range.
	CREATE:          {"CREATE", 0, []string{"value", "mem offset", "mem size"}, []string{"address or zero"}},
	CALL:            {"CALL", 0, []string{"gas", "address", "value", "in offset", "in size", "out offset", "out size"}, []string{"exitcode (1 for success)"}},
	RETURN:          {"RETURN", 0, []string{"offset", "size"}, nil},
	CALLCODE:        {"CALLCODE", 0, []string{"gas", "address", "value", "in offset", "in size", "out offset", "out size"}, []string{"exitcode (1 for success)"}},
	DELEGATECALL:    {"DELEGATECALL", 0, []string{"gas", "address", "in offset", "in size", "out offset", "out size"}, []string{"exitcode (1 for success)"}},
	CREATE2:         {"CREATE2", 0, []string{"value", "mem offset", "mem size", "salt"}, []string{"address or zero"}},
	RETURNDATALOAD:  {"RETURNDATALOAD", 0, []string{"offset"}, []string{"data"}},
	EXTCALL:         {"EXTCALL", 0, []string{"target_address", "input_offset", "input_size", "value"}, []string{"status"}},
	EXTDELEGATECALL: {"EXTDELEGATECALL", 0, []string{"target_address", "input_offset", "input_size"}, []string{"status"}},
	STATICCALL:      {"STATICCALL", 0, []string{"gas", "address", "in offset", "in size", "out offset", "out size"}, []string{"exitcode (1 for success)"}},
	EXTSTATICCALL:   {"EXTSTATICCALL", 0, []string{"target_address", "input_offset", "input_size"}, []string{"status"}},
	REVERT:          {"REVERT", 0, []string{"offset", "size"}, nil},
	INVALID:         {"INVALID", 0, nil, nil},
	SELFDESTRUCT:    {"SELFDESTRUCT", 0, []string{"beneficiary address"}, nil},
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

func (op OpCode) ExpandsMem() bool {
	if op < KECCAK256 {
		return false
	}
	switch op {
	case KECCAK256, CALLDATALOAD, CALLDATASIZE, CALLDATACOPY, CODECOPY,
		EXTCODECOPY, RETURNDATACOPY,
		MLOAD, MSTORE, MSTORE8, LOG0, LOG1, LOG2, LOG3, LOG4,
		CREATE, CALL, DELEGATECALL, CALLCODE, STATICCALL, RETURN, REVERT, CREATE2:
		return true
	default:
		return false
	}
}
