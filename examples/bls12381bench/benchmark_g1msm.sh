#! /usr/bin/env bash

echo "benchmarking g1msm"

for i in {1..32}
do
	echo -n "	$i pairs: "
	PRECOMPILE=g1msm INPUT_COUNT=$i ./benchmark.sh
done

for i in 64 128 256 512 2048 4096
do
	echo -n "	$i pairs: "
	PRECOMPILE=g1msm INPUT_COUNT=$i ./benchmark.sh
done
