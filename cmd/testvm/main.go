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
	if true {
		fmt.Fprint(os.Stderr, out2)
		fmt.Fprint(os.Stdout, out2)
	} else {
		fmt.Fprint(os.Stderr, out)
		fmt.Fprint(os.Stdout, out)
	}
}
