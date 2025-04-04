This is a dockerfile containing various L2 VMs, plus go-evmlab itself.
The evm binaries are available as ENV vars:

- "regular" vms
  - `$GETH_BIN`=/gethvm - The regular normal geth.
  - `$OP_BIN`=/opvm - optimism fork of geth. Supports up to Prague
  - `$ARB_BIN`=/arbvm - offchain labs arbitrum fork of geth. Supports up to Prague
  - `$CELO_BIN`=/celovm - celo fork of op-geth. Supports up to London (the codebase supports Prague ,in theory, but the Celo network has not activated Prague)
    - Berlin + London: Celo Mainnet "Espresso" HF, March 8 2022 
    - Ethereum alignment (?): Celo Mainnet "Gingerbread" HF, Sep 26 2023
      - https://forum.celo.org/t/introducing-celo-s-gingerbread-hard-fork-join-for-q-a-on-june-21/5918

- "non-regular" vms, where the consensus rules are changed.
  - `$BORGETH_BIN`=/borgovm - A bor-flavoured geth, which is same as regular geth but with a few tiny tweaks to make it execute statetests similarly.
  - `$BOR_BIN`=/borvm - the vm from https://github.com/maticnetwork/bor
    - This is based on an older geth, this needs it's own shim: `--bor` or `--borbatch`. (All the other use `--geth`/`--gethbatch`)



## Generating reference output

TOOD

## Checkslow

## Run a test against all clients

```
docker run -it -v /home/user/workspace/tests/fuzztmp:/fuzztmp --entrypoint bash holiman/omnifuzz

$ runtest $FUZZ_CLIENTS /fuzztmp/
```


## Fuzzing

If you want to do fuzzing, you should ensure that the directory where tests are saved is mounted outside the docker container

```
docker run -it -v /home/user/fuzzing:/fuzztmp

$ generic-fuzzer --outdir=/fuzztmp  --fork=Cancun $VMS_REGULAR

# or

$ generic-fuzzer --outdir=/fuzztmp  --fork=London $VMS_REGULAR
```
or
```
docker run -it -v /home/user/fuzzing:/fuzztmp

$ generic-fuzzer --outdir=/fuzztmp  --fork=London $VMS_BOR
```
or
```
docker run -it -v /home/user/fuzzing:/fuzztmp

$ generic-fuzzer --outdir=/fuzztmp  --fork=Prague $VMS_BSC
```

