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

package traces

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/golang/snappy"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/uint256"
)

type TraceLine struct {
	step      uint64
	address   *common.Address
	callStack []*callInfo
	log       *logger.StructLog
}

type Traces struct {
	Ops  []*TraceLine
	Errs []string
}

// ChunkSize is global variable
var ChunkSize = uint64(31)

func (traces *Traces) Get(index int) *TraceLine {
	if index < len(traces.Ops) && index >= 0 {
		return traces.Ops[index]
	}
	return nil
}

func (traces *Traces) Search(op string, from int) (*TraceLine, int) {
	if from >= len(traces.Ops) {
		return nil, 0
	}
	for i := from; i < len(traces.Ops); i++ {
		t := traces.Ops[i]
		if strings.HasPrefix(t.log.Op.String(), op) {
			return t, i
		}
	}
	return nil, 0
}

func (t *TraceLine) Get(title string) string {
	op := t.log
	switch strings.ToLower(title) {
	case "step":
		return fmt.Sprintf("%d", t.step)
	case "chunk":
		return fmt.Sprintf("%v (0x%x)", op.Pc/ChunkSize, op.Pc/ChunkSize)
	case "pc":
		return fmt.Sprintf("%v (0x%x)", op.Pc, op.Pc)
	// Depends on Geth EOF support
	//case "section":
	//	return fmt.Sprintf("%v", op.Section)
	case "opname":
		return op.OpName()
	case "opcode":
		return fmt.Sprintf("0x%x", byte(op.Op))
	case "gas":
		return fmt.Sprintf("%d", op.Gas)
	case "gascost":
		return fmt.Sprintf("%d", op.GasCost)
	case "depth":
		return fmt.Sprintf("%d", op.Depth)
	// Depends on Geth EOF support
	//case "functiondepth":
	//	return fmt.Sprintf("%d", op.FunctionDepth)
	case "refund":
		return fmt.Sprintf("%d", op.RefundCounter)
	case "memsize":
		return fmt.Sprintf("%d", op.MemorySize)
	case "address", "addr":
		if t.address != nil {
			return t.address.Hex()
		}
	}
	return "NA"
}

func (t *TraceLine) Stack() []uint256.Int {
	return t.log.Stack
}

func (t *TraceLine) Memory() []byte {
	return t.log.Memory
}

func (t *TraceLine) Op() byte {
	return byte(t.log.Op)
}
func (t *TraceLine) Step() uint64 {
	return t.step
}

func (t *TraceLine) Depth() int {
	return t.log.Depth
}

func (t *TraceLine) CallStack() []*callInfo {
	return t.callStack
}

func (t *TraceLine) Source() string {
	x, _ := json.Marshal(t.log)
	return string(x)
}

func (t *TraceLine) Equals(other *TraceLine) bool {
	if t.Op() != other.Op() ||
		t.log.Pc != other.log.Pc ||
		t.log.Depth != other.log.Depth ||
		len(t.log.Stack) != len(other.log.Stack) ||
		t.log.Gas != other.log.Gas {
		return false
	}
	// EIP-7756 fields.  If both are non-zero they must match
	// Depends on Geth EOF support
	//if (t.log.Section != 0 && other.log.Section != 0 && t.log.Section != other.log.Section) ||
	//	(t.log.FunctionDepth != 0 && other.log.FunctionDepth != 0 && t.log.FunctionDepth != other.log.FunctionDepth) {
	//	return false
	//}
	// Also inspect stack
	for i, elem := range t.log.Stack {
		if elem != other.log.Stack[i] {
			return false
		}
	}
	return true
	//t.Get("depth") == other.Get("pc")
}

func convertToStructLog(op map[string]interface{}) (*logger.StructLog, error) {
	intify := func(value interface{}) int {
		// Try to convert it
		if floatVal, ok := value.(float64); ok {
			return int(floatVal)
		}
		var (
			retval int64
			err    error
		)

		// Maybe a hexvalue?
		if strVal, ok := value.(string); ok {
			if len(strVal) > 1 && strVal[0:2] == "0x" {
				retval, err = strconv.ParseInt(strVal[2:], 16, 64)
			} else {
				retval, err = strconv.ParseInt(strVal, 10, 64)
			}
		} else {
			err = fmt.Errorf("could not convert %v to int", value)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return int(retval)
	}

	log := &logger.StructLog{}
	ok := false
	for k, v := range op {
		switch k {
		case "pc":
			log.Pc = uint64(intify(v))
		// Depends on Geth EOF support
		//case "section":
		//	log.Section = uint64(intify(v))
		case "memSize":
			log.MemorySize = intify(v)
		case "op":
			log.Op = vm.OpCode(uint64(intify(v)))
			ok = true
		case "gas":
			log.Gas = uint64(intify(v))
		case "gasCost":
			log.GasCost = uint64(intify(v))
		case "depth":
			log.Depth = int(v.(float64))
		// Depends on Geth EOF support
		//case "functionDepth":
		//	log.FunctionDepth = int(v.(float64))
		case "refund":
			log.RefundCounter = uint64(intify(v))
		case "stack":
			// v is a list of strings
			stack, err := parseStack(v.([]interface{}))
			if err != nil {
				return nil, err
			}
			log.Stack = stack
		case "memory":
			log.Memory = common.FromHex(v.(string))
		}
	}
	if ok {
		return log, nil
	}
	return nil, fmt.Errorf("incomplete op")
}

type traceTxLog struct {
	Pc      uint64
	GasCost uint64
	Stack   []interface{}
	// Note, traceTransaction uses 'op' for the human-readable name
	Op     string
	Depth  uint64
	Gas    uint64
	Memory []interface{}
}

type traceTxResult struct {
	Logs []traceTxLog `json:"structLogs"`
	// + some other fields we don't particularly care about
}
type traceTxRPCResponse struct {
	Result traceTxResult `json:"result"`
	// + some other fields we don't particularly care about
}

// ParseHex parses s as a 256 bit integer in hexadecimal syntax.
// Leading zeros are accepted. The empty string parses as zero.
func ParseHex(s string) (uint256.Int, error) {
	var n uint256.Int
	if s == "" {
		return n, nil
	}
	var bigint *big.Int
	var ok bool
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		bigint, ok = new(big.Int).SetString(s[2:], 16)
	} else {
		bigint, ok = new(big.Int).SetString(s, 16)
	}
	if !ok {
		return n, fmt.Errorf("could not convert %v to bigint", s)
	}
	if overflow := n.SetFromBig(bigint); overflow {
		return n, fmt.Errorf("conversion from bigint (%x) to uint256.Int caused overflow", bigint)
	}
	return n, nil
}

// parseStack takes a list of strings and returns a stack of *big.Ints
func parseStack(stackStrings []interface{}) ([]uint256.Int, error) {
	var (
		s []uint256.Int
	)
	for _, item := range stackStrings {
		n, err := ParseHex(item.(string))
		if err != nil {
			return nil, fmt.Errorf("parsing failed: %v", err)
		}
		s = append(s, n)
	}
	// reverse it
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s, nil
}

// parseMem takes a list of strings bundles them together into one nice
// byte array
func parseMem(memStrings []interface{}) []byte {
	s := make([]byte, 0, len(memStrings)*32)

	for _, item := range memStrings {
		data := common.FromHex(item.(string))
		s = append(s, data...)
	}
	return s
}

// readJSON attempts to slurp the file as a JSON file
func readJSON(data []byte) (*Traces, error) {
	var (
		traceData traceTxRPCResponse
		traces    Traces
	)
	// Attempt one: read directly into traceTxResult,
	// This will succeed if the file consist of the actual
	// 'result', but not the full RPC response
	err := json.Unmarshal(data, &traceData.Result)
	if err != nil {
		if err != nil {
			return nil, err
		}
	}
	if traceData.Result.Logs == nil {
		// Attempt two: read into traceTxRPCResponse, in case
		// the file is the complete RPC response from a
		// traceTransaction invocation
		err = json.Unmarshal(data, &traceData)
		if err != nil {
			return nil, err
		}
	}

	for step, log := range traceData.Result.Logs {
		structLog := &logger.StructLog{
			Depth:   int(log.Depth),
			Pc:      log.Pc,
			GasCost: log.GasCost,
			Gas:     log.Gas,
			Op:      vm.OpCode(ops.StringToOp(log.Op)),
		}
		stack, err := parseStack(log.Stack)
		if err != nil {
			return nil, err
		}
		structLog.Stack = stack
		structLog.Memory = parseMem(log.Memory)
		traces.Ops = append(traces.Ops, &TraceLine{
			step: uint64(step),
			log:  structLog,
		})

	}
	return &traces, nil
}

// readJSONLines attempts to read the file as json-lines, line by line
// delimited json objects
func readJSONLines(input io.Reader) (*Traces, error) {

	var traces Traces
	step := uint64(0)
	// Read line by line
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		l := scanner.Text()
		obj := make(map[string]interface{})

		if err := json.Unmarshal([]byte(l), &obj); err != nil {
			// An error here means it's not valid jsonl
			return nil, err
		}
		if log, err := convertToStructLog(obj); err != nil {
			// An error here just means it's not what we expected
			traces.Errs = append(traces.Errs, err.Error())
		} else {
			traces.Ops = append(traces.Ops, &TraceLine{
				log:  log,
				step: step,
			})
		}
		step++
		if strings.HasPrefix(l, `{"stateRoot"`) {
			// We're done, nothing more here
			break
		}

	}
	if err := scanner.Err(); err != nil {
		traces.Errs = append(traces.Errs, err.Error())
	}
	return &traces, nil

}

// ReadFile reads a trace from either a json-lines file or json-file, optionally
// snappy encoded
func ReadFile(location string) (*Traces, error) {
	var (
		err  error
		data []byte
	)
	data, err = os.ReadFile(location)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(location, "snappy") {
		data, err = snappy.Decode(nil, data)
		if err != nil {
			return nil, err
		}
	}
	if strings.HasSuffix(location, ".json") {
		// read as json
		t, err := readJSON(data)
		// Do a second pass to assign addresses, where applicable
		if err == nil {
			AnalyzeCalls(t)
		}
		return t, err
	}

	// First attempt to read as JSON struct
	t, err := readJSON(data)
	if err != nil {
		// Second attempt, read as json lines.
		// Need to reset the input
		t, err = readJSONLines(bytes.NewReader(data))
	}
	// Do a second pass to assign addresses, where applicable
	if err == nil {
		AnalyzeCalls(t)
	}
	return t, err
}
