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

// Package common contains some shared stuff
package common

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
)

type PrintingTracer struct{}

func (d *PrintingTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnOpcode: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, rData []byte, depth int, err error) {
			var st []string
			for _, elem := range scope.StackData() {
				st = append(st, elem.Hex())
			}
			var indent = " "
			for i := 1; i < depth; i++ {
				indent = indent + " "
			}
			fmt.Printf("%s pc %d, op %v, stack [%s]\n", indent, pc, vm.OpCode(op).String(), strings.Join(st, ","))
		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			fmt.Printf("CaptureFault %v\n", err)
		},
		OnTxStart: func(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
			fmt.Printf("Start: from %x to %x, value: %#x\n", from, *tx.To(), tx.Value())
		},
		OnTxEnd: func(receipt *types.Receipt, err error) {
			fmt.Printf("\nCaptureEnd\n")
		},
	}
}
