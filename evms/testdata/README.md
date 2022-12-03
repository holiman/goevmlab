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

# evm="/home/martin/workspace/evm"
# nethtest="/home/martin/workspace/nethtest"
# besuvm="/home/martin/workspace/besu-vm"
# erigonvm="/home/martin/workspace/erigon-evm"

### Geth

if [[ -n "geth" ]]; then
    echo "geth"
    for i in ./cases/*.json; do
        $evm --json --nomemory --noreturndata statetest $i 2>/dev/null 1>./traces/$i.geth.stdout.txt
        $evm --json --nomemory --noreturndata statetest $i 1>/dev/null 2>./traces/$i.geth.stderr.txt
    done
fi


### Nethermind

if [[ -n "$nethtest" ]]; then
    echo "nethermind"
    for i in ./cases/*.json; do
        $nethtest -m --trace --input $1  2>/dev/null 1>./traces/$i.nethermind.stdout.txt
        $nethtest -m --trace --input $i  1>/dev/null 2>./traces/$i.nethermind.stderr.txt
    done
fi


### Besu

if [[ -n "$besuvm" ]]; then
    echo "besu"
    for i in ./cases/*.json; do
        $besuvm --json --nomemory state-test $i 2>/dev/null 1>./traces/$i.besu.stdout.txt
        $besuvm --json --nomemory state-test $i 1>/dev/null 2>./traces/$i.besu.stderr.txt
    done
fi

### Erigon

if [[ -n "$erigonvm" ]]; then
    echo "erigon"
    for i in ./cases/*.json; do
        $erigonvm --json --nomemory --noreturndata statetest $i 2>/dev/null 1>./traces/$i.erigon.stdout.txt
        $erigonvm --json --nomemory --noreturndata statetest $i 1>/dev/null 2>./traces/$i.erigon.stderr.txt
    done
fi
```
