## Blake fuzzer

This tool fuzzes geth `evm` versus parity `parith-evm`. 
Usage: 
```
./blakefuzz --geth /home/user/go/src/github.com/ethereum/go-ethereum/build/bin/evm --parity /home/user/go/src/github.com/holiman/goevmlab/parity-evm
Fuzzing started 
Fuzzing started 
Fuzzing started 
Fuzzing started 
Generator started 
Generator started 
file /tmp/blake-0-blaketest-1.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-1-blaketest-0.json: stats: steps: 98, maxdepth: 1
file /tmp/blake-0-blaketest-2.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-0-blaketest-3.json: stats: steps: 98, maxdepth: 1
file /tmp/blake-0-blaketest-0.json: stats: steps: 98, maxdepth: 1
file /tmp/blake-1-blaketest-1.json: stats: steps: 98, maxdepth: 1
file /tmp/blake-1-blaketest-2.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-0-blaketest-5.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-1-blaketest-3.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-0-blaketest-4.json: stats: steps: 98, maxdepth: 1
file /tmp/blake-0-blaketest-6.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-0-blaketest-8.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-1-blaketest-4.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-0-blaketest-7.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-0-blaketest-9.json: stats: steps: 97, maxdepth: 1
file /tmp/blake-1-blaketest-6.json: stats: steps: 98, maxdepth: 1
file /tmp/blake-1-blaketest-5.json: stats: steps: 97, maxdepth: 1
...
```