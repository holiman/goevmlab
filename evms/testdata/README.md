## Testdata output

This folder contains 
 - statetests, 
 - For each statetest: 
   - output from vms, `stdout` and `stderr` output

The statetests have been chosen because they trigger some quirk in a vm, e.g a statetest may 
trigger a negative refund in besu. If we want to change the besu-shim at some later point, 
when said bug has been fixed, we need to regenerate the outputs and check if the 
tests passes. 

## Command to generate these

The script below, after setting the binaries to use, should recreate the outputs 
in `traces` based on the inputs in `cases`. 

```bash
#!/bin/bash

# evm="/home/user/go/src/github.com/ethereum/go-ethereum/cmd/evm/evm"
# nethtest="/home/martin/workspace/nethtest"
# besuvm="/home/martin/workspace/besu-vm"
# erigonvm="/home/martin/workspace/erigon-evm"

### Geth

if [[ -n "$evm" ]]; then
    echo "geth"
    cd ./cases
    for i in *.json; do
        $evm --json --nomemory --noreturndata statetest $i \
         2>../traces/$i.geth.stderr.txt \
         1>../traces/$i.geth.stdout.txt
    done
    cd ..
fi


### Nethermind

if [[ -n "$nethtest" ]]; then
    echo "nethermind"
    cd ./cases
    for i in *.json; do
        $nethtest -m --trace --input $1 \
         2>../traces/$i.nethermind.stderr.txt \
         1>../traces/$i.nethermind.stdout.txt
    done
fi


### Besu

if [[ -n "$besuvm" ]]; then
    echo "besu"
    cd ./cases
    for i in *.json; do
        $besuvm --json --nomemory state-test $i \
          2>../traces/$i.besu.stderr.txt \
          1>../traces/$i.besu.stdout.txt
    done
fi

### Erigon

if [[ -n "$erigonvm" ]]; then
    echo "erigon"
    cd ./cases
    for i in *.json; do
        $erigonvm --json --nomemory --noreturndata statetest $i \
          2>../traces/$i.erigon.stderr.txt \
          1>../traces/$i.erigon.stdout.txt
    done
fi
```
