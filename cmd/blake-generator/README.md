## Blake generator

This utility generates fuzzed blake `f` compression function statetests. 

Example to generate `100` blake tests into `/tmp`:
````
$ ./blake-generator --count 100 generate
INFO [12-01|09:39:00.676] Generating tests 
INFO [12-01|09:39:00.676] Generating tests                         location=/tmp prefix= fork=London limit=100

````
```
``$ ls /tmp | grep bla | head
blaketest-0.json
blaketest-1.json
blaketest-10.json
```s
