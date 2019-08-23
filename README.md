# Go evmlab

This project is based on [EVMlab](https://github.com/ethereum/evmlab), which was
written in Python. EVMlab featured a minimal "compiler", along with some tooling
to view traces in a UI, and execute scripts against EVMs (parity and geth). 

This is the golang version of that same project, rewritten in go to be more stable 
and nice to use. 

## Status

So far, it only contains a minimal "compiler", and not the complete feature set
of the Python version. 

## Examples

See [examples/calltree](the calltree example) to get an idea of how to use this
thing, along with an [analysis](examples/calltree/README.md) done using 
this framework. 