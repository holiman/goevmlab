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
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

// StateDBWithAlloc creates a statedb and populates with the given alloc.
func StateDBWithAlloc(alloc types.GenesisAlloc) *state.StateDB {
	statedb := NewEmptyStateDB()
	for addr, acc := range alloc {
		statedb.CreateAccount(addr)
		statedb.SetCode(addr, acc.Code)
		statedb.SetNonce(addr, acc.Nonce, 0)
		if acc.Balance != nil {
			statedb.SetBalance(addr, uint256.MustFromBig(acc.Balance), tracing.BalanceChangeUnspecified)
		}
	}
	return statedb
}

// NewEmptyStateDB creates an empty statedb, or panics on error
func NewEmptyStateDB() *state.StateDB {
	statedb, err := state.New(types.EmptyRootHash, NewMemStateDB())
	if err != nil {
		panic(err)
	}
	return statedb
}

func NewMemStateDB() state.Database {
	return state.NewDatabase(triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil), nil)
}
