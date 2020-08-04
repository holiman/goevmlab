## Jumpdest analysis

This is a test for restricted subroutines, where we make a more thorough 
analysis during jumpdest analysis.

A contract can do roughly ~300 CREATE operations on 10M gas, 
(each CREATE costs 32K). 

I made a test where we do an analysis of a contract which is 0xc000 bytes large, 
mainly consisting of PUSH1 01, but with an initial jump and stop to trigger the analysis.

I ran it with 100M (obs: `100M`, not `10M`) gas, and benchmarked it. 

With only old-style analysis, it blew through 100M gas in `40-59ms`:
```
      27          43398488 ns/op 25509711 B/op     22689 allocs/op
gas used: 10000000, 230.422774 Mgas/s

          22          53288897 ns/op 25716607 B/op     22871 allocs/op
gas used: 10000000, 187.656352 Mgas/s
```
With _both_ shadow-array and old analysis made, it blew through 100M gas in `~65ms` :

```
      18          67303719 ns/op 44805790 B/op     23490 allocs/op
gas used: 10000000, 148.580200 Mgas/s

      16          64121205 ns/op 45101226 B/op     23648 allocs/op
gas used: 10000000, 155.954651 Mgas/s

```
