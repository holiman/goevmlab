// Copyright 2023 Martin Holst Swende
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

package main

import (
	"bytes"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/ops"
)

type codeMutator interface {
	reset([]byte)
	proceed() bool
	code() []byte
	undo()
}

type replacingCodeMutator struct {
	op       ops.OpCode
	index    int
	current  []byte
	lastGood []byte
}

func newReplacingCodeMutator() codeMutator {
	return &replacingCodeMutator{}
}

func (m *replacingCodeMutator) reset(code []byte) {
	m.current = code
	m.lastGood = code
	m.index = 0
	m.op = ops.STOP
}

func (m *replacingCodeMutator) code() []byte {
	return m.current
}

// proceed tells the mutator to continue one mutation
// returns 'true' is the mutator is exhausted
func (m *replacingCodeMutator) proceed() bool {
	if m.index >= len(m.current) {
		m.index = 0
		m.op = ops.OpCode(byte(m.op) + 1)
		if m.op == ops.STOP {
			return true
		}
	}
	m.lastGood = m.current

	next := bytes.Clone(m.current)
	log.Info("Modified", "index", m.index, "replaced", next[m.index], "with", m.op.String())
	next[m.index] = byte(m.op)
	m.index++
	m.current = next
	return false
	//	return m.index >= len(m.current)
}

// undo tells the mutator to revert the last change
func (m *replacingCodeMutator) undo() {
	m.current = m.lastGood
}
