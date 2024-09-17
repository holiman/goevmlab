#! /usr/bin/env bash

echo "benchmarking g2msm"
for i in {1..32}
do
	echo -n "	$i pairs: "
	PRECOMPILE=g2msm INPUT_COUNT=$i ./benchmark.sh
done

for i in 64 128 256 512 2048 4096
do
	echo -n "	$i pairs: "
	PRECOMPILE=g2msm INPUT_COUNT=$i ./benchmark.sh
done
