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

## Sanity-check

To check if the clients seem to be behaving correctly, you can do
```
$ evms.test -test.run TestVMsFromEnv -test.v
```

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

## Blocktests

The fuzzers can also generate and execute _blockchain tests_ instead of state tests, in order to
test block-level behaviour such as the block access lists (EIP-7928) of Amsterdam. Only some
clients can execute blocktests; the env var `$FUZZ_BLOCK_CLIENTS` lists them.

```
$ generic-fuzzer --outdir=/fuzztmp --fork=Amsterdam --blocktest=both $FUZZ_BLOCK_CLIENTS
```

The `--blocktest` mode selects the fixture flavour(s):

- `construct`: `blockchain_test` fixtures. The blocks are RLP-encoded, and the clients construct the
  block access list themselves and verify it against the header. Executed with tracing.
- `consume`: `blockchain_test_engine` fixtures. The blocks are engine payloads which include the block
  access list, so the clients take the BAL-driven parallel execution path. Executed without tracing,
  comparing only the final state root and verdict.
- `both`: each generated test is written in both flavours.

Existing blocktests (e.g. from execution-spec-tests) can be executed with `runtest`, which detects
the test kind from the file content:

```
$ runtest $FUZZ_BLOCK_CLIENTS '/fuzztmp/*.json'
```
