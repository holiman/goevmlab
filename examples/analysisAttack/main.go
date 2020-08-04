// Copyright 2019 Martin Holst Swende, Hubert Ritzdorf
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

package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/goevmlab/ops"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/goevmlab/program"
)

func main() {

	if err := program.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func payload() []byte {
	size := 0xc000
	p := make([]byte, size)
	for i := 0; i < size-1; i++ {
		p[i] = 0x60
		p[i+1] = 0x01
	}
	// It needs to actually do a jump to trigger the analysis
	p[0] = byte(ops.PUSH1)
	p[1] = byte(0x03)
	p[2] = byte(ops.JUMP)
	p[3] = byte(ops.JUMPDEST)
	p[4] = byte(ops.STOP)
	return p
}

func runit() error {

	/**

	The scheme:
	- Contract A contains a payload (e.g. filled with JUMPDESTs)
	- Contract B loads 49152 bytes from A into memory (payload)
	- loop:
	- Contract B calls CREATE with payload
	- The CREATE fails
	- (Flip a byte in memory - this step is omitted, since we don't reuse jumpdest analysis for initcode anyway)
	- goto loop
	*/
	aAddress := common.HexToAddress("0xff0a")
	bAddress := common.HexToAddress("0xff0b")

	alloc := make(core.GenesisAlloc)
	alloc[aAddress] = core.GenesisAccount{
		Nonce:   0,
		Code:    payload(),
		Balance: big.NewInt(0xffffffff),
	}

	// Calling contract: Call contract B in a loop
	p := program.NewProgram()
	// Load payload
	p.ExtcodeCopy(aAddress, 0, 0, 0xc000)
	label := p.Jumpdest()
	// size, offset, value, CREATE
	p.Push(0xc000).Push(0).Push(0).Op(ops.CREATE)
	// ret value will be 1 (failure)
	p.Op(ops.POP)
	// Goto loop
	p.Push(label).Op(ops.JUMP)

	alloc[bAddress] = core.GenesisAccount{
		Nonce:   0,
		Code:    p.Bytecode(),
		Balance: big.NewInt(0xffffffff),
	}
	verbose := false
	if verbose {
		outp, _ := json.MarshalIndent(alloc, "", " ")
		fmt.Printf("output \n%v\n", string(outp))

	}
	//----------
	return execute(alloc, bAddress)
}

func execute(alloc core.GenesisAlloc, entryPoint common.Address) error {

	var (
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)
		sender     = common.BytesToAddress([]byte("sender"))
	)
	for addr, acc := range alloc {
		statedb.CreateAccount(addr)
		statedb.SetCode(addr, acc.Code)
		statedb.SetNonce(addr, acc.Nonce)
		if acc.Balance != nil {
			statedb.SetBalance(addr, acc.Balance)
		}
	}
	statedb.CreateAccount(sender)
	startGas := uint64(10000000)
	runtimeConfig := runtime.Config{
		Origin:      sender,
		State:       statedb,
		GasLimit:    startGas,
		Difficulty:  big.NewInt(0x200000),
		Time:        new(big.Int).SetUint64(0),
		Coinbase:    common.Address{},
		BlockNumber: new(big.Int).SetUint64(1),
		ChainConfig: &params.ChainConfig{
			ChainID:             big.NewInt(1),
			HomesteadBlock:      new(big.Int),
			ByzantiumBlock:      new(big.Int),
			ConstantinopleBlock: new(big.Int),
			EIP150Block:         new(big.Int),
			EIP155Block:         new(big.Int),
			EIP158Block:         new(big.Int),
			PetersburgBlock:     new(big.Int),
			IstanbulBlock:       new(big.Int),
		},
	}
	diagnose := true
	if diagnose {
		// Sane defaults for printing
		runtimeConfig.EVMConfig = vm.Config{
			Debug:  true,
			Tracer: &dumbTracer{},
		}
		runtimeConfig.GasLimit = 100000

		// Run with tracing
		runtime.Call(entryPoint, nil, &runtimeConfig)
		// reset
		runtimeConfig.EVMConfig = vm.Config{}
		runtimeConfig.GasLimit = startGas
	}
	// Diagnose it

	var err error
	var leftOverGas uint64
	benchmarkRes := testing.Benchmark(func(b *testing.B) {
		_, leftOverGas, err = runtime.Call(entryPoint, nil, &runtimeConfig)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _, err = runtime.Call(entryPoint, nil, &runtimeConfig)
		}
	})

	/*

	I've done some analysis. A contract can do roughly ~300 CREATE operations on 10M gas,
	(each CREATE costs 32K).

	I made a test where we do an analysis of a contract which is 0xc000 bytes large,
	mainly consisting of PUSH1 01, but with an initial jump and stop to trigger the analysis.

	I ran it with 100M (obs: `100M`, not `10M`) gas, and benchmarked it.

	With only old-style analysis, it blew through 100M gas in `43ms`:
	```
	      27          43398488 ns/op 25509711 B/op     22689 allocs/op
	gas used: 10000000, 230.422774 Mgas/s

		      22          53288897 ns/op 25716607 B/op     22871 allocs/op
	gas used: 10000000, 187.656352 Mgas/s
```

		With _both_ shadow-array and old analysis made, it blew through 100M gas in `65ms` :
```
	      18          67303719 ns/op 44805790 B/op     23490 allocs/op
	gas used: 10000000, 148.580200 Mgas/s

	      16          64121205 ns/op 45101226 B/op     23648 allocs/op
	gas used: 10000000, 155.954651 Mgas/s
```
	*/

	fmt.Printf("%v %v\n", benchmarkRes.String(), benchmarkRes.MemString())
	gasUsed := startGas - leftOverGas
	t := float64(benchmarkRes.NsPerOp())
	fmt.Printf("gas used: %v, %02f Mgas/s\n", gasUsed, float64(1000*gasUsed)/t)

	//gas/ns = kgas/us = mgas/ms = gas/s

	return err
}

type dumbTracer struct {
	counter uint64
}

func (d *dumbTracer) CaptureState(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, rStack *vm.ReturnStack, contract *vm.Contract, depth int, err error) error {
	fmt.Printf("depth %d pc %d op %v gas %d cost %d\n", depth, pc, op, gas, cost)
	if op == vm.CREATE {
		d.counter++
	}
	return nil
}

func (d *dumbTracer) CaptureFault(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, rStack *vm.ReturnStack, contract *vm.Contract, depth int, err error) error {
	fmt.Printf("CaptureFault: %v\n", err)
	return nil
}

func (d *dumbTracer) CaptureStart(from common.Address, to common.Address, call bool, input []byte, gas uint64, value *big.Int) error {
	fmt.Printf("captureStart\n")
	return nil
}

func (d *dumbTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error {
	fmt.Printf("\nCaptureEnd\n")
	fmt.Printf("Counter: %d\n", d.counter)
	return nil
}
