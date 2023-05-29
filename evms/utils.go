package evms

const (
	// Nethermind does not support refundcounter
	ClearRefunds = true

	// Nethermind reports the change in memory on step earlier than others. E.g.
	// MSTORE shows the _new_ memory, besu/geth shows the old memory until the next op.
	// This could be handled differently, e.g. clearing only on mem-expanding ops.
	ClearMemSize = true

	// Nethermind reports the change in memory on step earlier than others. E.g.
	// MSTORE shows the _new_ memory, besu/geth shows the old memory until the next op.
	// Unfortunately, nethermind also "forgets" the memsize when an error occurs, reporting
	// memsize zero (see testdata/traces/stackUnderflow_nonzeroMem.json). So we use
	// ClearMemSize instead
	ClearMemSizeOnExpand = true

	// Nethermind is missing returnData
	ClearReturndata = true

	// Besu sometimes reports GasCost of 0x7fffffffffffffff, along with ,"error":"Out of gas"
	ClearGascost = true

	// Evmone does not report depth
	ClearDepth = true
)
