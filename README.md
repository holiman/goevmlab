# Go evmlab

This project is inspired by [EVMlab](https://github.com/ethereum/evmlab), which was
written in Python. EVMlab featured a minimal "compiler", along with some tooling
to view traces in a UI, and execute scripts against EVMs (parity and geth). 

This is a golang rewrite from scratch of that same project, this time in go-lang to be more stable 
and nice to use. 

## Status

So far, it only contains 

- a minimal "compiler", which is _not_ a proper compiler, but it can be used to create EVM bytecode and do some basic higher level operations. 
- Some infrastructure for writing EVM-based fuzzers

## Examples

See [examples/calltree](the calltree example) to get an idea of how to use this
thing, along with an [analysis](examples/calltree/README.md) done using 
this framework. 

# Utilities

There are a few different utilities in this repo, based on this library. 


## Tracediff

Tracediff allows you to load evm (jsonl) files and find differences. 

![tracediff](docs/tracediff.png)
![tracediff](docs/tracediff2.png)

## Traceview

Traceview allows you to analyse an evm trace file

![traceview](docs/traceview.png)


## Trophy list

Bugs found via the fuzzer or via tests generated using go-evmlab in general

- Nethermind erroneous gas in BLS-multiexp: https://github.com/NethermindEth/nethermind/issues/2208
- Geth error calculating BLS gas: https://github.com/ethereum/go-ethereum/pull/21253

