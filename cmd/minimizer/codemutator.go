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

type codeMutator interface {
	reset([]byte)
	proceed() bool
	code() []byte
	undo()
}

type balancedCodeMutator struct {
	current  []byte
	lastGood []byte
}

func newBalancedCodeMutator() codeMutator {
	return &balancedCodeMutator{}
}

func (m *balancedCodeMutator) reset(code []byte) {
	m.current = code
	m.lastGood = code
}

func (m *balancedCodeMutator) code() []byte {
	return m.current
}

// proceed tells the mutator to continue one mutation
// returns 'true' is the mutator is exhausted
func (m *balancedCodeMutator) proceed() bool {
	m.lastGood = m.current
	// Now mutate current
	var (
		next     []byte
		max      = ops.InstructionCount(m.lastGood)
		removed  = 0
		cutPoint = int(0)
	)
	for removed == 0 {
		it := ops.NewInstructionIterator(m.lastGood)
		cutPoint = rand.Intn(max)
		next = make([]byte, 0)
		for it.Next() {
			// skip ahead until we reach cutpoint
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
	log.Info("Mutating code", "length", len(next), "cutpoint", cutPoint, "previous", len(m.lastGood))
	m.current = next
	return len(next) == len(m.lastGood)
}

// undo tells the mutator to revert the last change
func (m *balancedCodeMutator) undo() {
	m.current = m.lastGood
}

type naiveCodeMutator struct {
	current  []byte
	lastGood []byte
	stepSize float64
}

func newCodeShortener() codeMutator {
	return &naiveCodeMutator{}
}

func (m *naiveCodeMutator) reset(code []byte) {
	// Start by removing 50% of the code
	m.current = code
	m.lastGood = code
	m.stepSize = 0.50
}

func (m *naiveCodeMutator) code() []byte {
	return m.current
}

// proceed tells the mutator to continue one mutation
// returns 'true' is the mutator is exhausted
func (m *naiveCodeMutator) proceed() bool {
	m.lastGood = m.current
	// Now mutate current
	newSize := len(m.current) - int(float64(len(m.current))*m.stepSize)
	next := make([]byte, newSize) // Remove a chunk
	copy(next, m.lastGood)
	log.Info("Mutating code #1", "length", len(next), "previous", len(m.lastGood))
	m.current = next
	// We're done when we're no longer shortening the code
	return len(next) == len(m.lastGood)
}

// undo tells the mutator to revert the last change
func (m *naiveCodeMutator) undo() {
	m.stepSize = m.stepSize / 2
	m.current = m.lastGood
}

type codeRandomMutator struct {
	current  []byte
	lastGood []byte
}

func newCodeRandomMutator() codeMutator {
	return &codeRandomMutator{}
}

func (m *codeRandomMutator) reset(code []byte) {
	m.current = code
	m.lastGood = code
}

func (m *codeRandomMutator) code() []byte {
	return m.current
}

// proceed tells the mutator to continue one mutation
// returns 'true' is the mutator is exhausted
func (m *codeRandomMutator) proceed() bool {
	m.lastGood = m.current
	// Now mutate current
	var (
		next     []byte
		max      = ops.InstructionCount(m.lastGood)
		modified = 0
	)
	// Replace 1 of all ops with STOP, or PUSH0 or POP
	it := ops.NewInstructionIterator(m.lastGood)
	cutPoint := rand.Intn(max)
	next = make([]byte, 0)
	for it.Next() {
		if modified == 0 && it.PC() >= uint64(cutPoint) {
			if it.Op() == ops.PUSH1 || it.Op() == ops.POP {
				// just drop it this time
				log.Info("Dropped op", "prev", it.Op().String(), "index", cutPoint)
				modified++
				continue
			}
			delta := it.Op().Stackdelta()
			for i := 0; i < delta; i++ {
				log.Info("Swapped op", "prev", it.Op().String(), "to", "PUSH1 00", "index", cutPoint)
				next = append(next, byte(ops.PUSH1), byte(0x00))
			}
			for i := 0; i > delta; i-- {
				log.Info("Swapped op", "prev", it.Op().String(), "to", "POP", "index", cutPoint)
				next = append(next, byte(ops.POP))
			}
			// else just drop it..
			if delta == 0 {
				log.Info("Dropped op", "prev", it.Op().String(), "index", cutPoint)
			}
			modified++
			continue
		}
		next = append(next, byte(it.Op()))
		if arg := it.Arg(); arg != nil {
			next = append(next, arg...)
		}
	}
	log.Info("Mutating code", "length", len(next), "previous", len(m.lastGood))
	m.current = next
	return modified == 0
}

// undo tells the mutator to revert the last change
func (m *codeRandomMutator) undo() {
	m.current = m.lastGood
}
