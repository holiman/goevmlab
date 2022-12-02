## Command to generate these

### Geth

    echo "geth"
    for i in *.json; do
        $evm --json --nomemory --noreturndata statetest $i 2>/dev/null 1>$i.geth.stdout.txt
        $evm --json --nomemory --noreturndata statetest $i 1>/dev/null 2>$i.geth.stderr.txt
    done


### Nethermind

    echo "nethermind"
    for i in *.json; do
	$nethtest -m --trace --input $1  2>/dev/null 1>$i.nethermind.stdout.txt
        $nethtest -m --trace --input $i  1>/dev/null 2>$i.nethermind.stderr.txt
    done


### Besu

    echo "besu"
    for i in *.json; do
        $besuvm --json --nomemory state-test $i 2>/dev/null 1>$i.besu.stdout.txt
        $besuvm --json --nomemory state-test $i 1>/dev/null 2>$i.besu.stderr.txt
    done

### Erigon

    echo "erigon"
    for i in *.json; do
        $erigonvm --json --nomemory --noreturndata statetest $i 2>/dev/null 1>$i.erigon.stdout.txt
        $erigonvm --json --nomemory --noreturndata statetest $i 1>/dev/null 2>$i.erigon.stderr.txt
    done
