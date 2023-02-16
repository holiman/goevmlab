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

package fuzzing

import (
	crand "crypto/rand"
	"encoding/binary"
	"math"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

type memFunc func() (offset, size interface{})
type valFunc func() interface{}

// randHex produces some random hex data
func randHex(maxSize int) string {
	size := rand.Intn(maxSize)
	b := make([]byte, size)
	_, _ = crand.Read(b)
	return hexutil.Encode(b)
}

// randInt returns a valFunc which spits out bigints,
// - Chance of zero, expressed as N out of 255.
// - Chance of small value (< 255 ), expressed as N out of 255.
func randInt(chanceOfZero, chanceOfSmall byte) valFunc {
	return func() interface{} {
		b := make([]byte, 4)
		_, _ = crand.Read(b)
		// Zero or not?
		if b[0] < chanceOfZero {
			return big.NewInt(0)
		}
		if b[1] < chanceOfSmall {
			return (new(big.Int)).SetBytes(b[2:3])
		}
		val := make([]byte, 32)
		_, _ = crand.Read(val)
		return (new(big.Int)).SetBytes(val)
	}
}

// addressRandomizer randomizes from the given addresses
func addressRandomizer(addrs []common.Address) valFunc {
	return func() interface{} {
		return addrs[rand.Intn(len(addrs))]
	}
}

func ValueRandomizer() valFunc {
	// every 16th is zero
	// Most are small, but every 16th is unbounded
	return randInt(0x0f, 0xef)
}

func MemRandomizer() memFunc {
	// half are zero
	// most are small
	v := randInt(0x70, 0xef)
	memFn := func() (offset, size interface{}) {
		return v(), v()
	}
	return memFn
}
func GasRandomizer() valFunc {
	// Very few are zero,
	// 1/16th are small,
	// most are huge
	return randInt(0x02, 0x0f)

}

// staticcall disabled due to parity implementation of cheap staticcall-to-precompile
var callTypes = []ops.OpCode{ops.CALL, ops.CALLCODE, ops.DELEGATECALL} //, ops.STATICCALL}

func randCallType() ops.OpCode {
	return callTypes[rand.Intn(len(callTypes))]
}

func RandCall(gas, addr, val valFunc, memIn, memOut memFunc) []byte {
	p := program.NewProgram()
	if memOut != nil {
		memOutOffset, memOutSize := memOut()
		p.Push(memOutSize)   //mem out size
		p.Push(memOutOffset) // mem out start
	} else {
		p.Push(0)
		p.Push(0)
	}
	if memIn != nil {
		memInOffset, memInSize := memIn()
		p.Push(memInSize)   //mem in size
		p.Push(memInOffset) // mem in start
	} else {
		p.Push(0)
		p.Push(0)
	}
	op := randCallType()
	if op == ops.CALL || op == ops.CALLCODE {
		if val != nil {
			p.Push(val()) //value
		} else {
			p.Push(0)
		}
	}
	p.Push(addr())
	if gas != nil {
		p.Push(gas())
	} else {
		p.Op(ops.GAS)
	}
	p.Op(op)
	return p.Bytecode()
}

func randomBlakeArgs() []byte {
	//params are
	var rounds uint32
	data := make([]byte, 214)
	_, _ = crand.Read(data)
	// Now, modify the rounds, and the 'f'
	// rounds should be below 1024 for the most part
	rounds = uint32(math.Abs(1024 * rand.ExpFloat64()))
	binary.BigEndian.PutUint32(data, rounds)
	x := data[213]
	switch {
	case x == 0:
		// Leave f as is in 1/256th of the tests
	case x < 0x80:
		// set to zer0
		data[212] = 0

	default:
		data[212] = 1
	}
	return data[0:213]
}

func RandCallBlake() []byte {
	// fill the memory
	p := program.NewProgram()
	data := randomBlakeArgs()
	p.Mstore(data, 0)
	memInFn := func() (offset, size interface{}) {
		// todo:make mem generator which mostly outputs 0:213
		offset, size = 0, 213
		return
	}
	// blake outputs 64 bytes
	memOutFn := func() (offset, size interface{}) {
		offset, size = 0, 64
		return
	}
	addrGen := func() interface{} {
		return 9
	}
	p2 := RandCall(GasRandomizer(), addrGen, ValueRandomizer(), memInFn, memOutFn)
	p.AddAll(p2)
	// pop the ret value
	p.Op(ops.POP)
	// Store the output in some slot, to make sure the stateroot changes
	p.MemToStorage(0, 64, 0)
	return p.Bytecode()
}
