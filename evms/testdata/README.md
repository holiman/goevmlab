## Testdata output

This folder contains 
 - statetests, 
 - For each statetest: 
   - output from vms, `stdout` and `stderr` output

The statetests have been chosen because they trigger some quirk in a vm, e.g a statetest may 
trigger a negative refund in besu. If we want to change the besu-shim at some later point, 
when said bug has been fixed, we need to regenerate the outputs and check if the 
tests passes. 

## Command to generate these

The script below, after setting the binaries to use, should recreate the outputs 
in `traces` based on the inputs in `cases`. 

See [`run.sh`](./run.sh). 
