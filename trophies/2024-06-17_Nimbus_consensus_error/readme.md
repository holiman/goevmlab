Fuzzer report:
```
Consensus error
Testcase: /fuzztmp/00028899-mixed-8.json
- gethbatch-0: /fuzztmp/gethbatch-0-output.jsonl
  - command: /gethvm --json --noreturndata --nomemory statetest
- eelsbatch-0: /fuzztmp/eelsbatch-0-output.jsonl
  - command: /ethereum-spec-evm statetest --json --noreturndata --nomemory
- nethbatch-0: /fuzztmp/nethbatch-0-output.jsonl
  - command: /neth/nethtest -x --trace -m
- erigonbatch-0: /fuzztmp/erigonbatch-0-output.jsonl
  - command: /erigon_vm --json --noreturndata --nomemory statetest
- nimbus-0: /fuzztmp/nimbus-0-output.jsonl
  - command: /nimbvm --json --noreturndata --nomemory --nostorage /fuzztmp/00028899-mixed-8.json
- evmone-0: /fuzztmp/evmone-0-output.jsonl
  - command: /evmone --trace /fuzztmp/00028899-mixed-8.json
- revm-0: /fuzztmp/revm-0-output.jsonl
  - command: /revme statetest --json /fuzztmp/00028899-mixed-8.json

To view the difference with tracediff:
        tracediff /fuzztmp/gethbatch-0-output.jsonl /fuzztmp/eelsbatch-0-output.jsonl
-------
prev:           both: {"depth":1,"pc":17,"gas":7978860,"op":64,"opName":"BLOCKHASH","stack":["0xb3","0x97","0xd1","0xc8","0x7d","0x1000000000000000000000000000000000000000"]}
diff:    gethbatch-0: {"depth":1,"pc":18,"gas":7978840,"op":130,"opName":"DUP3","stack":["0xb3","0x97","0xd1","0xc8","0x7d","0x0"]}
diff:       nimbus-0: {"depth":1,"pc":18,"gas":7978840,"op":130,"opName":"DUP3","stack":["0xb3","0x97","0xd1","0xc8","0x7d","0x44852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d"]}

```
After minimization, the problem is pretty straight-forward. The operation `BLOCKHASH(0x1000000000000000000000000000000000000000)` when performed
at block `1` ought to result in `0` on the stack, but nimbus vm returns `0x44852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d`.
Note, the hash `0x044852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d` is `Keccak256("0")`, i.e the hash of `0`, not the hash of `0x1000000000000000000000000000000000000000`. So
it would appear that the error is due to some truncation of the higher bits of the input.

```
root@eb653f4cc6dd:/# /gethvm --json --noreturndata --nomemory statetest /fuzztmp/nimfail.json.min.2
{"pc":0,"op":127,"gas":"0x15030","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH32"}
{"pc":33,"op":95,"gas":"0x1502d","gasCost":"0x2","memSize":0,"stack":["0x44852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d"],"depth":1,"refund":0,"opName":"PUSH0"}
{"pc":34,"op":85,"gas":"0x1502b","gasCost":"0x5654","memSize":0,"stack":["0x44852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d","0x0"],"depth":1,"refund":0,"opName":"SSTORE"}
{"pc":35,"op":0,"gas":"0xf9d7","gasCost":"0x0","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"STOP"}
{"output":"","gasUsed":"0x5659"}
{"stateRoot": "0xac0905a39a462c7425051e1ff361297e41b988d5b7a6c206322e5a231c45011f"}
[
  {
    "name": "00028899-mixed-8",
    "pass": true,
    "stateRoot": "0xac0905a39a462c7425051e1ff361297e41b988d5b7a6c206322e5a231c45011f",
    "fork": "Cancun"
  }
]
```
Geth
```
root@eb653f4cc6dd:/# /nimbvm --json --noreturndata --nomemory --nostorage /fuzztmp/nimfail.json.min.2
{"pc":0,"op":127,"gas":"0x15030","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH32"}
{"pc":33,"op":95,"gas":"0x1502d","gasCost":"0x2","memSize":0,"stack":["0x44852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d"],"depth":1,"refund":0,"opName":"PUSH0"}
{"pc":34,"op":85,"gas":"0x1502b","gasCost":"0x5654","memSize":0,"stack":["0x44852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d","0x0"],"depth":1,"refund":0,"opName":"SSTORE"}
{"output":"","gasUsed":"0x5659"}
{"stateRoot":"0xac0905a39a462c7425051e1ff361297e41b988d5b7a6c206322e5a231c45011f"}
[
  {
    "name": "00028899-mixed-8",
    "pass": true,
    "stateRoot": "0xac0905a39a462c7425051e1ff361297e41b988d5b7a6c206322e5a231c45011f",
    "fork": "Cancun",
    "error": ""
  }
]
```
Minimized and filled statetest:
```json
{
  "00028899-mixed-8": {
    "env": {
      "currentCoinbase": "b94f5374fce5edbc8e2a8697c15331677e6ebf0b",
      "currentDifficulty": "0x200000",
      "currentRandom": "0x0000000000000000000000000000000000000000000000000000000000200000",
      "currentGasLimit": "0x47f1f476fe1e22",
      "currentNumber": "0x1",
      "currentTimestamp": "0x3e8",
      "previousHash": "0x044852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d",
      "currentBaseFee": "0x10"
    },
    "pre": {
      "0x00000000000000000000000000000000000000f1": {
        "_code": "0x334103405f55",
        "code": "0x7f044852b2a670ade5407e78fb2863c51de9fcb96542a07186fe3aeda6bb8a116d5f55",
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
        "0x996b223caadb5f"
      ],
      "gasLimit": [
        "0x1a2a8"
      ],
      "value": [
        "0x"
      ],
      "sender": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b",
      "secretKey": "0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"
    },
    "out": "0x",
    "post": {
      "Cancun": [
        {
          "hash": "0xac0905a39a462c7425051e1ff361297e41b988d5b7a6c206322e5a231c45011f",
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
