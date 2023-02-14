### Repro

`repro` is a utility to reproduce a line in a (potentially large) trace. It is very 
simplistic and naive, which works well in some situations. For now, it just looks
at the stack and produces a push-sequence to mimic that stack. 

Example, telling it to focos on step `7` of the trace at `./testdata/1.geth.jsonl`:

```
$ go run . ./testdata/1.geth.jsonl 7 
INFO [02-10|09:43:33.486] Read traces                              steps=9
INFO [02-10|09:43:33.486] Read traces                              steps=9 target step=7
Target line:

        {"pc":15337,"op":26,"gas":"0x22","gasCost":"0x0","memSize":0,"stack":["0x10000000000000000","0x77676767676760000000000000001002e000000000000040000000e000000000","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x0","0x0"],"depth":1,"refund":0,"opName":"BYTE"}

Reproing stack, at step 7: 
         push 0x0
         push 0x0
         push 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
         push 0x77676767676760000000000000001002e000000000000040000000e000000000
         push 0x10000000000000000
Adding op
        BYTE
Code: 0x600060007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f77676767676760000000000000001002e000000000000040000000e000000000680100000000000000001a
```

If you only have the interesting line, and not the full stack, you need to do a bit of special magic, see `./testdata/line.txt`. 
Namely, add an extra line with an empty `{}` -- this is to ensure that the parser sees it as `jsonl`, and does not 
confuse it with another json-based trace-format returned by  legacy rpc. 

```
$ go run . ./testdata/line.txt 
INFO [02-10|09:45:53.448] Read traces                              steps=1
INFO [02-10|09:45:53.448] Read traces                              steps=1 target step=0
Target line:

        {"pc":15337,"op":26,"gas":"0x22","gasCost":"0x0","memSize":0,"stack":["0x0","0x0","0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff","0x77676767676760000000000000001002e000000000000040000000e000000000","0x10000000000000000"],"depth":1,"refund":0,"opName":"BYTE"}

Reproing stack, at step 0: 
         push 0x10000000000000000
         push 0x77676767676760000000000000001002e000000000000040000000e000000000
         push 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
         push 0x0
         push 0x0
Adding op
        BYTE
Code: 0x680100000000000000007f77676767676760000000000000001002e000000000000040000000e0000000007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff600060001a
```