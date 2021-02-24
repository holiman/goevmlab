package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/goevmlab/program"
	"math/big"
	"os"
	"time"
)

func main() {

	if err := program.RunProgram(runit); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runit() error {

	fmt.Println(`## Current (Istanbul) costs

With Istanbul rules, [EIP-2200](testcases): https://eips.ethereum.org/EIPS/eip-2200#test-cases, the following gas usages 
apply for the various scenarions below: 
`)
	fmt.Println("")

	if err := showTable(false, false); err != nil {
		return err
	}
	fmt.Println(`## Berlin costs

With Berlin, and [EIP-2929](https://eips.ethereum.org/EIPS/eip-2929), the gas costs changes. Note, there is a difference between 'hot' and 'cold' slots. 
The following table is generated based on the slot in question (0) being 'cold'.`)
	fmt.Println("")

	if err := showTable(true, false); err != nil {
		return err
	}
	fmt.Println(`If the slot is already 'warm', this is the corresponding table: `)
	fmt.Println("")

	if err := showTable(true, true); err != nil {
		return err
	}
	fmt.Println(`## Without refunds 

If refunds were to be removed, this would be the comparative table`)

	showTable2()

	fmt.Println(`
In partcular, the following cases become more expensive: 

- '0-1-0' goes from 10K to 20K, 
- '0-1-0-1' goes from 21K to 40K,

For these cases, a better scheme under the no-refund rule would be to use non-zero slots, e.g. 
'1-2-1' and thus wind up with 3K gas. 

**Note**: In reality, there are never a negative gas cost, since the refund is capped at 0.5 * gasUsed. 
However, these tables show the negative values, since in a more real-world scenarion would likely spend the 
extra gas on other operations.notes'

`)
	return nil
}

func showTable(berlin, hot bool) error {
	fmt.Printf("| Code | Used Gas | Refund | Original | 1st | 2nd | 3rd | Effective gas (after refund)\n")
	fmt.Printf("| -- | -- | -- | -- | -- | -- | -- | -- | \n")

	tracer := &dumbTracer{}

	chainConfig := &params.ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      new(big.Int),
		ByzantiumBlock:      new(big.Int),
		ConstantinopleBlock: new(big.Int),
		EIP150Block:         new(big.Int),
		EIP155Block:         new(big.Int),
		EIP158Block:         new(big.Int),
		PetersburgBlock:     new(big.Int),
		IstanbulBlock:       new(big.Int).SetUint64(0),
	}
	if berlin {
		chainConfig.YoloV3Block = big.NewInt(0)
	}

	// The destructor, caller and sender
	execAddr := common.HexToAddress("0x000000000000000000000000000000000000c411")
	sender := common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	slot := 0
	for _, tc := range []struct {
		original byte
		changes  []byte
	}{
		{0, []byte{0, 0}},
		{0, []byte{0, 1}},
		{0, []byte{1, 0}},
		{0, []byte{1, 2}},
		{0, []byte{1, 1}},

		{1, []byte{0, 0}},
		{1, []byte{0, 1}},
		{1, []byte{0, 2}},
		{1, []byte{2, 0}},
		{1, []byte{2, 3}},
		{1, []byte{2, 1}},
		{1, []byte{2, 2}},
		{1, []byte{1, 0}},
		{1, []byte{1, 2}},
		{1, []byte{1, 1}},
		//
		{0, []byte{1, 0, 1}},
		{1, []byte{0, 1, 0}},
	} {

		// Program which does sstores
		caller := program.NewProgram()
		for _, val := range tc.changes {
			caller.Sstore(0, val)
		}
		var (
			ethdb    = rawdb.NewMemoryDatabase()
			db       = state.NewDatabase(ethdb)
			alloc    = make(core.GenesisAlloc)
			slotHash = common.HexToHash(fmt.Sprintf("%#064x", slot))
		)
		alloc[execAddr] = core.GenesisAccount{
			Nonce:   1,
			Code:    caller.Bytecode(),
			Balance: big.NewInt(0x1),
			Storage: map[common.Hash]common.Hash{
				slotHash: common.HexToHash(fmt.Sprintf("%#064x", tc.original)),
			},
		}
		alloc[sender] = core.GenesisAccount{
			Nonce:   0,
			Balance: big.NewInt(1000000000000000000), // 1 eth
		}

		gspec := core.Genesis{
			Config:     chainConfig,
			Alloc:      alloc,
			Number:     0,
			GasUsed:    0,
			ParentHash: common.Hash{},
		}
		b := gspec.MustCommit(ethdb)
		r := b.Root()
		var (
			statedb, _ = state.New(r, db, nil)
			err        error
		)
		//statedb.AddAddressToAccessList()
		runtimeConfig := runtime.Config{
			Origin:      sender,
			State:       statedb,
			GasLimit:    10000000,
			Difficulty:  big.NewInt(0x200000),
			Time:        new(big.Int).SetUint64(0),
			Coinbase:    common.Address{},
			BlockNumber: new(big.Int).SetUint64(1),
			ChainConfig: chainConfig,
			EVMConfig: vm.Config{
				Debug: true,
				//Tracer: vm.NewMarkdownLogger(nil, os.Stdout),
				Tracer: tracer,
			},
		}
		if hot {
			statedb.AddSlotToAccessList(execAddr, slotHash)
		}
		// Run with tracing
		_, _, err = runtime.Call(execAddr, nil, &runtimeConfig)
		if err != nil {
			return err
		}
		fmt.Printf("| `%#x` | %d | %d| %d |",
			caller.Bytecode(),
			tracer.usedGas,
			tracer.refund,
			tc.original)

		for i := 0; i < 3; i++ {
			if i < len(tc.changes) {
				fmt.Printf(" %d | ", tc.changes[i])
			} else {
				fmt.Printf(" | ")
			}
		}
		fmt.Printf(" %d |", effectiveGas(tracer.usedGas, tracer.refund))
		fmt.Printf("\n")
	}
	fmt.Printf("\n")

	return nil
}

func showTable2() error {
	fmt.Printf("| Code | Original | 1st | 2nd | 3rd |  Istanbul | Berlin (cold) | Berlin (hot)| Berlin (hot)+norefund |\n")
	fmt.Printf("| -- | -- | -- | -- | -- |  -- | -- | -- | -- | \n")

	// The destructor, caller and sender
	execAddr := common.HexToAddress("0x000000000000000000000000000000000000c411")
	sender := common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	slot := 0
	for _, tc := range []struct {
		original byte
		changes  []byte
	}{
		{0, []byte{0, 0}},
		{0, []byte{0, 1}},
		{0, []byte{1, 0}},
		{0, []byte{1, 2}},
		{0, []byte{1, 1}},

		{1, []byte{0, 0}},
		{1, []byte{0, 1}},
		{1, []byte{0, 2}},
		{1, []byte{2, 0}},
		{1, []byte{2, 3}},
		{1, []byte{2, 1}},
		{1, []byte{2, 2}},
		{1, []byte{1, 0}},
		{1, []byte{1, 2}},
		{1, []byte{1, 1}},
		//
		{0, []byte{1, 0, 1}},
		{1, []byte{0, 1, 0}},
	} {

		// Program which does sstores
		caller := program.NewProgram()
		for _, val := range tc.changes {
			caller.Sstore(0, val)
		}
		exec := func(berlin, hot bool) *dumbTracer {
			tracer := &dumbTracer{}
			chainConfig := &params.ChainConfig{
				ChainID:             big.NewInt(1),
				HomesteadBlock:      new(big.Int),
				ByzantiumBlock:      new(big.Int),
				ConstantinopleBlock: new(big.Int),
				EIP150Block:         new(big.Int),
				EIP155Block:         new(big.Int),
				EIP158Block:         new(big.Int),
				PetersburgBlock:     new(big.Int),
				IstanbulBlock:       new(big.Int).SetUint64(0),
			}
			if berlin {
				chainConfig.YoloV3Block = big.NewInt(0)
			}
			var (
				ethdb    = rawdb.NewMemoryDatabase()
				db       = state.NewDatabase(ethdb)
				alloc    = make(core.GenesisAlloc)
				slotHash = common.HexToHash(fmt.Sprintf("%#064x", slot))
			)
			alloc[execAddr] = core.GenesisAccount{
				Nonce:   1,
				Code:    caller.Bytecode(),
				Balance: big.NewInt(0x1),
				Storage: map[common.Hash]common.Hash{
					slotHash: common.HexToHash(fmt.Sprintf("%#064x", tc.original)),
				},
			}
			alloc[sender] = core.GenesisAccount{
				Nonce:   0,
				Balance: big.NewInt(1000000000000000000), // 1 eth
			}

			gspec := core.Genesis{
				Config:     chainConfig,
				Alloc:      alloc,
				Number:     0,
				GasUsed:    0,
				ParentHash: common.Hash{},
			}
			b := gspec.MustCommit(ethdb)
			r := b.Root()
			var (
				statedb, _ = state.New(r, db, nil)
				err        error
			)
			//statedb.AddAddressToAccessList()
			runtimeConfig := runtime.Config{
				Origin:      sender,
				State:       statedb,
				GasLimit:    10000000,
				Difficulty:  big.NewInt(0x200000),
				Time:        new(big.Int).SetUint64(0),
				Coinbase:    common.Address{},
				BlockNumber: new(big.Int).SetUint64(1),
				ChainConfig: chainConfig,
				EVMConfig: vm.Config{
					Debug: true,
					//Tracer: vm.NewMarkdownLogger(nil, os.Stdout),
					Tracer: tracer,
				},
			}
			if hot {
				statedb.AddSlotToAccessList(execAddr, slotHash)
			}
			// Run with tracing
			_, _, err = runtime.Call(execAddr, nil, &runtimeConfig)
			if err != nil {
				panic(err)
			}
			return tracer
		}
		fmt.Printf("| `%#x` | %d | ",
			caller.Bytecode(),
			tc.original)
		for i := 0; i < 3; i++ {
			if i < len(tc.changes) {
				fmt.Printf(" %d | ", tc.changes[i])
			} else {
				fmt.Printf(" | ")
			}
		}

		tracer := exec(false, false)
		fmt.Printf(" %d |", effectiveGas(tracer.usedGas, tracer.refund))
		tracer = exec(true, false)
		fmt.Printf(" %d |", effectiveGas(tracer.usedGas, tracer.refund))
		tracer = exec(true, true)
		fmt.Printf(" %d |", effectiveGas(tracer.usedGas, tracer.refund))
		tracer = exec(true, true)
		fmt.Printf(" %d |", effectiveGas(tracer.usedGas, 0))
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
	return nil
}

func effectiveGas(gasUsed, refund uint64) int64 {
	// Apply refund counter, capped to half of the used gas.
	if false {
		refundMax := gasUsed / 2
		if refundMax > refund {
			return int64(gasUsed - refund)
		}
		return int64(gasUsed - refundMax)
	}
	return int64(gasUsed - refund)
}

type dumbTracer struct {
	usedGas uint64
	refund  uint64
	statedb vm.StateDB
}

func (d *dumbTracer) CaptureStart(from common.Address, to common.Address, call bool, input []byte, gas uint64, value *big.Int) error {
	return nil
}

func (d *dumbTracer) CaptureState(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, rstack *vm.ReturnStack, rData []byte, contract *vm.Contract, depth int, err error) error {
	//if op == vm.SSTORE{
	//	fmt.Printf("pc %d op %v gas %d cost %d refund %d\n", pc, op, gas, cost,  env.StateDB.GetRefund())
	//}
	d.statedb = env.StateDB
	return nil
}

func (d *dumbTracer) CaptureFault(env *vm.EVM, pc uint64, op vm.OpCode, gas, cost uint64, memory *vm.Memory, stack *vm.Stack, rstack *vm.ReturnStack, contract *vm.Contract, depth int, err error) error {
	//fmt.Printf("CaptureFault %v\n", err)
	return nil
}

func (d *dumbTracer) CaptureEnd(output []byte, gasUsed uint64, t time.Duration, err error) error {
	//fmt.Printf("Used gas %d\n",gasUsed)
	d.usedGas = gasUsed
	d.refund = d.statedb.GetRefund()
	return nil
}
