This is a dockerfile containing all VMs, plus go-evmlab itself.
The evm binaries are available as ENV vars:

- `$ERIG_BIN`=/erigon_vm
- `$NIMB_BIN`=/nimbvm
- `$EVMO_BIN`=/evmone
- `$RETH_BIN`=/revme
- `$NETH_BIN`=/neth/nethtest
- `$BESU_BIN`=/evmtool/bin/evm

## Fuzzing

If you want to do fuzzing, you should ensure that the directory where tests are
saved is mounted outside the docker container

```
docker run -it -v /home/user/fuzzing:/fuzztmp --entrypoint /generic-fuzzer --outdir=/fuzztmp --nethbatch=/nethtest --nimbus=/nimbvm --revme=/revme --erigonbatch=/erigon_vm --besubatch=/besu-vm --evmone=/evmone --fork=Cancun
```

## Generating reference output


## Checkslow

```

docker run -it -v /home/user/workspace/goevmlab/trophies/2024-02-20_slow_tests/fuzztmp:/fuzztmp --entrypoint bash holiman/omnifuzz

$ /checkslow  --nethbatch=$NETH_BIN --evmone=$EVMO_BIN --verbosity -4  /fuzztmp/
```
