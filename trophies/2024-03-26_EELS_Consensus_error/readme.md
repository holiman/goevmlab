EELS ticket: https://github.com/ethereum/execution-specs/issues/914

```
INFO [03-26|12:37:35.814] Shortcutting through abort
Consensus error
Testcase: /fuzztmp/00002797-mixed-67.json
- eelsbatch-0: /fuzztmp/eelsbatch-0-output.jsonl
  - command: /ethereum-spec-evm statetest --json --noreturndata --nomemory
- nethbatch-0: /fuzztmp/nethbatch-0-output.jsonl
  - command: /neth/nethtest -x --trace -m
- besubatch-0: /fuzztmp/besubatch-0-output.jsonl
  - command: /evmtool/bin/evm --nomemory --notime --json state-test
- erigonbatch-0: /fuzztmp/erigonbatch-0-output.jsonl
  - command: /erigon_vm --json --noreturndata --nomemory statetest
- nimbus-0: /fuzztmp/nimbus-0-output.jsonl
  - command: /nimbvm --json --noreturndata --nomemory --nostorage /fuzztmp/00002797-mixed-67.json
- evmone-0: /fuzztmp/evmone-0-output.jsonl
  - command: /evmone --trace /fuzztmp/00002797-mixed-67.json
- revm-0: /fuzztmp/revm-0-output.jsonl
  - command: /revme statetest --json /fuzztmp/00002797-mixed-67.json

To view the difference with tracediff:
        tracediff /fuzztmp/eelsbatch-0-output.jsonl /fuzztmp/nethbatch-0-output.jsonl
-------
prev:           both: {"depth":1,"pc":57,"gas":7978861,"op":245,"opName":"CREATE2","stack":["0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x0","0x11531063346aa0fa774654"]}
diff:    eelsbatch-0: {"depth":1,"pc":58,"gas":7869055,"op":123,"opName":"PUSH28","stack":["0x50","0xab","0x70","0xcc","0x6c","0x0"]}
diff:    nethbatch-0: {"stateRoot":"0x33533c6e741de8fa56b689559a834e73ba2c050fa7a096f4b6b1d0c9018735c9"}
```
I've verified that `eels` is the odd one out, all other clients agree.

Minmized testcase `00002797-mixed-67.json.min`:
```json
{
  "00002797-mixed-67": {
    "env": {
      "currentCoinbase": "b94f5374fce5edbc8e2a8697c15331677e6ebf0b",
      "currentDifficulty": "0x200000",
      "currentRandom": "0x0000000000000000000000000000000000000000000000000000000000020000",
      "currentGasLimit": "0x26e1f476fe1e22",
      "currentNumber": "0x1",
      "currentTimestamp": "0x3e8",
      "previousHash": "0x044852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d",
      "currentBaseFee": "0x10"
    },
    "pre": {
      "0x00000000000000000000000000000000000000f1": {
        "code": "0x6043605060ab607060cc606c609251194465049d3a5b5b0172fd5b623153803c715b7c8482055e49071233651b6a11531063346aa0fa774654f57b",
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
        "0xbf71a3a5d0"
      ],
      "gasLimit": [
        "0x1ff81"
      ],
      "value": [
        "0x392aca"
      ],
      "sender": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b",
      "secretKey": "0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"
    },
    "out": "0x",
    "post": {
      "Cancun": [
        {
          "hash": "0x0000000000000000000000000000000000000000000000000000000000000000",
          "logs": "0x0000000000000000000000000000000000000000000000000000000000000000",
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
Outputs

eel:
```
root@84d25ae6ae59:/# cat eels-0-output.jsonl
{"depth":1,"pc":0,"gas":109865,"op":96,"opName":"PUSH1","stack":[]}
{"depth":1,"pc":2,"gas":109862,"op":96,"opName":"PUSH1","stack":["0x43"]}
{"depth":1,"pc":4,"gas":109859,"op":96,"opName":"PUSH1","stack":["0x43","0x50"]}
{"depth":1,"pc":6,"gas":109856,"op":96,"opName":"PUSH1","stack":["0x43","0x50","0xab"]}
{"depth":1,"pc":8,"gas":109853,"op":96,"opName":"PUSH1","stack":["0x43","0x50","0xab","0x70"]}
{"depth":1,"pc":10,"gas":109850,"op":96,"opName":"PUSH1","stack":["0x43","0x50","0xab","0x70","0xcc"]}
{"depth":1,"pc":12,"gas":109847,"op":96,"opName":"PUSH1","stack":["0x43","0x50","0xab","0x70","0xcc","0x6c"]}
{"depth":1,"pc":14,"gas":109844,"op":81,"opName":"MLOAD","stack":["0x50","0xab","0x70","0xcc","0x6c","0x92"]}
{"depth":1,"pc":15,"gas":109823,"op":25,"opName":"NOT","stack":["0x50","0xab","0x70","0xcc","0x6c","0x0"]}
{"depth":1,"pc":16,"gas":109820,"op":68,"opName":"DIFFICULTY","stack":["0x50","0xab","0x70","0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"]}
{"depth":1,"pc":17,"gas":109818,"op":101,"opName":"PUSH6","stack":["0xab","0x70","0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000"]}
{"depth":1,"pc":24,"gas":109815,"op":114,"opName":"PUSH19","stack":["0x70","0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x49d3a5b5b01"]}
{"depth":1,"pc":44,"gas":109812,"op":27,"opName":"SHL","stack":["0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x49d3a5b5b01","0xfd5b623153803c715b7c8482055e4907123365"]}
{"depth":1,"pc":45,"gas":109809,"op":106,"opName":"PUSH11","stack":["0x70","0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x0"]}
{"depth":1,"pc":57,"gas":109806,"op":245,"opName":"CREATE2","stack":["0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x0","0x11531063346aa0fa774654"]}
{"depth":1,"pc":58,"gas":0,"op":123,"opName":"PUSH28","stack":["0x50","0xab","0x70","0xcc","0x6c","0x0"]}
{"stateRoot":"0xa93360d73966aff182499970e779fcf4cefad3b99129879ce658eb7d806ee8d4"}
```
other client (erigon):
```
root@84d25ae6ae59:/# cat erigon-0-output.jsonl
{"depth":1,"pc":0,"gas":109865,"op":96,"opName":"PUSH1","stack":[]}
{"depth":1,"pc":2,"gas":109862,"op":96,"opName":"PUSH1","stack":["0x43"]}
{"depth":1,"pc":4,"gas":109859,"op":96,"opName":"PUSH1","stack":["0x43","0x50"]}
{"depth":1,"pc":6,"gas":109856,"op":96,"opName":"PUSH1","stack":["0x43","0x50","0xab"]}
{"depth":1,"pc":8,"gas":109853,"op":96,"opName":"PUSH1","stack":["0x43","0x50","0xab","0x70"]}
{"depth":1,"pc":10,"gas":109850,"op":96,"opName":"PUSH1","stack":["0x43","0x50","0xab","0x70","0xcc"]}
{"depth":1,"pc":12,"gas":109847,"op":96,"opName":"PUSH1","stack":["0x43","0x50","0xab","0x70","0xcc","0x6c"]}
{"depth":1,"pc":14,"gas":109844,"op":81,"opName":"MLOAD","stack":["0x50","0xab","0x70","0xcc","0x6c","0x92"]}
{"depth":1,"pc":15,"gas":109823,"op":25,"opName":"NOT","stack":["0x50","0xab","0x70","0xcc","0x6c","0x0"]}
{"depth":1,"pc":16,"gas":109820,"op":68,"opName":"DIFFICULTY","stack":["0x50","0xab","0x70","0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"]}
{"depth":1,"pc":17,"gas":109818,"op":101,"opName":"PUSH6","stack":["0xab","0x70","0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000"]}
{"depth":1,"pc":24,"gas":109815,"op":114,"opName":"PUSH19","stack":["0x70","0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x49d3a5b5b01"]}
{"depth":1,"pc":44,"gas":109812,"op":27,"opName":"SHL","stack":["0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x49d3a5b5b01","0xfd5b623153803c715b7c8482055e4907123365"]}
{"depth":1,"pc":45,"gas":109809,"op":106,"opName":"PUSH11","stack":["0x70","0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x0"]}
{"depth":1,"pc":57,"gas":109806,"op":245,"opName":"CREATE2","stack":["0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x0","0x11531063346aa0fa774654"]}
{"stateRoot":"0xa93360d73966aff182499970e779fcf4cefad3b99129879ce658eb7d806ee8d4"}
```
Diff on minimized testcase:
```
prev:           both: {"depth":1,"pc":57,"gas":109806,"op":245,"opName":"CREATE2","stack":["0xcc","0x6c","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x20000","0x0","0x11531063346aa0fa774654"]}
diff:         eels-0: {"depth":1,"pc":58,"gas":0,"op":123,"opName":"PUSH28","stack":["0x50","0xab","0x70","0xcc","0x6c","0x0"]}
diff:       erigon-0: {"stateRoot":"0x3eae60a44b8bc72e14f80dbc7ad73c0eb8d6d5152e1eaf9e66b739f8ecd89c37"}
```
It looks like you are emitting/executing an op even though there's `0` gas left?