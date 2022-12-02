## Command to generate these

# First set the binaries to use
# evm="/home/martin/workspace/evm"
# nethtest="/home/martin/workspace/nethtest"
# besuvm="/home/martin/workspace/besu-vm"
# erigonvm="/home/martin/workspace/erigon-evm"

### Geth

if [[ -n "$geth" ]]; then
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