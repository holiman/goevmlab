#! /usr/bin/env bash

if [ -z ${INPUT_COUNT+x} ];then
        INPUT_COUNT=1
fi

if [ "$INPUT_COUNT" -gt "32" ]; then
        bench_stats=$($GETH_EVM --bench statetest benchmarks/bench-$PRECOMPILE-$INPUT_COUNT.json  --subtest.fork "Prague" --subtest.index 0 --subtest.name bench-$PRECOMPILE-$INPUT_COUNT 2>&1 | python3 calc_stats.py) 
        noop_stats=$($GETH_EVM --bench statetest benchmarks/noop-$PRECOMPILE-$INPUT_COUNT.json  --subtest.fork "Prague" --subtest.index 0 --subtest.name noop-$PRECOMPILE-$INPUT_COUNT 2>&1 | python3 calc_stats.py)
	iter=100
else
        bench_stats=$($GETH_EVM --bench statetest benchmarks/bench-$PRECOMPILE-$INPUT_COUNT.json  --subtest.fork "Prague" --subtest.index 0 --subtest.name bench-$PRECOMPILE-$INPUT_COUNT 2>&1 | python3 calc_stats.py) 
        noop_stats=$($GETH_EVM --bench statetest benchmarks/noop-$PRECOMPILE-$INPUT_COUNT.json  --subtest.fork "Prague" --subtest.index 0 --subtest.name noop-$PRECOMPILE-$INPUT_COUNT 2>&1 | python3 calc_stats.py)
	iter=2850
fi

echo "$noop_stats,$bench_stats,$iter" | python3 measure_perf.py
