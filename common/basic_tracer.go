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
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"
)

/*
var (
	// compile time type check
	_ tracers.Tracer = (*BasicTracer)(nil)
)*/

type BasicTracer struct{}

func (n *BasicTracer) CaptureTxStart(uint64) {}
func (n *BasicTracer) CaptureTxEnd(uint64)   {}
func (n *BasicTracer) CaptureStart(*vm.EVM, common.Address, common.Address, bool, []byte, uint64, *big.Int) {
}
func (n *BasicTracer) CaptureEnd([]byte, uint64, error) {}
func (n *BasicTracer) CaptureEnter(vm.OpCode, common.Address, common.Address, []byte, uint64, *big.Int) {
}
func (n *BasicTracer) CaptureExit([]byte, uint64, error) {}
func (n *BasicTracer) CaptureState(uint64, vm.OpCode, uint64, uint64, *vm.ScopeContext, []byte, int, error) {
}
func (n *BasicTracer) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
	fmt.Printf("CaptureFault %v\n", err)
}

func (n *BasicTracer) GetResult() (json.RawMessage, error) {
	return nil, nil
}
func (n *BasicTracer) Stop(err error) {}
