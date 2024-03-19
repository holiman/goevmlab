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
docker run -it -v /home/user/fuzzing:/fuzztmp --entrypoint --outdir=/fuzztmp --nethbatch=/nethtest --nimbus=/nimbvm --revme=/revme --erigon=/erigon_vm --besubatch=/besu-vm --evmone=/evmone --fork=Cancun /fuzztmp
```
