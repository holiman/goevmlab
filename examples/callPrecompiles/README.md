# EIP 2046

This is a benchmarking tool for testing [EIP 2046: Reduced gas cost for static calls made to precompiles](https://eips.ethereum.org/EIPS/eip-2046).
All benchmarks are basd on geth.

## Approach


- We use the `identity` precompile (address `0x4`), with zero input 
- The precompile has cost `15`, 
- and does not execute any code
 
One contract is used:

Contract `a`:
```
while true:
   staticcall(gas, 0x4, ...) 
```

The actual code of `a` is 
```
	PC       // This is a NOP, to make JUMPDEST wind up at 1
	JUMPDEST    // []
	MSIZE       // [0] out size
	MSIZE       // [0,0] out offset
	MSIZE       // [0,0,0] insize
	MSIZE       // [0,0,0,0] inoffset
	PUSH1 4     // [0,0,0,0,4] address
	GAS         // [0,0,0,0,4,gas] Gas
	STATICCALL  // [1] pops 6, pushes 1
	JUMP        // []
```
Since it does no mem expansion, we use `MSIZE` to push zeroes on the stack (cost `3` instead of `5`)

Resulting genesis alloc:
```json
{
 "0x000000000000000000000000000000000000ff0a": {
  "code": "0x585b5959595960045afa56",
  "balance": "0xffffffff"
 }
}
```



## Pre EIP-2046

With `10M` gas, we can do the call `13569` times, which takes around `20ms`

```
Time elapsed: 23.834943ms
Time elapsed: 65.726332ms
Time elapsed: 14.182627ms
Time elapsed: 14.274849ms
```

Rough calculation: `10000 / 13569 ~= 737 gas`  per round, which, accounting for `715` for the 
call means that the remainng  ops in the loop are around `22 gas`

## Post EIP-2046

To simulate EIP-2046, geth was modified in the following to make all `STATICCALL` cost `40`: 

```diff 
diff --git a/core/vm/jump_table.go b/core/vm/jump_table.go
index b26b55284c..bb3dac78ae 100644
--- a/core/vm/jump_table.go
+++ b/core/vm/jump_table.go
@@ -130,7 +130,7 @@ func newByzantiumInstructionSet() JumpTable {
        instructionSet := newSpuriousDragonInstructionSet()
        instructionSet[STATICCALL] = operation{
                execute:     opStaticCall,
-               constantGas: params.CallGasEIP150,
+               constantGas: 40,
                dynamicGas:  gasStaticCall,
                minStack:    minStack(6, 1),
                maxStack:    maxStack(6, 1),
```

With `10M` gas, we can do `129870` calls. 

Sanity check: `~77` per loop, so with `40+15` as cost, the remainder is `22` gas, which matches the result above. 
The execution time varied quite a lot between warmed-up or not:
```
Time elapsed: 220.176693ms
Time elapsed: 149.782772ms
Time elapsed: 145.534365ms
Time elapsed: 148.180509ms
```
I think `~150ms` is closest to the truth. 

If we mod `geth` a bit more, to remove those `15` gas, the results are
, with `161291` loops:

```
Time elapsed: 342.355984ms
Time elapsed: 190.665687ms
Time elapsed: 196.041588ms
Time elapsed: 199.760443ms
```
Still around `200ms`.

## Summary

Burning through `10M` gas by calling precompiles would still take below `150ms` on a standard laptop, which is within reason. 

The bigger problem, IMO, is that _all_ existing precompiles needs to be individually benchmarked, since currently some
of them are priced with the intrinsic `700` in mind. For example, the `ModExp` and `Blake2f` precompiles have very low
gas usage -- the bulk of the cost for calling them is the `700`. If that is lowered, we need to also
evaluate if those need to be raised.

I see no blocker to this EIP, other than that the EIP should also include benchkmarks and possibly 
repricing of (some of the) precompiles. 