## Command to generate these

#### Geth 

	../../evm --json --nomemory statetest ./statetest1.json 1> statetest1_geth_stdout.jsonl
	../../evm --json --nomemory statetest ./statetest1.json 2> statetest1_geth_sterr.jsonl


### Parity


	../../parity-evm --std-json state-test ./statetest1.json 1> statetest1_parity_stdout.jsonl
	../../parity-evm --std-json state-test ./statetest1.json 2> statetest1_parity_stderr.jsonl

### Aleth / Testeth


	../../testeth -t GeneralStateTests --  --testfile ./statetest1.json --jsontrace '{"disableMemory": true}' 1> statetest1_testeth_stdout.jsonl
	../../testeth -t GeneralStateTests --  --testfile ./statetest1.json --jsontrace '{"disableMemory": true}' 2> statetest1_testeth_stderr.jsonl


### Nethermind

	../../nethtest -m  --input statetest1.json --trace 1> statetest1_nethermind_stdout.jsonl
	../../nethtest -m --input statetest1.json --trace 2> statetest1_nethermind_stderr.jsonl

