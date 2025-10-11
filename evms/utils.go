package evms

import (
	"bytes"
	"errors"
	"os/exec"
)

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
)

// StdErrOutput runs the command and returns its standard error.
func StdErrOutput(c *exec.Cmd) ([]byte, error) {
	if c.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b bytes.Buffer
	c.Stderr = &b
	err := c.Run()
	return b.Bytes(), err
}
