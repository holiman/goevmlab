### Consensus error

This is a consensus error affecting `nimbus-eth1`, and the `BYTE` operation. 

```
prev:           both: {"pc":15337,"op":26,"gas":"0x22","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x77676767676760000000000000001002e000000000000040000000e000000000","0x10000000000000000"],"depth":1,"refund":0,"opName":"BYTE"}
diff:         geth-0: {"pc":15338,"op":10,"gas":"0x1f","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x0"],"depth":1,"refund":0,"opName":"EXP"}
diff:         nimb-0: {"pc":15338,"op":10,"gas":"0x1f","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x77"],"depth":1,"refund":0,"opName":"EXP"}
Consensus error
Testcase: 2.json.min
- geth-0: ./geth-0-output.jsonl
  - command: /home/martin/workspace/evm --json --noreturndata --nomemory statetest 2.json.min
- nimb-0: ./nimb-0-output.jsonl
  - command: /home/martin/workspace/evmstate --json --noreturndata --nomemory --nostorage 2.json.min

```
Geth trace
```
martin@mediaNUK:~/workspace/goevmlab/cmd/runtest$ tail ./geth-0-output.jsonl
{"pc":15289,"op":144,"gas":"0x39","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0x0","0x7000000000000000","0x8000000000000000000000000000000000000000000000000000000000000000"],"depth":1,"refund":0,"opName":"SWAP1"}
{"pc":15290,"op":4,"gas":"0x36","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0x0","0x8000000000000000000000000000000000000000000000000000000000000000","0x7000000000000000"],"depth":1,"refund":0,"opName":"DIV"}
{"pc":15291,"op":25,"gas":"0x31","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0x0","0x0"],"depth":1,"refund":0,"opName":"NOT"}
{"pc":15292,"op":1,"gas":"0x2e","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"],"depth":1,"refund":0,"opName":"ADD"}
{"pc":15293,"op":104,"gas":"0x2b","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"],"depth":1,"refund":0,"opName":"PUSH9"}
{"pc":15303,"op":127,"gas":"0x28","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x10000000000000000"],"depth":1,"refund":0,"opName":"PUSH32"}
{"pc":15336,"op":144,"gas":"0x25","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x10000000000000000","0x77676767676760000000000000001002e000000000000040000000e000000000"],"depth":1,"refund":0,"opName":"SWAP1"}
{"pc":15337,"op":26,"gas":"0x22","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x77676767676760000000000000001002e000000000000040000000e000000000","0x10000000000000000"],"depth":1,"refund":0,"opName":"BYTE"}
{"pc":15338,"op":10,"gas":"0x1f","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x0"],"depth":1,"refund":0,"opName":"EXP"}
{"stateRoot":"0x347349e4362fc665ac00256223e62932d64a7d03e80e71799742d24115d965aa"}
```
Nim trace

```
martin@mediaNUK:~/workspace/goevmlab/cmd/runtest$ tail ./nimb-0-output.jsonl
{"pc":15289,"op":144,"gas":"0x39","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0x0","0x7000000000000000","0x8000000000000000000000000000000000000000000000000000000000000000"],"depth":1,"refund":0,"opName":"SWAP1"}
{"pc":15290,"op":4,"gas":"0x36","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0x0","0x8000000000000000000000000000000000000000000000000000000000000000","0x7000000000000000"],"depth":1,"refund":0,"opName":"DIV"}
{"pc":15291,"op":25,"gas":"0x31","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0x0","0x0"],"depth":1,"refund":0,"opName":"NOT"}
{"pc":15292,"op":1,"gas":"0x2e","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"],"depth":1,"refund":0,"opName":"ADD"}
{"pc":15293,"op":104,"gas":"0x2b","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"],"depth":1,"refund":0,"opName":"PUSH9"}
{"pc":15303,"op":127,"gas":"0x28","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x10000000000000000"],"depth":1,"refund":0,"opName":"PUSH32"}
{"pc":15336,"op":144,"gas":"0x25","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x10000000000000000","0x77676767676760000000000000001002e000000000000040000000e000000000"],"depth":1,"refund":0,"opName":"SWAP1"}
{"pc":15337,"op":26,"gas":"0x22","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x77676767676760000000000000001002e000000000000040000000e000000000","0x10000000000000000"],"depth":1,"refund":0,"opName":"BYTE"}
{"pc":15338,"op":10,"gas":"0x1f","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x77"],"depth":1,"refund":0,"opName":"EXP"}
{"stateRoot":"0x347349e4362fc665ac00256223e62932d64a7d03e80e71799742d24115d965aa"}

```
Difference seems to be
```
BYTE(0x10000000000000000, 0x77676767676760000000000000001002e000000000000040000000e000000000) 
 -> 0x0 (geth)
 -> 0x77 (evmstate)
```

Another example, similar bug: 
```
prev:           both: {"pc":10034,"op":26,"gas":"0xf38e4a","gasCost":"0x0","memSize":0,"stack":["0x1","0x1f000000000000000000000000000000200000000100000000000000000000","0x80000000000000000000000000000001"],"depth":1,"refund":0,"opName":"BYTE"}
diff:         geth-0: {"pc":10035,"op":144,"gas":"0xf38e47","gasCost":"0x0","memSize":0,"stack":["0x1","0x0"],"depth":1,"refund":0,"opName":"SWAP1"}
diff:         nimb-0: {"pc":10035,"op":144,"gas":"0xf38e47","gasCost":"0x0","memSize":0,"stack":["0x1","0x1f"],"depth":1,"refund":0,"opName":"SWAP1"}
```
```
BYTE(0x80000000000000000000000000000001, 0x1f000000000000000000000000000000200000000100000000000000000000) 
 -> 0x0 (geth)
 -> 0x1f (evmstate)
```


### Fix 

This was fixed in [#1464](https://github.com/status-im/nimbus-eth1/pull/1464)
