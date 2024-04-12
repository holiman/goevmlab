## TSTORE bug #1

This example builds a testcase for https://github.com/ethereum/execution-specs/issues/917

The bug is pretty subtle, it pertains to T-storage inside contract deployments.

1. `a`: Run a `CREATE2`
2. `b`: In the initcode, `SSTORE(slot:1, value:1)`
3. `b`: Return too large bytecode, thus failing the creation
4. `a`: Run the `CREATE2` again
5. `b`: check `SLOAD(slot:1)`. If this is non-zero, we have hit the bug.

The testcase implements one way of doing it.

Finished testcase: [./tstore_test-2.json]

