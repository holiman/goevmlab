package evms

import (
	"bufio"
	"encoding/json"
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
func (vm *ParityVM) StartStateTest(path string) (chan *OutputItem, error) {
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
	ch := make(chan *OutputItem)
	vm.wg.Add(1)
	go vm.feed(stderr, ch)
	return ch, nil

}

func (vm *ParityVM) Close() {
	vm.wg.Wait()
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (vm *ParityVM) feed(input io.Reader, ch chan (*OutputItem)) {
	defer close(ch)
	defer vm.wg.Done()
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		// Calling bytes means that bytes in 'l' will be overwritten
		// in the next loop. Fine for now though, we immediately marshal it
		data := scanner.Bytes()
		var elem OutputItem
		json.Unmarshal(data, &elem)
		ch <- &elem
	}
}
