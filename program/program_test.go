// Copyright 2019 Martin Holst Swende
// This file is part of the go-evmlab library.
//
// The go-evmlab library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package program

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"testing"
)

func TestPush(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		// native ints
		{0, "6000"},
		{uint64(1), "6001"},
		{0xfff, "610fff"},
		// bigints
		{big.NewInt(0), "6000"},
		{big.NewInt(1), "6001"},
		{big.NewInt(0xfff), "610fff"},
		// Addresses
		{common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef"), "73deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"},
		{&common.Address{}, "6000"},
	}
	for i, tc := range tests {
		p := NewProgram()
		p.Push(tc.input)
		got := fmt.Sprintf("%02x", p.Bytecode())
		if got != tc.expected {
			t.Errorf("test %d: got %v expected %v", i, got, tc.expected)
		}
	}
}
func TestCall(t *testing.T) {
	{ // Nil gas
		p := NewProgram()
		p.Call(nil, common.HexToAddress("0x1337"), big.NewInt(1), 1, 2, 3, 4)
		got := fmt.Sprintf("%02x", p.Bytecode())
		exp := "600460036002600160016113375af1"
		if got != exp {
			t.Errorf("got %v expected %v", got, exp)
		}
	}
	{ // Non nil gas
		p := NewProgram()
		p.Call(big.NewInt(0xffff), common.HexToAddress("0x1337"), big.NewInt(1), 1, 2, 3, 4)
		got := fmt.Sprintf("%02x", p.Bytecode())
		exp := "6004600360026001600161133761fffff1"
		if got != exp {
			t.Errorf("got %v expected %v", got, exp)
		}
	}
}

func TestMstore(t *testing.T) {

	{
		p := NewProgram()
		p.Mstore(common.FromHex("0xaabb"), 0)
		got := fmt.Sprintf("%02x", p.Bytecode())
		exp := "60aa60005360bb600153"
		if got != exp {
			t.Errorf("got %v expected %v", got, exp)
		}
	}

	{
		p := NewProgram()
		p.Mstore(common.FromHex("0xaabb"), 3)
		got := fmt.Sprintf("%02x", p.Bytecode())
		exp := "60aa60035360bb600453"
		if got != exp {
			t.Errorf("got %v expected %v", got, exp)
		}
	}

	{
		// 34 bytes
		data := common.FromHex("0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF" +
			"FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF" +
			"FFFF")

		p := NewProgram()
		p.Mstore(data, 0)
		got := fmt.Sprintf("%02x", p.Bytecode())
		exp := "7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff60005260ff60205360ff602153"
		if got != exp {
			t.Errorf("got %v expected %v", got, exp)
		}
	}

}

func TestMemToStorage(t *testing.T) {
	{
		p := NewProgram()
		p.MemToStorage(0, 33, 1)
		got := fmt.Sprintf("%02x", p.Bytecode())
		exp := "600051600155602051600255"
		if got != exp {
			t.Errorf("got %v expected %v", got, exp)
		}
	}
}
