### `Shanghai` consensus bug

This is a test which causes consensus-failure on `besu`, but only on the (not yet active) `Shanghai` fork. It does not 
trigger on `Merge` (or earlier), so it is not a live consensus issue.

The stateroot acccording to other clients is `0x0de332eee2c84244c64a7191027029d2a744836009c33f258fc9d651329d097b`, but`
besu` obtains `0x28bcaeeee2532c712f103227bd4a08306723b9fb7d2c86835f65d5eca2faf8b3`. 
The in-consensus-clients end on an `OOG` related to `CREATE`, so most likely Besu is differing on 
the implementation "EIP-3860: Limit and meter initcode", which is part of Shanghai.  

### Fix

["moves check for init code length before balance check"](https://github.com/hyperledger/besu/pull/5077)