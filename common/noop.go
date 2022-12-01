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
	"math/big"
	"time"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/tracers"
)

var (
	// compile time type check
	_ tracers.Tracer = (*NoOpTracer)(nil)
)

type NoOpTracer struct{}

func (n *NoOpTracer) CaptureTxStart(uint64) {}
func (n *NoOpTracer) CaptureTxEnd(uint64)   {}
func (n *NoOpTracer) CaptureStart(*vm.EVM, common.Address, common.Address, bool, []byte, uint64, *big.Int) {
}
func (n *NoOpTracer) CaptureEnd([]byte, uint64, time.Duration, error) {}
func (n *NoOpTracer) CaptureEnter(vm.OpCode, common.Address, common.Address, []byte, uint64, *big.Int) {
}
func (n *NoOpTracer) CaptureExit([]byte, uint64, error) {}
func (n *NoOpTracer) CaptureState(uint64, vm.OpCode, uint64, uint64, *vm.ScopeContext, []byte, int, error) {
}
func (n *NoOpTracer) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
	fmt.Printf("CaptureFault %v\n", err)
}

func (n *NoOpTracer) GetResult() (json.RawMessage, error) {
	return nil, nil
}
func (n *NoOpTracer) Stop(err error) {}
