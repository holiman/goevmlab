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

// Package traces contain some helper-utils to visualise/track traces.
package traces

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
)

type callInfo struct {
	Ctx  *common.Address
	Dest *common.Address
	Kind string
}

func (info *callInfo) String() string {
	var (
		cur  = "NA"
		dest = "NA"
	)
	if info.Ctx != nil {
		cur = fmt.Sprintf("0x%x...", info.Ctx[:8])
	}
	if info.Dest != nil {
		dest = fmt.Sprintf("0x%x...", info.Dest[:8])
	}
	if dest == cur {
		return fmt.Sprintf("%s to %v", info.Kind, dest)
	}
	return fmt.Sprintf("%s to %v (ctx: %v)", info.Kind, dest, cur)
}

// stack is an object for basic stack operations.
type stack struct {
	data []*callInfo
}

func newStack() *stack {
	return &stack{data: make([]*callInfo, 0, 5)}
}

func (st *stack) copy() *stack {
	cpy := &stack{data: make([]*callInfo, len(st.data))}
	copy(cpy.data, st.data)
	return cpy
}

func (st *stack) push(info *callInfo) {
	st.data = append(st.data, info)
}
func (st *stack) pop() (ret *callInfo) {
	if len(st.data) == 0 {
		return nil
	}
	ret = st.data[len(st.data)-1]
	st.data = st.data[:len(st.data)-1]
	return
}

func (st *stack) peek() (ret *callInfo) {
	if len(st.data) == 0 {
		return nil
	}
	return st.data[len(st.data)-1]
}

// AnalyzeCalls scans through the ops, and assigns context-addresses to the
// lines.
func AnalyzeCalls(trace *Traces) {
	callStack := newStack()
	var currentAddress *common.Address
	var prevLine *TraceLine
	for _, line := range trace.Ops {
		if prevLine != nil {
			if cDepth, pDepth := line.Depth(), prevLine.Depth(); cDepth != pDepth {
				// Make a new callstack
				callStack = callStack.copy()
				if cDepth > pDepth {
					// A call or create was made here
					newAddress, callDest, callName := determineDestination(prevLine.log, currentAddress)
					currentAddress = newAddress
					callStack.push(&callInfo{
						Ctx:  newAddress,
						Dest: callDest,
						Kind: callName,
					})
				} else {
					// We backed out
					callStack.pop()
					info := callStack.peek()
					if info != nil {
						currentAddress = info.Ctx
					} else {
						currentAddress = nil
					}
				}
			}
			line.address = currentAddress
			line.callStack = callStack.data
		}
		prevLine = line
	}
}

// determineDestination looks at the stack args and determines what the call
// address is. Returns:
// contextAddr -- the execution context
// callDest    -- the call destination
// name        -- type of call
func determineDestination(log *logger.StructLog, current *common.Address) (contextAddr, callDest *common.Address, name string) {
	switch log.Op {
	case vm.CALL:
		name = "CALL"
		if len(log.Stack) > 1 {
			a := common.Address(log.Stack[1].Bytes20())
			callDest = &a
			contextAddr = &a
		}
	case vm.STATICCALL:
		name = "SCALL"
		if len(log.Stack) > 1 {
			a := common.Address(log.Stack[1].Bytes20())
			callDest = &a
			contextAddr = &a
		}
	case vm.DELEGATECALL:
		// The stack index is 1, but the actual execution context remains the same
		name = "DCALL"
		if len(log.Stack) > 1 {
			a := common.Address(log.Stack[1].Bytes20())
			callDest = &a
			contextAddr = current
		}
	case vm.CALLCODE:
		// The stack index is 1, but the actual execution context remains the same
		name = "CCALL"
		if len(log.Stack) > 1 {
			a := common.Address(log.Stack[1].Bytes20())
			callDest = &a
			contextAddr = current
		}
	case vm.CREATE:
		// In order to figure this out, we would need both nonce and current address
		// while we _may_ have the address, we don't have the nonce
		name = "CREATE"
	case vm.CREATE2:
		// In order to figure this out, we needs salt, initcode and current address
		// while we _may_ theoretically be able to sort it out, by inspecting the
		// memory, it would be pretty flaky
		name = "CREATE2"
	}
	return contextAddr, callDest, name
}
