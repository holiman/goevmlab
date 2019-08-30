package evms

import (
	"bufio"
	"encoding/json"
	"fmt"
	vm2 "github.com/ethereum/go-ethereum/core/vm"
	"io"
	"os/exec"
	"sync"
)

type ParityVM struct {
	path string
	wg   sync.WaitGroup
}

func NewParityVM(path string) *ParityVM {
	return &ParityVM{
		path: path,
	}
}

// StartStateTest implements the Evm interface
func (vm *ParityVM) StartStateTest(path string) (chan *vm2.StructLog, error) {
	var (
		stderr io.ReadCloser
		err    error
	)
	cmd := exec.Command(vm.path, "--std-json", "state-test", path)
	if stderr, err = cmd.StderrPipe(); err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	ch := make(chan *vm2.StructLog)
	vm.wg.Add(1)
	go vm.feed(stderr, ch)
	return ch, nil

}

func (vm *ParityVM) Close() {
	vm.wg.Wait()
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (vm *ParityVM) feed(input io.Reader, opsCh chan (*vm2.StructLog)) {
	defer close(opsCh)
	defer vm.wg.Done()
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		// Calling bytes means that bytes in 'l' will be overwritten
		// in the next loop. Fine for now though, we immediately marshal it
		data := scanner.Bytes()
		var elem vm2.StructLog
		json.Unmarshal(data, &elem)
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			/*  Most likely one of these:
			{"stateRoot":"0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134"}
			*/
			fmt.Printf("parity non-op, line is:\n\t%v\n", string(data))
			// For now, just ignore these
			continue
		}
		fmt.Printf("parity: %v\n", string(data))
		opsCh <- &elem
	}
}
