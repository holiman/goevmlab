import sys

input = sys.stdin.read()

stats_noop = input.split(',')[0]
noop_exec_time = int(stats_noop.split("-")[0])
noop_gas_used = int(stats_noop.split("-")[1])

stats_bench = input.split(',')[1]
bench_exec_time = int(stats_bench.split("-")[0])
bench_gas_used = int(stats_bench.split("-")[1])

iter_count = int(input.split(',')[2])

#print("bench gas used: {}, bench time: {}".format(bench_gas_used, bench_exec_time))
#print("noop gas used: {}, noop time: {}".format(noop_gas_used, noop_exec_time))

# calc throughput after subtracting noop gas/time from bench gas/time
corrected_gas_used = bench_gas_used + (15 * iter_count) - noop_gas_used
corrected_exec_time = bench_exec_time + (15 * iter_count) - noop_exec_time

gas_throughput = corrected_gas_used / (corrected_exec_time / 1e9)
gas_throughput /= 1e6
print("{} mgas/sec".format(round(gas_throughput)))
