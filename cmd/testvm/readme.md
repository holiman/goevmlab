### TestVM

This is a small utility to aid testing. It can be used to simulate the execution of clients. 
It behaves somewhat like a geth-client. 

E.g. 
```
generic-fuzzer --geth ./testvm-01 --geth ./testvm-02 --geth ./testvm-03
```
Where the three testvm's are compiled with different speeds (so one takes 500ms and another 200ms), 
or where one of them triggers a consensus issue. 

The idea is to use this to help test tools like `generic-fuzzer`, `runtest` or `minimizer`. 