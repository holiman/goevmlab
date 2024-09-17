// Copyright 2022 Martin Holst Swende
// This file is part of the go-evmlab library.
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

package common

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"
	"strings"
)

/*
var (
	// compile time type check
	_ tracers.Tracer = (*BasicTracer)(nil)
)*/

type PrintingTracer struct {
	BasicTracer
}

func (n *PrintingTracer) CaptureStart(vm *vm.EVM, from, to common.Address, create bool, input []byte, gas uint64, value *big.Int) {
	fmt.Printf("Start: from %x to %x, value: %#x\n", from, to, value)
}
func (n *PrintingTracer) CaptureState(pc uint64, op vm.OpCode, gas uint64, cost uint64, scope *vm.ScopeContext, input []byte, depth int, err error) {
	var st []string
	for _, elem := range scope.Stack.Data() {
		st = append(st, elem.Hex())
	}
	var indent = " "
	for i := 1; i < depth; i++ {
		indent = indent + " "
	}
	fmt.Printf("%s pc %d, op %v, stack [%s]\n", indent, pc, op.String(), strings.Join(st, ","))
}
