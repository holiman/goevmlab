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
	"math/rand"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/ops"
)

type codeMutator struct {
	current  []byte
	lastGood []byte
}

// proceed tells the mutator to continue one mutation
// returns 'true' is the mutator is exhausted
func (m *codeMutator) proceed() bool {
	m.lastGood = m.current
	// Now mutate current
	var (
		next    []byte
		max     = ops.InstructionCount(m.lastGood)
		removed = 0
	)
	for removed == 0 {
		it := ops.NewInstructionIterator(m.lastGood)
		cutPoint := rand.Intn(max)
		next = make([]byte, 0)
		for it.Next() {
			if removed == 0 && it.PC() > uint64(cutPoint) {
				// Remove until the stack balances out
				delta := 0
				for {
					removed += 1
					delta += it.Op().Stackdelta()
					if delta == 0 {
						break
					}
					// Skip next one too
					if !it.Next() {
						break
					}
				}
			}
			next = append(next, byte(it.Op()))
			if arg := it.Arg(); arg != nil {
				next = append(next, arg...)
			}
		}
	}
	log.Info("Mutating code", "length", len(next), "previous", len(m.lastGood))
	m.current = next
	return len(next) == len(m.lastGood)
}

// undo tells the mutator to revert the last change
func (m *codeMutator) undo() {
	m.current = m.lastGood
}

type naiveCodeMutator struct {
	current  []byte
	lastGood []byte
}

// proceed tells the mutator to continue one mutation
// returns 'true' is the mutator is exhausted
func (m *naiveCodeMutator) proceed() bool {
	m.lastGood = m.current
	// Now mutate current
	var next []byte = make([]byte, len(m.current)*2/3) // Remove one third
	copy(next, m.lastGood)
	log.Info("Mutating code #1", "length", len(next), "previous", len(m.lastGood))
	m.current = next
	return len(next) == len(m.lastGood)
}

// undo tells the mutator to revert the last change
func (m *naiveCodeMutator) undo() {
	m.current = m.lastGood
}
