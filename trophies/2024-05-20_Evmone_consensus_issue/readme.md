```
Consensus error
Testcase: /fuzztmp/01085683-mixed-0.json
- gethbatch-0: /fuzztmp/gethbatch-0-output.jsonl
  - command: /gethvm --json --noreturndata --nomemory statetest
- eelsbatch-0: /fuzztmp/eelsbatch-0-output.jsonl
  - command: /ethereum-spec-evm statetest --json --noreturndata --nomemory
- besubatch-0: /fuzztmp/besubatch-0-output.jsonl
  - command: /evmtool/bin/evm --nomemory --notime --json state-test
- nimbus-0: /fuzztmp/nimbus-0-output.jsonl
  - command: /nimbvm --json --noreturndata --nomemory --nostorage /fuzztmp/01085683-mixed-0.json
- evmone-0: /fuzztmp/evmone-0-output.jsonl
  - command: /evmone --trace /fuzztmp/01085683-mixed-0.json

To view the difference with tracediff:
        tracediff /fuzztmp/gethbatch-0-output.jsonl /fuzztmp/eelsbatch-0-output.jsonl
-------
prev:           both: {"depth":1,"pc":17,"gas":7978563,"op":245,"opName":"CREATE2","stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1"]}
diff:    gethbatch-0: {"depth":2,"pc":0,"gas":7822366,"op":239,"opName":"opcode 0xef not defined","stack":[]}
diff:       evmone-0: {"depth":1,"pc":18,"gas":7946530,"op":22,"opName":"AND","stack":["0x74","0x50","0x0"]}
```

Consensus-split is `[evmone] vs [geth, besu, eels, nimbus]`

Minimized testcase, which is filled with `geth` stateroot, and produces a differing stateroot on `evmone`.
```json
{
  "01085683-mixed-0": {
    "env": {
      "currentCoinbase": "b94f5374fce5edbc8e2a8697c15331677e6ebf0b",
      "currentDifficulty": "0x200000",
      "currentRandom": "0x0000000000000000000000000000000000000000000000000000000000200000",
      "currentGasLimit": "0x26e1f476fe1e22",
      "currentNumber": "0x1",
      "currentTimestamp": "0x3e8",
      "previousHash": "0x044852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d",
      "currentBaseFee": "0x10"
    },
    "pre": {
      "0x00000000000000000000000000000000000000f1": {
        "code": "0x607460506026605d601f60c160ef803552f50052797b749a5746",
        "storage": {},
        "balance": "0x0",
        "nonce": "0x0"
      },
      "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b": {
        "code": "0x",
        "storage": {},
        "balance": "0xffffffffff",
        "nonce": "0x0"
      }
    },
    "transaction": {
      "gasPrice": "0x10",
      "nonce": "0x0",
      "to": "0x00000000000000000000000000000000000000f1",
      "data": [
        "0x092222b40402f77d2de3ade69e6adcc15ff3a49b00e2379edfbf"
      ],
      "gasLimit": [
        "0xd0de"
      ],
      "value": [
        "0x455ffb"
      ],
      "sender": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b",
      "secretKey": "0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"
    },
    "out": "0x",
    "post": {
      "Cancun": [
        {
          "hash": "0x8fe21ebf36fee628e930be403e2f33a2475a670daa48276d52cda143393253a4",
          "logs": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
          "indexes": {
            "data": 0,
            "gas": 0,
            "value": 0
          }
        }
      ]
    }
  }
}
```


Exection on `geth`
```
root@5ef08fb2382f:/# /gethvm --json statetest  /fuzztmp/01085683-mixed-0.mod.json.2.min
{"pc":0,"op":96,"gas":"0x7d42","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":2,"op":96,"gas":"0x7d3f","gasCost":"0x3","memSize":0,"stack":["0x74"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":4,"op":96,"gas":"0x7d3c","gasCost":"0x3","memSize":0,"stack":["0x74","0x50"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":6,"op":96,"gas":"0x7d39","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":8,"op":96,"gas":"0x7d36","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":10,"op":96,"gas":"0x7d33","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":12,"op":96,"gas":"0x7d30","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":14,"op":128,"gas":"0x7d2d","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1","0xef"],"depth":1,"refund":0,"opName":"DUP1"}
{"pc":15,"op":53,"gas":"0x7d2a","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1","0xef","0xef"],"depth":1,"refund":0,"opName":"CALLDATALOAD"}
{"pc":16,"op":82,"gas":"0x7d27","gasCost":"0x6","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1","0xef","0x0"],"depth":1,"refund":0,"opName":"MSTORE"}
{"pc":17,"op":245,"gas":"0x7d21","gasCost":"0x7d21","memSize":32,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1"],"depth":1,"refund":0,"opName":"CREATE2"}
{"pc":0,"op":239,"gas":"0x0","gasCost":"0x0","memSize":0,"stack":[],"depth":2,"refund":0,"opName":"opcode 0xef not defined"}
{"pc":0,"op":239,"gas":"0x0","gasCost":"0x0","memSize":0,"stack":[],"depth":2,"refund":0,"opName":"opcode 0xef not defined","error":"invalid opcode: opcode 0xef not defined"}
{"output":"","gasUsed":"0x0","error":"invalid opcode: opcode 0xef not defined"}
{"pc":18,"op":0,"gas":"0x0","gasCost":"0x0","memSize":128,"stack":["0x74","0x50","0x0"],"depth":1,"refund":0,"opName":"STOP"}
{"output":"","gasUsed":"0x7d42"}
{"stateRoot": "0x8fe21ebf36fee628e930be403e2f33a2475a670daa48276d52cda143393253a4"}
[
  {
    "name": "01085683-mixed-0",
    "pass": true,
    "stateRoot": "0x8fe21ebf36fee628e930be403e2f33a2475a670daa48276d52cda143393253a4",
    "fork": "Cancun"
  }
]
```
Execution on `evmone`:
```
root@5ef08fb2382f:/# /evmone --trace /fuzztmp/01085683-mixed-0.mod.json.2.min
Note: Google Test filter = -stCreateTest.CreateOOGafterMaxCodesize:stQuadraticComplexityTest.Call50000_sha256:stTimeConsuming.static_Call50000_sha256:stTimeConsuming.CALLBlake2f_MaxRounds:VMTests/vmPerformance.*:
[==========] Running 1 test from 1 test suite.
[----------] Global test environment set-up.
[----------] 1 test from /fuzztmp
[ RUN      ] /fuzztmp.01085683-mixed-0.mod.json.2
/evmone/test/statetest/statetest_runner.cpp:66: Failure
Expected equality of these values:
  state_root
    Which is: 0xa282415f5c6c41917f42abcb6dd340dbcd1a1f439362bff7ddf66a8c33484487
  expected.state_hash
    Which is: 0x8fe21ebf36fee628e930be403e2f33a2475a670daa48276d52cda143393253a4
Google Test trace:
/evmone/test/statetest/statetest_runner.cpp:20: Cancun/0
/evmone/test/statetest/statetest_runner.cpp:14: 01085683-mixed-0

[  FAILED  ] /fuzztmp.01085683-mixed-0.mod.json.2 (0 ms)
[----------] 1 test from /fuzztmp (0 ms total)

[----------] Global test environment tear-down
[==========] 1 test from 1 test suite ran. (0 ms total)
[  PASSED  ] 0 tests.
[  FAILED  ] 1 test, listed below:
[  FAILED  ] /fuzztmp.01085683-mixed-0.mod.json.2

 1 FAILED TEST
{"pc":0,"op":96,"gas":"0x7d42","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":2,"op":96,"gas":"0x7d3f","gasCost":"0x3","memSize":0,"stack":["0x74"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":4,"op":96,"gas":"0x7d3c","gasCost":"0x3","memSize":0,"stack":["0x74","0x50"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":6,"op":96,"gas":"0x7d39","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":8,"op":96,"gas":"0x7d36","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":10,"op":96,"gas":"0x7d33","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":12,"op":96,"gas":"0x7d30","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":14,"op":128,"gas":"0x7d2d","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1","0xef"],"depth":1,"refund":0,"opName":"DUP1"}
{"pc":15,"op":53,"gas":"0x7d2a","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1","0xef","0xef"],"depth":1,"refund":0,"opName":"CALLDATALOAD"}
{"pc":16,"op":82,"gas":"0x7d27","gasCost":"0x3","memSize":0,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1","0xef","0x0"],"depth":1,"refund":0,"opName":"MSTORE"}
{"pc":17,"op":245,"gas":"0x7d21","gasCost":"0x7d00","memSize":32,"stack":["0x74","0x50","0x26","0x5d","0x1f","0xc1"],"depth":1,"refund":0,"opName":"CREATE2"}
{"pc":18,"op":0,"gas":"0x0","gasCost":"0x0","memSize":128,"stack":["0x74","0x50","0x0"],"depth":1,"refund":0,"opName":"STOP"}
{"pass":true,"gasUsed":"0xd0de","stateRoot":"0xa282415f5c6c41917f42abcb6dd340dbcd1a1f439362bff7ddf66a8c33484487"}

```

## Follow-up

Commit which caused the bug: https://github.com/ethereum/evmone/commit/0b426937
Fix: https://github.com/ethereum/evmone/pull/893