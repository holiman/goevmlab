#! /usr/bin/env bash

./bls12381bench --precompile g1add evaluate
./bls12381bench --precompile g1mul evaluate

./bls12381bench --precompile g2add evaluate
./bls12381bench --precompile g2mul evaluate

./bls12381bench --precompile mapfp evaluate
./bls12381bench --precompile mapfp2 evaluate

for input_count in {1..8}
do
	./bls12381bench --precompile pairing --input-count $input_count evaluate
done

for input_count in {1..32}
do
	./bls12381bench --precompile g1msm --input-count $input_count evaluate
done

./bls12381bench --precompile g1msm --input-count 64 --iter-count 100 evaluate
./bls12381bench --precompile g1msm --input-count 128 --iter-count 100 evaluate
./bls12381bench --precompile g1msm --input-count 256 --iter-count 100 evaluate
./bls12381bench --precompile g1msm --input-count 512 --iter-count 100 evaluate
./bls12381bench --precompile g1msm --input-count 2048 --iter-count 100 evaluate
./bls12381bench --precompile g1msm --input-count 4096 --iter-count 100 evaluate

for input_count in {1..32}
do
	./bls12381bench --precompile g2msm --input-count $input_count evaluate
done

./bls12381bench --precompile g2msm --input-count 64 --iter-count 100 evaluate
./bls12381bench --precompile g2msm --input-count 128 --iter-count 100 evaluate
./bls12381bench --precompile g2msm --input-count 256 --iter-count 100 evaluate
./bls12381bench --precompile g2msm --input-count 512 --iter-count 100 evaluate
./bls12381bench --precompile g2msm --input-count 2048 --iter-count 100 evaluate
./bls12381bench --precompile g2msm --input-count 4096 --iter-count 100 evaluate
