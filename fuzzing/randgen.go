package fuzzing

import (
	"encoding/binary"
	"math"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
)

type memFunc func() (offset, size interface{})
type valFunc func() interface{}

// randInt returns a valFunc which spits out bigints,
// - Chance of zero, expressed as N out of 255.
// - Chance of small value (< 255 ), expressed as N out of 255.
func randInt(chanceOfZero, chanceOfSmall byte) valFunc {
	return func() interface{} {
		b := make([]byte, 4)
		rand.Read(b)
		// Zero or not?
		if b[0] < chanceOfZero {
			return big.NewInt(0)
		}
		if b[1] < chanceOfSmall {
			return (new(big.Int)).SetBytes(b[2:3])
		}
		val := make([]byte, 32)
		rand.Read(val)
		return (new(big.Int)).SetBytes(val)
	}
}

// addressRandomizer randomizes from the given addresses
func addressRandomizer(addrs []common.Address) valFunc {
	return func() interface{} {
		return addrs[rand.Intn(len(addrs))]
	}
}

func valueRandomizer() valFunc {
	// every 16th is zero
	// Most are small, but every 16th is unbounded
	return randInt(0x0f, 0xef)
}

func memRandomizer() memFunc {
	// half are zero
	// most are small
	v := randInt(0x70, 0xef)
	memFn := func() (offset, size interface{}) {
		return v(), v()
	}
	return memFn
}
func gasRandomizer() valFunc {
	// Very few are zero,
	// 1/16th are small,
	// most are huge
	return randInt(0x02, 0x0f)

}

var callTypes = []ops.OpCode{ops.CALL, ops.CALLCODE, ops.DELEGATECALL, ops.STATICCALL}

func RandCall(gas, addr, val valFunc, memIn, memOut memFunc) []byte {
	p := program.NewProgram()
	memOutOffset, memOutSize := memOut()
	p.Push(memOutSize)   //mem out size
	p.Push(memOutOffset) // mem out start
	memInOffset, memInSize := memIn()
	p.Push(memInSize)   //mem in size
	p.Push(memInOffset) // mem in start
	op := callTypes[rand.Intn(len(callTypes))]
	if op == ops.CALL || op == ops.CALLCODE {
		p.Push(val()) //value
	}
	p.Push(addr())
	p.Push(gas())
	p.Op(op)
	return p.Bytecode()
}

func randomBlakeArgs() []byte {
	//params are
	var rounds uint32
	data := make([]byte, 214)
	rand.Read(data)
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
	p2 := RandCall(gasRandomizer(), addrGen, valueRandomizer(), memInFn, memOutFn)
	p.AddAll(p2)
	// pop the ret value
	p.Op(ops.POP)
	// Store the output in some slot, to make sure the stateroot changes
	p.MemToStorage(0, 64, 0)
	return p.Bytecode()
}
