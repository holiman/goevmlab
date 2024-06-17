## Besu consensus error

Version
```
root@98bafd804a49:/# /evmtool/bin/evm --version
Hyperledger Besu evm 24.6-develop-529bd33
```
Fuzzer output
```
Consensus error
Testcase: /fuzztmp/00062761-mixed-1.json
- gethbatch-0: /fuzztmp/gethbatch-0-output.jsonl
  - command: /gethvm --json --noreturndata --nomemory statetest
- eelsbatch-0: /fuzztmp/eelsbatch-0-output.jsonl
  - command: /ethereum-spec-evm statetest --json --noreturndata --nomemory
- nethbatch-0: /fuzztmp/nethbatch-0-output.jsonl
  - command: /neth/nethtest -x --trace -m
- besubatch-0: /fuzztmp/besubatch-0-output.jsonl
  - command: /evmtool/bin/evm --nomemory --notime --json state-test
- erigonbatch-0: /fuzztmp/erigonbatch-0-output.jsonl
  - command: /erigon_vm --json --noreturndata --nomemory statetest
- nimbus-0: /fuzztmp/nimbus-0-output.jsonl
  - command: /nimbvm --json --noreturndata --nomemory --nostorage /fuzztmp/00062761-mixed-1.json
- evmone-0: /fuzztmp/evmone-0-output.jsonl
  - command: /evmone --trace /fuzztmp/00062761-mixed-1.json
- revm-0: /fuzztmp/revm-0-output.jsonl
  - command: /revme statetest --json /fuzztmp/00062761-mixed-1.json

To view the difference with tracediff:
        tracediff /fuzztmp/gethbatch-0-output.jsonl /fuzztmp/eelsbatch-0-output.jsonl
-------
prev:           both: {"depth":1,"pc":151,"gas":7978029,"op":240,"opName":"CREATE","stack":["0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33","0x8273017316916c16891d123103f23d5a6474725a93160303"]}
diff:    gethbatch-0: {"stateRoot":"0xf87ea454a5d4d7afdf61317c2f7c2feb42579c4657ae4f79a974c8b3eafdf586"}
diff:    besubatch-0: {"depth":1,"pc":152,"gas":7931597,"op":130,"opName":"DUP3","stack":["0x22","0xbe","0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x0"]}
```
It seems to be related to `CREATE`:
```
root@98bafd804a49:/# tail -n5 /fuzztmp/gethbatch-0-output.jsonl
{"depth":1,"pc":124,"gas":7978039,"op":4,"opName":"DIV","stack":["0xbe","0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042"]}
{"depth":1,"pc":125,"gas":7978034,"op":54,"opName":"CALLDATASIZE","stack":["0x22","0xbe","0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799"]}
{"depth":1,"pc":126,"gas":7978032,"op":119,"opName":"PUSH24","stack":["0xbe","0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33"]}
{"depth":1,"pc":151,"gas":7978029,"op":240,"opName":"CREATE","stack":["0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33","0x8273017316916c16891d123103f23d5a6474725a93160303"]}
{"stateRoot":"0xf87ea454a5d4d7afdf61317c2f7c2feb42579c4657ae4f79a974c8b3eafdf586"}
root@98bafd804a49:/# tail -n5 /fuzztmp/besubatch-0-output.jsonl
{"depth":1,"pc":126,"gas":7978032,"op":119,"opName":"PUSH24","stack":["0xbe","0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33"]}
{"depth":1,"pc":151,"gas":7978029,"op":240,"opName":"CREATE","stack":["0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33","0x8273017316916c16891d123103f23d5a6474725a93160303"]}
{"depth":1,"pc":152,"gas":7931597,"op":130,"opName":"DUP3","stack":["0x22","0xbe","0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x0"]}
{"depth":1,"pc":153,"gas":7931594,"op":57,"opName":"CODECOPY","stack":["0xbe","0x1","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x0","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47"]}
{"stateRoot":"0xf87ea454a5d4d7afdf61317c2f7c2feb42579c4657ae4f79a974c8b3eafdf586"}

```
Split is
```
[geth, eels, nethermid, erigon, nimbus, evmone, reth] vs [besu]
```
## Reproducability

Minimized testcase is below. During minimization, the process managed to make the stateroots different, which is better for testability.

```json
{
  "00062761-mixed-1": {
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
        "code": "0x5f5f6b3b5d8d4a5c9f757c3242358d7a4906509357421a3e765d59515f78087a716276775f1415535b6f477204936e000a47183870643055465dfa6d07723a817c3d801972415f388c663d1c336756325f02651756495aa00605046e50423640500436778273017316916c16891d123103f23d5a6474725a93160303f0",
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
        "0xc6a4570168b41bcfdb62d4f9b6839f06210e839bf972ec380299910a3cc754e8c5aa10ebd0144bf1e900d634d07058496550cf"
      ],
      "gasLimit": [
        "0x10ae5"
      ],
      "value": [
        "0x2d05"
      ],
      "sender": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b",
      "secretKey": "0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"
    },
    "out": "0x",
    "post": {
      "Cancun": [
        {
          "hash": "0x225b4e26a029d5217ad34be3292fd89795bfa02da4566b590da160eb3b6d9a4b",
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
Geth trace
```
root@98bafd804a49:/# /gethvm --json statetest /fuzztmp/00062761-mixed-1.filled.min.json
{"pc":0,"op":95,"gas":"0xb5b9","gasCost":"0x2","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH0"}
{"pc":1,"op":95,"gas":"0xb5b7","gasCost":"0x2","memSize":0,"stack":["0x0"],"depth":1,"refund":0,"opName":"PUSH0"}
{"pc":2,"op":107,"gas":"0xb5b5","gasCost":"0x3","memSize":0,"stack":["0x0","0x0"],"depth":1,"refund":0,"opName":"PUSH12"}
{"pc":15,"op":122,"gas":"0xb5b2","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d"],"depth":1,"refund":0,"opName":"PUSH27"}
{"pc":43,"op":114,"gas":"0xb5af","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47"],"depth":1,"refund":0,"opName":"PUSH19"}
{"pc":63,"op":129,"gas":"0xb5ac","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a"],"depth":1,"refund":0,"opName":"DUP2"}
{"pc":64,"op":124,"gas":"0xb5a9","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47"],"depth":1,"refund":0,"opName":"PUSH29"}
{"pc":94,"op":54,"gas":"0xb5a6","gasCost":"0x2","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042"],"depth":1,"refund":0,"opName":"CALLDATASIZE"}
{"pc":95,"op":64,"gas":"0xb5a4","gasCost":"0x14","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042","0x33"],"depth":1,"refund":0,"opName":"BLOCKHASH"}
{"pc":96,"op":80,"gas":"0xb590","gasCost":"0x2","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042","0x0"],"depth":1,"refund":0,"opName":"POP"}
{"pc":97,"op":4,"gas":"0xb58e","gasCost":"0x5","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042"],"depth":1,"refund":0,"opName":"DIV"}
{"pc":98,"op":54,"gas":"0xb589","gasCost":"0x2","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799"],"depth":1,"refund":0,"opName":"CALLDATASIZE"}
{"pc":99,"op":119,"gas":"0xb587","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33"],"depth":1,"refund":0,"opName":"PUSH24"}
{"pc":124,"op":240,"gas":"0xb584","gasCost":"0x7d00","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33","0x8273017316916c16891d123103f23d5a6474725a93160303"],"depth":1,"refund":0,"opName":"CREATE","error":"out of gas: max initcode size exceeded: size 55193"}
{"output":"","gasUsed":"0xb5b9","error":"out of gas: max initcode size exceeded: size 55193"}
{"stateRoot": "0x225b4e26a029d5217ad34be3292fd89795bfa02da4566b590da160eb3b6d9a4b"}
[
  {
    "name": "00062761-mixed-1",
    "pass": true,
    "stateRoot": "0x225b4e26a029d5217ad34be3292fd89795bfa02da4566b590da160eb3b6d9a4b",
    "fork": "Cancun"
  }
]
```
Besu trace
```
root@98bafd804a49:/# /evmtool/bin/evm --nomemory --notime --json state-test /fuzztmp/00062761-mixed-1.filled.min.json
{"pc":0,"op":95,"gas":"0xb5b9","gasCost":"0x2","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH0"}
{"pc":1,"op":95,"gas":"0xb5b7","gasCost":"0x2","memSize":0,"stack":["0x0"],"depth":1,"refund":0,"opName":"PUSH0"}
{"pc":2,"op":107,"gas":"0xb5b5","gasCost":"0x3","memSize":0,"stack":["0x0","0x0"],"depth":1,"refund":0,"opName":"PUSH12"}
{"pc":15,"op":122,"gas":"0xb5b2","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d"],"depth":1,"refund":0,"opName":"PUSH27"}
{"pc":43,"op":114,"gas":"0xb5af","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47"],"depth":1,"refund":0,"opName":"PUSH19"}
{"pc":63,"op":129,"gas":"0xb5ac","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a"],"depth":1,"refund":0,"opName":"DUP2"}
{"pc":64,"op":124,"gas":"0xb5a9","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47"],"depth":1,"refund":0,"opName":"PUSH29"}
{"pc":94,"op":54,"gas":"0xb5a6","gasCost":"0x2","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042"],"depth":1,"refund":0,"opName":"CALLDATASIZE"}
{"pc":95,"op":64,"gas":"0xb5a4","gasCost":"0x14","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042","0x33"],"depth":1,"refund":0,"opName":"BLOCKHASH"}
{"pc":96,"op":80,"gas":"0xb590","gasCost":"0x2","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042","0x0"],"depth":1,"refund":0,"opName":"POP"}
{"pc":97,"op":4,"gas":"0xb58e","gasCost":"0x5","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x3d801972415f388c663d1c336756325f02651756495aa00605046e5042"],"depth":1,"refund":0,"opName":"DIV"}
{"pc":98,"op":54,"gas":"0xb589","gasCost":"0x2","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799"],"depth":1,"refund":0,"opName":"CALLDATASIZE"}
{"pc":99,"op":119,"gas":"0xb587","gasCost":"0x3","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33"],"depth":1,"refund":0,"opName":"PUSH24"}
{"pc":124,"op":240,"gas":"0xb584","gasCost":"0xb578","memSize":0,"stack":["0x0","0x0","0x3b5d8d4a5c9f757c3242358d","0x4906509357421a3e765d59515f78087a716276775f1415535b6f47","0x4936e000a47183870643055465dfa6d07723a","0xd799","0x33","0x8273017316916c16891d123103f23d5a6474725a93160303"],"depth":1,"refund":0,"opName":"CREATE"}
{"output":"","gasUsed":"0x10ad9","test":"00062761-mixed-1","fork":"Cancun","d":0,"g":0,"v":0,"postHash":"0x2d5cff0902bf31f0b19274cef7253dcc674b9f2df9d8b2ba8256e520fdd525da","postLogsHash":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","pass":false}
```

## Analysis

The geth-trace shows that it exits with an "max initcode size exceeded: size 55193". Those parts were touched recently, in Besu. I suppose that this commit is the culprit:

https://github.com/hyperledger/besu/commit/85d286aa85a68e19522f9da99dad607895b7e11f#diff-23062cb5b161cc36ccf31141b291e64df4a32002a98c8f4492eeda39c9a74d82R116

Fix: https://github.com/hyperledger/besu/pull/7233