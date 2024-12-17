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

package program

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
)

func TestCreateAndCall(t *testing.T) {

	// A constructor that stores a slot
	ctor := program.New()
	ctor.Sstore(0, big.NewInt(5))

	// A runtime bytecode which reads the slot and returns
	deployed := program.New()
	deployed.Push(0)
	deployed.Op(vm.SLOAD) // [value] in stack
	deployed.Push(0)      // [value, 0]
	deployed.Op(vm.MSTORE)
	deployed.Return(0, 32)

	// Pack them
	ctor.ReturnData(deployed.Bytes())
	// Verify constructor + runtime code
	{
		exp := "6005600055606060005360006001536054600253606060035360006004536052600553606060065360206007536060600853600060095360f3600a53600b6000f3"
		if got := ctor.Hex(); got != exp {
			t.Fatalf("1: got %v expected %v", got, exp)
		}
	}

	{ // Verify CREATE + CALL
		p := program.New()
		CreateAndCall(p, ctor.Bytes(), false, vm.CALL)
		exp := "7f60056000556060600053600060015360546002536060600353600060045360526000527f600553606060065360206007536060600853600060095360f3600a53600b600060205260f3604053604160006000f060006000600060006000855af15050"
		if got := p.Hex(); got != exp {
			t.Fatalf("2: got %v expected %v", got, exp)
		}
	}

	{ // Verify CREATE + DELEGATECALL
		p := program.New()
		CreateAndCall(p, ctor.Bytes(), false, vm.DELEGATECALL)
		exp := "7f60056000556060600053600060015360546002536060600353600060045360526000527f600553606060065360206007536060600853600060095360f3600a53600b600060205260f3604053604160006000f06000600060006000845af45050"
		if got := p.Hex(); got != exp {
			t.Fatalf("3: got %v expected %v", got, exp)
		}
	}

	{ // Verify CREATE2 + STATICCALL
		p := program.New()
		CreateAndCall(p, ctor.Bytes(), true, vm.STATICCALL)
		exp := "7f60056000556060600053600060015360546002536060600353600060045360526000527f600553606060065360206007536060600853600060095360f3600a53600b600060205260f36040536000604160006000f56000600060006000845afa5050"
		if got := p.Hex(); got != exp {
			t.Fatalf("2: got %v expected %v", got, exp)
		}
	}
}
