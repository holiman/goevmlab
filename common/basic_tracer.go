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

	"github.com/ethereum/go-ethereum/core/tracing"
)

type BasicTracer struct{}

func (n *BasicTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnFault: n.CaptureFault,
	}
}

func (n *BasicTracer) CaptureFault(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
	fmt.Printf("CaptureFault %v\n", err)
}

func (n *BasicTracer) GetResult() (json.RawMessage, error) {
	return nil, nil
}
func (n *BasicTracer) Stop(err error) {}
