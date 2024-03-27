package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	time.Sleep(500 * time.Millisecond)
	out := `{"pc":0,"op":96,"gas":"0x4757","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":2,"op":96,"gas":"0x4754","gasCost":"0x3","memSize":0,"stack":["0x0"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":13,"op":96,"gas":"0xf2","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":15,"op":84,"gas":"0xef","gasCost":"0x834","memSize":0,"stack":["0x3"],"depth":1,"refund":0,"opName":"SLOAD","error":"out of gas"}
{"output":"","gasUsed":"0x4757","error":"out of gas"}
{"stateRoot": "0x75dc56643cc707a2e6c9a4cf7e28061e9598bd02ecac22c406365c058088d59b"}`

	out2 := `{"pc":0,"op":96,"gas":"0x4757","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":2,"op":96,"gas":"0x4754","gasCost":"0x3","memSize":0,"stack":["0x0"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":13,"op":96,"gas":"0xf2","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":15,"op":84,"gas":"0xef","gasCost":"0x834","memSize":0,"stack":["0x1"],"depth":1,"refund":0,"opName":"SLOAD","error":"out of gas"}
{"output":"","gasUsed":"0x4757","error":"out of gas"}
{"stateRoot": "0xaaaaaaaaaaabbbbbbbbbbbbbbbcccccccccccccccccccccccccccccccccccccc"}`
	out3 := `{"pc":0,"op":96,"gas":"0x4757","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":2,"op":96,"gas":"0x4754","gasCost":"0x3","memSize":0,"stack":["0x0"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":13,"op":96,"gas":"0xf2","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x40 pc=0xd05322]

goroutine 1 [running]:
github.com/ethereum/go-ethereum/eth/tracers/logger.(*jsonLogger).OnOpcode(0xc0000128d0, 0x0, 0x7f, 0x79bff8, 0x3, {0x1634008, 0xc0000138d8}, {0x0, 0x0, 0x0}, ...)
        /home/user/go/src/github.com/ethereum/go-ethereum/eth/tracers/logger/logger_json.go:67 +0xa2
github.com/ethereum/go-ethereum/core/vm.(*EVMInterpreter).Run(0xc000040780, 0xc000000540, {0x20d8400, 0x0, 0x0}, 0x0)
        /home/user/go/src/github.com/ethereum/go-ethereum/core/vm/interpreter.go:280 +0xb95
github.com/ethereum/go-ethereum/core/vm.(*EVM).Call(0xc0002e90e0, {0x16253c0, 0xc000681860}, {0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, ...}, ...)
        /home/user/go/src/github.com/ethereum/go-ethereum/core/vm/evm.go:223 +0x8a5
github.com/ethereum/go-ethereum/core.(*StateTransition).TransitionDb(0xc00052ccf8)
        /home/user/go/src/github.com/ethereum/go-ethereum/core/state_transition.go:436 +0x6d6
github.com/ethereum/go-ethereum/core.ApplyMessage(0x0?, 0x14ca298?, 0x14caN318?)
        /home/user/go/src/github.com/ethereum/go-ethereum/core/state_transition.go:184 +0x57
github.com/ethereum/go-ethereum/tests.(*StateTest).RunNoVerify(0xc00052d4c8, {{0xc000133fa0?, 0x0?}, 0x0?}, {0xc00032e090, 0x0, 0x0, {0x0, 0x0, 0x0}}, ...)
        /home/user/go/src/github.com/ethereum/go-ethereum/tests/state_test_util.go:302 +0x9f0
github.com/ethereum/go-ethereum/tests.(*StateTest).Run(0xc00052d4c8, {{0xc000133fa0?, 0x41451b?}, 0x10?}, {0xc00032e090, 0x0, 0x0, {0x0, 0x0, 0x0}}, ...)
        /home/user/go/src/github.com/ethereum/go-ethereum/tests/state_test_util.go:199 +0xf7
main.runStateTest({0x7ffd29a4afdd?, 0x134ac6c?}, {0xc00032e090, 0x0, 0x0, {0x0, 0x0, 0x0}}, 0x0)
        /home/user/go/src/github.com/ethereum/go-ethereum/cmd/evm/staterunner.go:103 +0x537
main.stateTestCmd(0xc0002d17c0)
        /home/user/go/src/github.com/ethereum/go-ethereum/cmd/evm/staterunner.go:70 +0x698
github.com/ethereum/go-ethereum/internal/flags.MigrateGlobalFlags.func2.1(0xc0002d17c0)
        /home/user/go/src/github.com/ethereum/go-ethereum/internal/flags/helpers.go:100 +0x34
github.com/urfave/cli/v2.(*Command).Run(0x1ec98a0, 0xc0002d17c0, {0xc00028d8c0, 0x2, 0x2})
        /home/user/go/pkg/mod/github.com/urfave/cli/v2@v2.25.7/command.go:274 +0x93f

{"pc":15,"op":84,"gas":"0xef","gasCost":"0x834","memSize":0,"stack":["0x1"],"depth":1,"refund":0,"opName":"SLOAD","error":"out of gas"}
{"output":"","gasUsed":"0x4757","error":"out of gas"}
{"stateRoot": "0xaaaaaaaaaaabbbbbbbbbbbbbbbcccccccccccccccccccccccccccccccccccccc"}`

	if true {
		fmt.Fprint(os.Stderr, out3)
		fmt.Fprint(os.Stdout, out3)
	} else if false {
		fmt.Fprint(os.Stderr, out2)
		fmt.Fprint(os.Stdout, out2)
	} else if false {
		fmt.Fprint(os.Stderr, out)
		fmt.Fprint(os.Stdout, out)
	}
}
