## CREATE2 initcode-size check bug

This example builds a testcase for https://github.com/ethereum/execution-specs/issues/915


This program does a CREATE2 which fails. The CREATE2 can fail for two reasons:

1. it is way to large initcode. This failure exits the current scope.
2. it tries to use too large endowment. This failure fails the create2-op, butdoes not exit the current scope.

The consensus-correct way to fail is 1).

Finished testcase: [./create2_test.json]