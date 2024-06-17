Something related to TSTORE, reported in https://github.com/ethereum/execution-specs/issues/911

Original fuzzer-finding
```
INFO [03-21|13:19:35.821] Shortcutting through abort
Consensus error
Testcase: /fuzztmp/00024752-mixed-1.json
- gethbatch-0: ./gethbatch-0-output.jsonl
  - command: /gethvm --json --noreturndata --nomemory statetest
- eelsbatch-0: ./eelsbatch-0-output.jsonl
  - command: /ethereum-spec-evm statetest --json --noreturndata --nomemory
- nethbatch-0: ./nethbatch-0-output.jsonl
  - command: /nethtest -x --trace -m
- besubatch-0: ./besubatch-0-output.jsonl
  - command: /evmtool/bin/evm --nomemory --notime --json state-test
- erigonbatch-0: ./erigonbatch-0-output.jsonl
  - command: /erigon_vm --json --noreturndata --nomemory statetest
- nimbus-0: ./nimbus-0-output.jsonl
  - command: /nimbvm --json --noreturndata --nomemory --nostorage /fuzztmp/00024752-mixed-1.json
- evmone-0: ./evmone-0-output.jsonl
  - command: /evmone --trace /fuzztmp/00024752-mixed-1.json
- revm-0: ./revm-0-output.jsonl
  - command: /revme statetest --json /fuzztmp/00024752-mixed-1.json
-------
prev:           both: {"depth":1,"pc":290,"gas":248898,"op":92,"opName":"TLOAD","stack":["0x4"]}
diff:    gethbatch-0: {"depth":1,"pc":291,"gas":248798,"op":80,"opName":"POP","stack":["0x0"]}
diff:    eelsbatch-0: {"depth":1,"pc":291,"gas":248798,"op":80,"opName":"POP","stack":["0x1"]}
INFO [03-21|13:19:36.110] Waiting for processes to exit

```
Explainer: they were both in agreement up to `TLOAD`, but in the next op they differed in the stack.
I have verified that eels is the odd one out, the others agree with geth.

Here is the minimized test:
```
{
  "00024752-mixed-1": {
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
        "code": "0x7f7f6005600255600160045d6000600060006000600060065af1506000545060016000527f546000527f503a1a1cfa8877600560045d313b821d61600354506003545060036020527f5c506005606020526000604053605d60415360606042536000604353605c60446040527f53605060455360606046536002604753605c60485360506049536060604a536060605260206080536060608153604b6082536053608353606060845360606085536060608653604c608753605360885360606089536000608a536060608b53604d608c536053608d536060608e5360f3608f536060609053604e60915360536092536060609353604f6094536060609553600060965360f36097536000609860006000f560006000600060006000855af2505060045c50600760015d60005450600354506000600060006000600060f45af15060035450246000600060006000600060065af1507f7f767d78991c686f6d7f600160045560026004556003545060035450600160046000527f5d6000527f600160035560",
        "storage": {
          "0x0000000000000000000000000000000000000000000000000000000000000002": "0x0000000000000000000000000000000000000000000000000000000000000007"
        },
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
        "0x54209a9947898df95aec05ca1f6cb1d561b4211196025ee52cf812152c62e9b6aa348fadd9ac1095"
      ],
      "gasLimit": [
        "0x12c5f"
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

Geth ends with
```
{"pc":287,"op":80,"gas":"0x69","gasCost":"0x2","memSize":160,"stack":["0x6981f674fbaf9d6f293eaffc5855402eb64b8cbd"],"depth":1,"refund":0,"opName":"POP"}
{"pc":288,"op":96,"gas":"0x67","gasCost":"0x3","memSize":160,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":290,"op":92,"gas":"0x64","gasCost":"0x64","memSize":160,"stack":["0x4"],"depth":1,"refund":0,"opName":"TLOAD"}
{"pc":291,"op":80,"gas":"0x0","gasCost":"0x2","memSize":160,"stack":["0x0"],"depth":1,"refund":0,"opName":"POP","error":"out of gas"}
{"output":"","gasUsed":"0xd7d7","error":"out of gas"}
{"stateRoot": "0x7bfb83635d46a6250419c942b189a6dec350d140ec57cf6cb8030a4d26747441"}
```
eels ends with
```
{"pc":286,"op":80,"gas":"0x6b","gasCost":"0x2","memSize":160,"stack":["0x6981f674fbaf9d6f293eaffc5855402eb64b8cbd","0x0"],"depth":1,"refund":0,"opName":"POP"}
{"pc":287,"op":80,"gas":"0x69","gasCost":"0x2","memSize":160,"stack":["0x6981f674fbaf9d6f293eaffc5855402eb64b8cbd"],"depth":1,"refund":0,"opName":"POP"}
{"pc":288,"op":96,"gas":"0x67","gasCost":"0x3","memSize":160,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":290,"op":92,"gas":"0x64","gasCost":"0x64","memSize":160,"stack":["0x4"],"depth":1,"refund":0,"opName":"TLOAD"}
{"pc":291,"op":80,"gas":"0x0","gasCost":"0x2","memSize":160,"stack":["0x1"],"depth":1,"refund":0,"opName":"POP","error":"OutOfGasError"}
{"output":"","gasUsed":"0xd7d7","error":"OutOfGasError"}
{"stateRoot": "0x7bfb83635d46a6250419c942b189a6dec350d140ec57cf6cb8030a4d26747441"}
```

------------------------

Note, I don't think the out-of-error exit is related. Before minimizing the testcase:


geth
```
{"pc":287,"op":80,"gas":"0x3cc47","gasCost":"0x2","memSize":160,"stack":["0x6981f674fbaf9d6f293eaffc5855402eb64b8cbd"],"depth":1,"refund":0,"opName":"POP"}
{"pc":288,"op":96,"gas":"0x3cc45","gasCost":"0x3","memSize":160,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":290,"op":92,"gas":"0x3cc42","gasCost":"0x64","memSize":160,"stack":["0x4"],"depth":1,"refund":0,"opName":"TLOAD"}
{"pc":291,"op":80,"gas":"0x3cbde","gasCost":"0x2","memSize":160,"stack":["0x0"],"depth":1,"refund":0,"opName":"POP"}
{"pc":292,"op":96,"gas":"0x3cbdc","gasCost":"0x3","memSize":160,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
```

eels
```
{"pc":287,"op":80,"gas":"0x3cc47","gasCost":"0x2","memSize":160,"stack":["0x6981f674fbaf9d6f293eaffc5855402eb64b8cbd"],"depth":1,"refund":0,"opName":"POP"}
{"pc":288,"op":96,"gas":"0x3cc45","gasCost":"0x3","memSize":160,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":290,"op":92,"gas":"0x3cc42","gasCost":"0x64","memSize":160,"stack":["0x4"],"depth":1,"refund":0,"opName":"TLOAD"}
{"pc":291,"op":80,"gas":"0x3cbde","gasCost":"0x2","memSize":160,"stack":["0x1"],"depth":1,"refund":0,"opName":"POP"}
{"pc":292,"op":96,"gas":"0x3cbdc","gasCost":"0x3","memSize":160,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
```
