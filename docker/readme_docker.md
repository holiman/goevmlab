This is a dockerfile containing all VMs, plus go-evmlab itself.
The evm binaries are available as ENV vars:

- `$GETH_BIN`=/gethvm
- `$ERIG_BIN`=/erigon_vm
- `$NIMB_BIN`=/nimbvm
- `$EVMO_BIN`=/evmone
- `$RETH_BIN`=/revme
- `$NETH_BIN`=/neth/nethtest
- `$BESU_BIN`=/evmtool/bin/evm
- `$EELS_BIN`=/ethereum-spec-evm


## Fuzzing

If you want to do fuzzing, you should ensure that the directory where tests are
saved is mounted outside the docker container

```
docker run -it -v /home/user/fuzzing:/fuzztmp

$ /generic-fuzzer --outdir=/fuzztmp \
    --gethbatch=$GETH_BIN \
    --nethbatch=$NETH_BIN \
    --nimbus=$NIMB_BIN \
    --revme=$RETH_BIN \
    --erigonbatch=$ERIG_BIN \
    --besubatch=$BESU_BIN \
    --evmone=$EVMO_BIN \
    --eelsbatch=$EELS_BIN \
    --fork=Cancun
```

## Generating reference output

Mount the reference tests, and execute the `run.sh` to create them:
```
docker run -it -v /home/user/workspace/goevmlab/evms/testdata/:/testdata \
    --entrypoint /bin/bash
$ bash run.sh

```
## Checkslow

```
docker run -it -v /home/user/workspace/goevmlab/trophies/2024-02-20_slow_tests/fuzztmp:/fuzztmp --entrypoint bash holiman/omnifuzz

$ /checkslow  --nethbatch=$NETH_BIN --evmone=$EVMO_BIN --verbosity -4  /fuzztmp/
```
