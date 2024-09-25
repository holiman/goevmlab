This is a dockerfile containing all VMs, plus go-evmlab itself.
The evm binaries are available as ENV vars:

- `$GETH_BIN`=/gethvm
- `$ERIG_BIN`=/erigon_vm
- `$NIMB_BIN`=/nimbvm
- `$EVMO_BIN`=/evmone
- `$RETH_BIN`=/revme
- `$NETH_BIN`=/neth/nethtest
- `$BESU_BIN`=/evmtool/bin/evmtool
- `$EELS_BIN`=/ethereum-spec-evm

There's also an env var $FUZZ_CLIENTS which provides the arguments if you want to do fuzzing with all clients.

## Generating reference output

Mount the reference tests, and execute the `run.sh` to create them:
```
docker run -it -v /home/user/workspace/goevmlab/evms/testdata/:/testdata  holiman/omnifuzz
$ cd /testdata
$ bash run.sh

```
## Checkslow

```
docker run -it -v /home/user/workspace/goevmlab/trophies/2024-02-20_slow_tests/fuzztmp:/fuzztmp --entrypoint bash holiman/omnifuzz

$ checkslow  --nethbatch=$NETH_BIN --evmone=$EVMO_BIN --verbosity -4  /fuzztmp/
```

## Run a test against all clients

```
docker run -it -v /home/user/workspace/tests/fuzztmp:/fuzztmp --entrypoint bash holiman/omnifuzz

$ runtest $FUZZ_CLIENTS /fuzztmp/
```


## Fuzzing

If you want to do fuzzing, you should ensure that the directory where tests are saved is mounted outside the docker container

```
docker run -it -v /home/user/fuzzing:/fuzztmp

$ generic-fuzzer --outdir=/fuzztmp  --fork=Cancun $FUZZ_CLIENTS
```
