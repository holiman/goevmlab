## Command to generate these

### Geth

    evm="/home/user/go/src/github.com/ethereum/go-ethereum/build/bin/evm"
    for i in *.json; do
        $evm --json --nomemory --noreturndata statetest $i 2>/dev/null 1>$i.geth.stdout.txt
        $evm --json --nomemory --noreturndata statetest $i 1>/dev/null 2>$i.geth.stderr.txt
    done


### Nethermind

    for i in *.json; do
	$nethtest -m --trace --input $1  2>/dev/null 1>$i.nethermind.stdout.txt
        $nethtest -m --trace --input $i  1>/dev/null 2>$i.nethermind.stderr.txt
    done


### Besu

    for i in *.json; do
        $besuvm --json --nomemory state-test $i 2>/dev/null 1>$i.besu.stdout.txt
        $besuvm --json --nomemory state-test $i 1>/dev/null 2>$i.besu.stderr.txt
    done

### Erigon

    erigonvm="/home/user/go/src/github.com/ledgerwatch/erigon/build/bin/evm"
    for i in *.json; do
        $erigonvm --json --nomemory --noreturndata statetest $i 2>/dev/null 1>$i.erigon.stdout.txt
        $erigonvm --json --nomemory --noreturndata statetest $i 1>/dev/null 2>$i.erigon.stderr.txt
    done
