package evms

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"sync"
)

// GethEVM is s Evm-interface wrapper around the `evm` binary, based on go-ethereum.
type GethEVM struct {
	path string
	wg   sync.WaitGroup
}

func NewGethEVM(path string) *GethEVM {
	return &GethEVM{
		path: path,
	}
}

// StartStateTest implements the Evm interface
func (vm *GethEVM) StartStateTest(path string) (chan *OutputItem, error) {
	var (
		stderr io.ReadCloser
		err    error
	)
	cmd := exec.Command(vm.path, "--json", "--nomemory", "statetest", path)
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

func (vm *GethEVM) Close() {
	vm.wg.Wait()
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (vm *GethEVM) feed(input io.Reader, ch chan(*OutputItem)){
	defer close(ch)
	defer vm.wg.Done()
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		// Calling bytes means that bytes in 'l' will be overwritten
		// in the next loop. Fine for now though, we immediately marshal it
		data := scanner.Bytes()
		var elem OutputItem
		json.Unmarshal(data, &elem)

		// Geth outputs gasUsed and time to stderr, we ignore it
		if _, present := elem["time"]; present {
			continue
		}
		ch <- &elem
	}
}

