// Copyright 2019 Martin Holst Swende
// This file is part of the goevmlab library.
//
// The library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the goevmlab library. If not, see <http://www.gnu.org/licenses/>.

package evms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/core/vm"
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
func (evm *GethEVM) StartStateTest(path string) (chan *vm.StructLog, error) {
	var (
		stderr io.ReadCloser
		err    error
	)
	cmd := exec.Command(evm.path, "--json", "--nomemory", "statetest", path)
	if stderr, err = cmd.StderrPipe(); err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	ch := make(chan *vm.StructLog)
	evm.wg.Add(1)
	go func(){
		evm.feed(stderr, ch)
		cmd.Wait()
	}()
	return ch, nil

}

func (vm *GethEVM) Close() {
	vm.wg.Wait()
}

// feed reads from the reader, does some geth-specific filtering and
// outputs items onto the channel
func (evm *GethEVM) feed(input io.Reader, opsCh chan (*vm.StructLog)) {
	defer close(opsCh)
	defer evm.wg.Done()
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		// Calling bytes means that bytes in 'l' will be overwritten
		// in the next loop. Fine for now though, we immediately marshal it
		data := scanner.Bytes()
		var elem vm.StructLog
		err := json.Unmarshal(data, &elem)
		if err != nil {
			fmt.Printf("geth err: %v, line\n\t%v\n", err, string(data))
			continue
		}
		//fmt.Printf("geth: %v\n", string(data))
		// If the output cannot be marshalled, all fields will be blanks.
		// We can detect that through 'depth', which should never be less than 1
		// for any actual opcode
		if elem.Depth == 0 {
			/*  Most likely one of these:
			{"output":"","gasUsed":"0x2d1cc4","time":233624,"error":"gas uint64 overflow"}

			{"stateRoot": "a2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134"}

			*/
			// For now, just ignore these
			continue
		}
		// When geth encounters end of code, it continues anyway, on a 'virtual' STOP.
		// In order to handle that, we need to drop all STOP opcodes.
		if elem.Op == 0x0 {
			continue
		}

		opsCh <- &elem
	}
}

func has0xPrefix(input string) bool {
	return len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X')
}

// addPrefix prefixes hexStr with 0x, if needed
func addPrefix(hexStr string) string {
	if len(hexStr) >= 2 && hexStr[1] != 'x' {
		enc := make([]byte, len(hexStr)+2)
		copy(enc, "0x")
		copy(enc[2:], hexStr)
		return string(enc)
	}
	return hexStr
}
