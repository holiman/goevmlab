A consensus error was found on `revme`, the evm used in `reth`.

```
Consensus error
Testcase: /fuzztmp/00000022-mixed-6.json
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
  - command: /nimbvm --json --noreturndata --nomemory --nostorage /fuzztmp/00000022-mixed-6.json
- evmone-0: /fuzztmp/evmone-0-output.jsonl
  - command: /evmone --trace /fuzztmp/00000022-mixed-6.json
- revm-0: /fuzztmp/revm-0-output.jsonl
  - command: /revme statetest --json /fuzztmp/00000022-mixed-6.json

To view the difference with tracediff:
        tracediff /fuzztmp/gethbatch-0-output.jsonl /fuzztmp/eelsbatch-0-output.jsonl
-------
prev:           both: {"depth":122,"pc":2,"gas":1080,"op":84,"opName":"SLOAD","stack":["0x3"]}
diff:    gethbatch-0: {"depth":121,"pc":5846,"gas":17,"op":96,"opName":"PUSH1","stack":["0x0"]}
diff:         revm-0: {"depth":122,"pc":3,"gas":980,"op":80,"opName":"POP","stack":["0x0"]}

INFO [06-13|12:22:05.214] Waiting for processes to exit
```

Consensus-split is
```
[geth,besu,eels,nimbus,nethermind,evmone] vs [revme]
```

Minimized file `00000022-mixed-6.json.min`

```json
{
  "00000022-mixed-6": {
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
        "code": "0x6000600060006000600060f15af150600354507f600354506000600255600160045d6001600155600060045d60025450600454506000527f7f60005c50600460025d6003600255c4159639471638667f60065c50600060047f5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f7f60005d60006020527f600055600260025d7f466003545080f4600660015560027f545060026020527f546040527f507f60055450600260045d60036000527f5c507f60005c50600160015560015c606060527f40527f5060055c5060016000527f607f025d600160016020527f5d7f7f600260036080527f5560016060527f60045d607f055450600060025d6000606020527e55600060405260a0527f7f60015d60026060e0527f6080527e527f606000527f015d600260025d600060045d600060c0527f6040527f7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e0527f7f606020527f01536054600253608060c0527f527f6060527f60506003536060610140527f60610100527f04536004600553606040527f6000527f546060e0527e5260a0527f7f7f60405260610120527f80527f7f60065360506007536060600853602060097f610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f7f6020527f536060527f60610120527ff3610160527f600c6020526053604060e07f527f536060606080527f4153600d60c0527f604261610180527f0140527f53606101e0527f606043536000604460608052610100527f7f40527f5360f360456101a0527f53610200527f600060610160527fa0527f60e0527f604660006000f560006000606101205261610220527f01c0527f7e60006000855af150610180527f60a0527f5060096060527f610100610240527f527f60c06101e0527f527f60025d610140527f6000ff606101a0527f0cff7f60610260527f015450600554506002610200527f6003557f6060c052610120527f7f01610160610280527f526101c0527f7f545060e0527f7f610220527f6002606080527f035d600154506102a0527f600354506000527f6006616101e0527f016101610240527f80527f40527f54506102c0527f60026060e052610100527f7f015d7f60005c506000610200610260527f527f526102e0527f7f60a0527f6101a0527f60025c610160527f508b6000527ffe4a6d3139610280610300527f527f6020610220527f610120527f52610100527f6101c0527f7f60025c50600d610320527fff616102a0527f0180527f7f6005610240527f5450600160c0527f5450600260610340527f2052616101e0526102c0527f7f0140527f7f6000527f6061610260527f012061610360527f01a0527f527f0355600160606102e0527f40527f20527f610200527f54506006610380527f5450610280527f6060e0610160527f527f610300527feb606101c0527e5360606103a0527f600153610140610220527f526102a0527f7f60600060610320527f40527f52606103c0527f20602053606052610180526101e0527f7f7f606060616102c0527f610340527f6103e0527f0240527f20527f610100527f60604052610160527f7f21536002602253605361610400527f610360527f6102e0527f0200527f610260527f606101a0527f23536060606060610420527f527f245360610380527f80527f6060610300527f61012052610180527f610280610440527f527f610220527f7f60256103a0527f5360606101c0527f6026610320527f5360610460527f036027606060527f40527f536102a06103c0527f527f6053602853610240527f610480527f606060610340527f80606101a0527fa06101e0526103e0527f7f610140527f526104a0527f6102c0527f7f527f6029536000610360527f602a610260527f610400527f53606104c0527f60602b536004602c536053606061616102e0527f020052610380527f7f0161046104e0527f20527fc0527f80527f2d536061610280527f0160527f606060c0527f527f6103610500527e6103610440527fa0527f527f60a0527f60602e5360610220527ffd602f5361610520527f016102a0527fe052610460527f7f60606103c0527f6030610320527f53600560610540527f315360610180527f5360325360610480527f6060610240527fe06103e0527f61610560527f02c0527f527f610340527fa0527f606033616104a0527f0200527f60c0527f53610580527f60066060610400527f80527f34536060610161610261036104c0527f60527fe06105a0527f527f0260527fa0527f6035536000610420527f60365360f3616102206104e0526105c0527f7f527f0100527f6037610380527f536000610300527f6038610440527f60c0616105e0527f02610500527f80527f5260e0527f7f60006101c0527f60006103a0527ff56000610600527f606161046052610520527f7f024052610320527f7e600060a0527f60006102a0610620527f527f610120527f6103c052610540527f7f610480527f6000855af25050600360610640527f025561610340527f01e0527f60610260610560527f527f056101006104a0527f610660527f6103e0527f6102c0527f527fff60e0527f60036003610580527f550761036052610680527f7f600461016104c0527f40527f5450610400527f60035c5081616105a0527f616106a0527f0280526102e0527f7f0200527f60c06104e0527f52610380527f7f60026061616106c0527f05c0527f0420527e5d60f9ff610120527f6000600060610100610500527f61016106e0527f610300526105e0527f7f6061026103a052610440527f7fa0527f527f527e6102610700527f20527f610520527f60610600527e60f75af45060005c5060045c506006610361610720527f0460527f206103c0527f527f61610620527f0540527f6060e0527f616102c052610740527f7f0140527f015d7f61016102405261048052610640527f7f7f8052610560527f610760527f7f60026103e0527f6002610340527f5d60610120527f01610660527f60025d61610780527f02e06104a0610580527f527f527f60025c50600154507f610400527f610680526107a0527f7f5d5e3194866102610360527f60526105a0527f7f616104c0527f01606101a06107c0527f526106a0527f7f527f11008f61030052610420527f7f5611616105c0527f01006107e0527f527f3e6003616106c0527f04e0527f60610380527f610140527e527f60610280610800527f527e5d6105e0527f6061046106e0527f40527e5c5060006161610500527f0320610820527f527f01c0527f60006103a0527f610600610700527f527f610180527f60006000610840527f60610460527e60f8610520527f600052606102a052610720527f7f5a61062052610860527f7f606101610340527f206103c0527f5261016052610480610540610740527f52610880527f7f527f7f7f20610640527f6101e0527f5360f2602153606101a0527f205261616108a0527f0760527f02c052616103e0610560527f610660527f527f6104a0527f0360527f6108c0527f7f7f6050610780527f60225360606023536004602453605c6102610680527f616108e0527f0580527e527f6025536107a0527f616104c0527f610400527f0180527f610380610900527f527f60616106a0527f02e0527f616107c0527f05a0527f6101406101c0527f52610920527f7f50606104e0527f2653606060616106c0527f6107e0527f0420527f27536007610940527f606105c0527f406102206103a0527f527f527f6028610361610800527f6106e0610960527f527f0500527e527f53606060295360616105e0527f0440527f03606101610820610980527f527fa0526101e0610700527f527f7f2a536103c0610520527f527f6055602b616109a0527f0600610840527f527f6101606102405261610720527f03610460527f20527f7f6109c0527f527f536060602c610860527f53610540527f60610620527f20602d610740527f6109e0527f5360606103e0527f602e6060610880527f52605361610480527f0200527f6061610a00527f0161610640610760527f527f0560527fc06108a0527f610340527f5261026052610a20527f7f7f80536060610400527f60815361610780527f04a06108c0527f610660527f610a40527f527f6000610580527f6082610180527f536060608353602f6061076108e0527f610a60527fa0527f610360527f8461610680527f0220527f61042061046105a0527fc0527f610a80527f610900527f527f616107c0527f0280527f53605360855361016106a0527fe052610aa0527f7f60606086610920527f5360f360875361616107e0527f05c0527f0380526104610ac0527fe0527f7f606104406106610940527fc0527f527f60608853606101a061080052610ae0527f7f527f6102a0527f61026105e0527f610960527f40527f30608953606106e052610b00527f7f610500527f610820527f53608a536060616104610980527f60527f03a0527f610b20527f610200610600527f527f60610700527f610840527f8b5360316109a0527f608c610b40527f53606060610520527f6102c0527f8d5360006061026104610620610860526109610b60527fc0527f7f610720527f527f80527f60527f8e6103c0527f5360f3606101610540610b80527f527fc06109e0527f52610880527f608f6101e0610740527f5360610640527f61610ba0527f0220527f53610261610a00527f04a0527fe0526108a0527f7f6101e153606061610bc0527f6105610760527f60527f03e052610a20527f7f610660527f6101e253616108c0610be0527f527f0280527f60906101e353606061610461610a40527f0780527fc0527f01e4610c00527f536000610580526108e0527f610680527f7f6101e56103610a60527e527f5361610c20527f0400527f6107a0527f60606161024052610900527f7f01e6536000610a80527f610c40527f6104e0526106a0527f7f6105a0527f616102a06107c0527f527f610920527f01610c60527f610aa0527fe75360f06101e85360606161610420527f03206106c0527f527f01610c80527fe953610761610ac0527f0940527fe0527f606105c0527f610500527e6101ea53610ca0527f60606101eb610260527f610ae0527f53616106610960527fe0527f610800527f610cc0527f02c0527f600061610440526105e052610b00527f7f7f01ec5361052052610980610ce0527f527f7f60610340527f60610820527f610700527f610b20527f6101ed53600061610d00527f01ee53606061016109a0527fef53610600527f60006101f053610b40527f6108610d20527f40527f6104606105610720527f40527f526109c0527f7f60616102e0527f610b610d40527f60527f0280610360527f527f61610860527f0620527f846101f1536109e0527f610d60527f606107610b80527f40527f5a6101f25360f461610560527f01f3610461088052610d80527f7f80527f53610a00610ba0527f527f60506101f4610640527f53610760527f60610da0527f506101f56103610380527e6108610bc0527fa052610a20527f7f527f536101f6610dc0527f610580527f6102a052606061616161078052610be0527f7f0660527f04a0610a610de0527f40527f526108c0527f7f02c05360006102c15360f36102610c00527fc2536102610e00527fc361036105a05261610a60527f07a0527f7fa06108e0527f527f6106610c2052610e20527f7f80527f6000606103205260006104c0527f610a80527f6103405360f0610341610e40527f61610c40527f07610900527fc0527f5360606103426105c06106a052610aa052610e60527f7f7f527f5360610c60527e61034353606103c0527f610920527f6061036107e0610e80527f527f446104610ac0527f610c80527fe0527f53600061034553606106c0527f60610ea0527f616105e0610940527f527f03465360610ca0527f610ae0527e61610800527f03610ec0527f47536060610348536000616103e0526105006109610cc0527f60527f61610b00610ee0527f527f06e0527f527f7f0349536061610820527f0600527f6061610ce0527f034a610f00527f53600061034b61610b20527f0980527f53608561034c53605a6161070052610d610f20527e527f7f034d53610840527f60f1610361610b40527f0520527f6109a0527f61610f40527f062052610d20527f7f4e6104005260536104205360606104610720610b60527f610f60527f610860527f527f21610d40527f536109c0527f60506104225360616104235360610f80527f0361610640610b80527f527f61610d60527f0540527f042453610880526109e0610fa0527f527f7f604f610740527f6104255360610ba0610d80527f527f53610426536060610fc0527f610427536050610428536106610a00527f606108a0527f610da0527f527f610b610fe0527fc0527f6061610429610760527f610560527f53600361042a53605061610dc052611000527f7f0a20527f61042b610be0527f5360536108c0527f61042c5360616161068052611020527f7f610de0527f610780527f042d5360610a40610c00527f527f0361042e536061611040527f0580527f6108610e00527fe0527f5161042f53606061043053600061610c2052611060527f7f0431610a60527f536107610e20527fa0527f6106a0527f60f3610432610900611080527f527f53610433610c40527f60006000f0610e40527f6061610a80527f05a052606110a0527e6105c05360606107c0527f6105c153610c60527f610e60527f60610920527f6110c0527f6106c0527e61610aa0527f05c25360606105c35360006105c453610e80527f616110e0527f0c80527f60606105c5536107e0527f610940527f60610ac0527e6105c6536061611100527f0ea0527f84616106e0610ca0527f527f05c753605a6105c85360f46105c95360611120527f50616161610ec0527f0ae0527f0960527f6108610cc0527e527f05ca53605061611140527f05cb536105cc600061610ee0527f07005260f3610720610b00527f53610ce052611160527f7f610721610980527f60006000f0610f00527f61082052606061084053600061611180527f084153606061610d00527f0b20527f61084253610f20527f60006108436109a06111a0527f527f5360606108445360006108455360610d20527f606108610f40527f46610b6111c0527f40527f53600061084753606061084853606109c0527e61084953610d40610f606111e0527f527f527f608561084a53605a61610b60527f084b5360f161084c53605061084d611200527f5360610f80527f5061610d60527f6109e0526008610a0053604e610a610b8052611220527f7f01536053610a610fa0527f02536061610a03610d80527f536008610a045360611240527f4f610a05536060610a065361610fc0527f0ba0527f6000610a075360f3610da0611260527f527f610a08536000610a0960006000f560610fe0527e6000600060006000610b611280527fc0526085610b610dc0527fe053605a610be15360f2611000527f610be25360506112a0527f610be3536050610be453610be56000f3610de0526000610e00606110205260006112c0527f6110405360606110415360006110425360f561104353606061104453600061106112e0527f4553606061104653600061104753606061104853600061104953606061104a53611300527f600061104b53606061104c53600061104d53608561104e53605a61104f5360f161132052606161134053601061134153605061134253605361134353606061134453605061134553606161134653601061134753605161134853605361134953606061134a53605061134b53606161134c53601061134d53605261134e53605361134f5360616113505360106113515360536113525360606113535360006113545360f36113555361135660006000f06000600060006000845af45050",
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
        "0x1b7c2815682fb49c96def5ae0eff932dd2171d9108817115d7e4cdc8aeab6f96a2a5a5b16265cbddb1f4cfd74ca602c522f23a10e95af16db91621beaa1b0ddcc82ab6941b8400bbec0791"
      ],
      "gasLimit": [
        "0x24938d"
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

Output from the minimized version
```
INFO [06-13|12:49:14.996] Stats geth-0                             execSpeed=1.978s  longest=2.972476869s  count=2
INFO [06-13|12:49:14.996] Stats revm-0                             execSpeed=9.7864s longest=12.232949474s count=1
Consensus error
Testcase: /fuzztmp/00000022-mixed-6.json.min
- geth-0: /tmp/geth-0-output.jsonl
  - command: /gethvm --json --noreturndata --nomemory statetest /fuzztmp/00000022-mixed-6.json.min
- revm-0: /tmp/revm-0-output.jsonl
  - command: /revme statetest --json /fuzztmp/00000022-mixed-6.json.min

To view the difference with tracediff:
        tracediff /tmp/geth-0-output.jsonl /tmp/revm-0-output.jsonl
-------
prev:           both: {"depth":2,"pc":2,"gas":643,"op":84,"opName":"SLOAD","stack":["0x3"]}
diff:         geth-0: {"depth":1,"pc":5804,"gas":10,"op":96,"opName":"PUSH1","stack":["0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0"]}
diff:         revm-0: {"depth":2,"pc":3,"gas":543,"op":80,"opName":"POP","stack":["0x0"]}

```

Note: the execution on `revm` takes a _very_ long time:

```
root@9ad68c26cd2a:/# /revme  statetest --json   /fuzztmp/00000022-mixed-6.json.min 2>revmeout.txt
...
Finished execution. Total CPU time: 28.830958s
Encountered 1 errors out of 1 total tests

```
The trace is roughly 833K lines, and the discrepancy happens at the end, geth et al hit an OOG during an `SLOAD`, which reth doesn't agree with.
```
root@9ad68c26cd2a:/# tail gethout.txt -n 7
{"pc":2,"op":84,"gas":"0x283","gasCost":"0x834","memSize":0,"stack":["0x3"],"depth":2,"refund":0,"opName":"SLOAD","error":"out of gas"}
{"pc":5804,"op":96,"gas":"0xa","gasCost":"0x3","memSize":4960,"stack":["0x7f60005c50600460025d6003600255c4159639471638667f60065c5060006004","0x5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f","0x60005d60006020527f600055600260025d7f466003545080f460066001556002","0x545060026020527f546040527f507f60055450600260045d60036000527f5c50","0x60005c50600160015560015c606060527f40527f5060055c5060016000527f60","0x25d600160016020527f5d7f7f600260036080527f5560016060527f60045d60","0x6080527e527f606000527f015d600260025d600060045d600060c0527f604052","0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":5806,"op":96,"gas":"0x7","gasCost":"0x3","memSize":4960,"stack":["0x7f60005c50600460025d6003600255c4159639471638667f60065c5060006004","0x5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f","0x60005d60006020527f600055600260025d7f466003545080f460066001556002","0x545060026020527f546040527f507f60055450600260045d60036000527f5c50","0x60005c50600160015560015c606060527f40527f5060055c5060016000527f60","0x25d600160016020527f5d7f7f600260036080527f5560016060527f60045d60","0x6080527e527f606000527f015d600260025d600060045d600060c0527f604052","0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0","0x0"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":5808,"op":96,"gas":"0x4","gasCost":"0x3","memSize":4960,"stack":["0x7f60005c50600460025d6003600255c4159639471638667f60065c5060006004","0x5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f","0x60005d60006020527f600055600260025d7f466003545080f460066001556002","0x545060026020527f546040527f507f60055450600260045d60036000527f5c50","0x60005c50600160015560015c606060527f40527f5060055c5060016000527f60","0x25d600160016020527f5d7f7f600260036080527f5560016060527f60045d60","0x6080527e527f606000527f015d600260025d600060045d600060c0527f604052","0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0","0x0","0x0"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":5810,"op":96,"gas":"0x1","gasCost":"0x3","memSize":4960,"stack":["0x7f60005c50600460025d6003600255c4159639471638667f60065c5060006004","0x5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f","0x60005d60006020527f600055600260025d7f466003545080f460066001556002","0x545060026020527f546040527f507f60055450600260045d60036000527f5c50","0x60005c50600160015560015c606060527f40527f5060055c5060016000527f60","0x25d600160016020527f5d7f7f600260036080527f5560016060527f60045d60","0x6080527e527f606000527f015d600260025d600060045d600060c0527f604052","0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0","0x0","0x0","0x0"],"depth":1,"refund":0,"opName":"PUSH1","error":"out of gas"}
{"output":"","gasUsed":"0x243ce1","error":"out of gas"}
{"stateRoot": "0x8bc022b1d34b46b316440fb9d6d932d688033f74c8c12e7b39d1fa6d444fa837"}


root@9ad68c26cd2a:/# tail revmeout.txt  -n 11
{"pc":2,"op":84,"gas":"0x283","gasCost":"0x64","stack":["0x3"],"depth":2,"returnData":"0x","refund":"0x0","memSize":"0","opName":"SLOAD"}
{"pc":3,"op":80,"gas":"0x21f","gasCost":"0x2","stack":["0x0"],"depth":2,"returnData":"0x","refund":"0x0","memSize":"0","opName":"POP"}
{"pc":4,"op":96,"gas":"0x21d","gasCost":"0x3","stack":[],"depth":2,"returnData":"0x","refund":"0x0","memSize":"0","opName":"PUSH1"}
{"pc":6,"op":96,"gas":"0x21a","gasCost":"0x3","stack":["0x0"],"depth":2,"returnData":"0x","refund":"0x0","memSize":"0","opName":"PUSH1"}
{"pc":8,"op":85,"gas":"0x217","gasCost":"0x0","stack":["0x0","0x2"],"depth":2,"returnData":"0x","refund":"0x0","memSize":"0","opName":"SSTORE","error":"OutOfGas"}
{"pc":5804,"op":96,"gas":"0xa","gasCost":"0x3","stack":["0x7f60005c50600460025d6003600255c4159639471638667f60065c5060006004","0x5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f","0x60005d60006020527f600055600260025d7f466003545080f460066001556002","0x545060026020527f546040527f507f60055450600260045d60036000527f5c50","0x60005c50600160015560015c606060527f40527f5060055c5060016000527f60","0x25d600160016020527f5d7f7f600260036080527f5560016060527f60045d60","0x6080527e527f606000527f015d600260025d600060045d600060c0527f604052","0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0"],"depth":1,"returnData":"0x","refund":"0x0","memSize":"4960","opName":"PUSH1"}
{"pc":5806,"op":96,"gas":"0x7","gasCost":"0x3","stack":["0x7f60005c50600460025d6003600255c4159639471638667f60065c5060006004","0x5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f","0x60005d60006020527f600055600260025d7f466003545080f460066001556002","0x545060026020527f546040527f507f60055450600260045d60036000527f5c50","0x60005c50600160015560015c606060527f40527f5060055c5060016000527f60","0x25d600160016020527f5d7f7f600260036080527f5560016060527f60045d60","0x6080527e527f606000527f015d600260025d600060045d600060c0527f604052","0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0","0x0"],"depth":1,"returnData":"0x","refund":"0x0","memSize":"4960","opName":"PUSH1"}
{"pc":5808,"op":96,"gas":"0x4","gasCost":"0x3","stack":["0x7f60005c50600460025d6003600255c4159639471638667f60065c5060006004","0x5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f","0x60005d60006020527f600055600260025d7f466003545080f460066001556002","0x545060026020527f546040527f507f60055450600260045d60036000527f5c50","0x60005c50600160015560015c606060527f40527f5060055c5060016000527f60","0x25d600160016020527f5d7f7f600260036080527f5560016060527f60045d60","0x6080527e527f606000527f015d600260025d600060045d600060c0527f604052","0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0","0x0","0x0"],"depth":1,"returnData":"0x","refund":"0x0","memSize":"4960","opName":"PUSH1"}
{"pc":5810,"op":96,"gas":"0x1","gasCost":"0x0","stack":["0x7f60005c50600460025d6003600255c4159639471638667f60065c5060006004","0x5d6000527f6006545060085c5060005c50600260015d60025c5060026000527f","0x60005d60006020527f600055600260025d7f466003545080f460066001556002","0x545060026020527f546040527f507f60055450600260045d60036000527f5c50","0x60005c50600160015560015c606060527f40527f5060055c5060016000527f60","0x25d600160016020527f5d7f7f600260036080527f5560016060527f60045d60","0x6080527e527f606000527f015d600260025d600060045d600060c0527f604052","0x7f6060527f600260a0527f5d60016002557f7f6060606020527e53600460e052","0x60610100527f04536004600553606040527f6000527f546060e0527e5260a052","0x7f7f60405260610120527f80527f7f6006536050600753606060085360206009","0x610100527f53606060527f610140527f60c0527f60600a536000600b60a0527f","0x6020527f536060527f60610120527ff3610160527f600c6020526053604060e0","0x0","0x0","0x0","0x0"],"depth":1,"returnData":"0x","refund":"0x0","memSize":"4960","opName":"PUSH1","error":"OutOfGas"}
{"stateRoot":"0x8bc022b1d34b46b316440fb9d6d932d688033f74c8c12e7b39d1fa6d444fa837","logsRoot":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","output":"0x","gasUsed":2397069,"pass":false,"errorMsg":"logs root mismatch: got 0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347, expected 0x0000000000000000000000000000000000000000000000000000000000000000","evmResult":"Halt: OutOfGas(Basic)","postLogsHash":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","fork":"CANCUN","test":"00000022-mixed-6","d":0,"g":0,"v":0}
Error: Statetest(TestError { name: "00000022-mixed-6", kind: LogsRootMismatch { got: 0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347, expected: 0x0000000000000000000000000000000000000000000000000000000000000000 } })

```

even more minimal
```json
{
  "00000156-mixed-1": {
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
        "code": "0x5f5f5f5f5f60f15af17f60016002556005545060065450600054507f3e5008a06d86625276586a1b9c605f52505f5f5f61038d53605361038f5360616103905360026103915360b56103925360606103935360006103945360f36103955361039660006000f06000600060006000845af4",
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
        "0x9a583d8df426fcc320f9db97b4eb1780051e9d7942ce353fac97e0c5b55280235e31fd359f144fe6ad0d"
      ],
      "gasLimit": [
        "0x340dd9"
      ],
      "value": [
        "0xfb9c"
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
And finally, a minimalistic testcase which also produces differing stateroot:

```
{
  "00001692-mixed-0": {
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
        "code": "0x60005f5f5f5f60f15af27f60026002557f6000600060006000600060f55af150861a6079417f60065450605f52507f60602b536000602c536060602d536000602e536060606080527f2f6060527f5360a0527f600e603053605a60315360f4603253605060335360606034536060a0527f206060c0527f355360608052606060a053606060a153603660a253605360a353606060a460c060e0527f527f53600060a553606060a653603760a753605360a853606060a95360fd60aa610100527f536060e0527f6060ab53603860ac53605360ad53606060ae53603960af536060610120527f60b053600060610100527fb15360f360b253600060b360006000f56000600060610140527e6000845af45050d06001610120526060610140536000610141536055610142610160527f53604261014353605c61014453600761014553606061014653600a6101475360610180527fff61014853606061014953602061014a53606061014b53600061014c5360fd616101a05260016101c053604d6101c15360536101c25360616101c35360016101c453604e6101c55360606101c65360006101c75360f36101c85360006101c960006000f5",
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
        "0x468c7d2abe7d791f84e5bdb4eeb095dc995346fb19f5667f42462b44f110d3f3ac1ddcd5e6702b00e94d4ede0701f836fd"
      ],
      "gasLimit": [
        "0x35730c"
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

## Fix

Fix was included in https://github.com/bluealloy/revm/pull/1518

Although this was a real consensus issue, it was recent enough (5 days) to not have been
included into any `reth` release as of yet.