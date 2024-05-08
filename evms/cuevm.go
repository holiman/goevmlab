// Copyright 2023 Martin Holst Swende
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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"github.com/pkg/errors"
)

type CuEVM struct {
	path string
	name string

	stats *VmStat
}

type account struct {
	Address  string
	Balance  string
	Nonce    string
	CodeHash string `json:"codeHash"`
	Storage  [][]string
}

func (Account account) String() string {
	return fmt.Sprintf("Account{Address: %v, Balance: %v, Nonce: %v, CodeHash: %v, Storage: %v}",
		Account.Address, Account.Balance, Account.Nonce, Account.CodeHash, Account.Storage)
}

type cuevmState struct {
	StateRoot string `json:"stateRoot"`
	Accounts  []account
}

type evmStateRoot struct {
	StateRoot string `json:"stateRoot"`
}

// Remove prefix "0x" and padd with zeros to length `length`
func padWithZeros(s string, length int) string {
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	if len(s) < length {
		return strings.Repeat("0", length-len(s)) + s
	}
	return s
}

func addHexPrefix(s string) string {
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return s
	}
	return "0x" + s
}

func String2Hex(s string, length int) ([]byte, error) {
	s = padWithZeros(s, length)
	data, err := hex.DecodeString(s)
	return data, errors.WithStack(err)
}

func (state *cuevmState) ComputeStateRoot() error {
	if state.StateRoot != "" {
		return nil
	}

	stateTrie := trie.NewEmpty(triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil))

	zero := uint256.NewInt(0)

	for i := range state.Accounts {
		account := state.Accounts[i]
		stateAccount := types.NewEmptyStateAccount()
		nonceBig, err := uint256.FromHex(addHexPrefix(account.Nonce))

		if err != nil {
			return errors.WithStack(err)
		}

		nonce := nonceBig.Uint64()

		balance, err := uint256.FromHex(addHexPrefix(account.Balance))

		if err != nil {
			return errors.WithStack(err)
		}

		// skip empty account
		if nonceBig.Eq(zero) && balance.Eq(zero) && len(account.Storage) == 0 && strings.Compare(strings.ToLower(account.CodeHash), "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470") == 0 {
			continue
		}

		codeHash, err := String2Hex(account.CodeHash, 64)
		if err != nil {
			return errors.WithStack(err)
		}

		storageTrie := trie.NewEmpty(triedb.NewDatabase(rawdb.NewMemoryDatabase(), nil))
		for i := range account.Storage {
			storageKey := account.Storage[i][0]
			storageVal := account.Storage[i][1]

			if storageVal == "0x0" {
				continue
			}

			key, err := uint256.FromHex(addHexPrefix(storageKey))

			if err != nil {
				return errors.WithStack(err)
			}

			value, err := uint256.FromHex(addHexPrefix(storageVal))

			if err != nil {
				return errors.WithStack(err)
			}

			trieKey := crypto.Keccak256(key.PaddedBytes(32))

			trieValue, err := rlp.EncodeToBytes(value)

			if err != nil {
				return errors.WithStack(err)
			}

			storageTrie.Update(trieKey, trieValue)
		}
		root := storageTrie.Hash()

		stateAccount.Nonce = nonce
		stateAccount.Balance = balance

		stateAccount.CodeHash = codeHash
		stateAccount.Root = root

		temp, err := String2Hex(account.Address, 40)
		if err != nil {
			return errors.WithStack(err)
		}

		stateKey := crypto.Keccak256(temp)
		stateVal, err := rlp.EncodeToBytes(stateAccount)

		if err != nil {
			return errors.WithStack(err)
		}

		stateTrie.Update(stateKey, stateVal)
	}

	root := stateTrie.Hash().Hex()
	state.StateRoot = root
	return nil
}

func NewCuEVM(path string, name string) *CuEVM {
	return &CuEVM{
		path:  path,
		name:  name,
		stats: &VmStat{},
	}
}

func (evm *CuEVM) Instance(int) Evm {
	return evm
}

func (evm *CuEVM) Name() string {
	return fmt.Sprintf("cuevm-%s", evm.name)
}

func (evm *CuEVM) GetStateRoot(path string) (root, command string, err error) {
	// todo update this for stateRoot only tests
	cmd := exec.Command(evm.path, "statetest", "--json-outcome", path)
	data, err := StdErrOutput(cmd)

	// If cuevm exits with 1 on stateroot errors, uncomment to ignore:
	//if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
	//	err = nil
	//}
	if err != nil {
		return "", cmd.String(), err
	}

	root, err = evm.ParseStateRoot(data)
	if err != nil {
		log.Error("Failed to find stateroot", "vm", evm.Name(), "cmd", cmd.String())
		return "", cmd.String(), err
	}
	return root, cmd.String(), nil
}

func (evm *CuEVM) ParseStateRoot(data []byte) (root string, err error) {
	pattern := []byte(`"stateRoot":"`)
	idx := bytes.Index(data, pattern)
	start := idx + len(pattern)
	end := start + 32*2 + 2
	if idx == -1 || end >= len(data) {
		return "", fmt.Errorf("%v: no stateroot found", evm.Name())
	}
	return string(data[start:end]), nil
}

func (evm *CuEVM) RunStateTest(path string, out io.Writer, speedTest bool) (*tracingResult, error) {
	var (
		t0     = time.Now()
		stderr io.ReadCloser
		err    error
		cmd    *exec.Cmd
	)
	cmd = exec.Command(evm.path, "--input", path, "--output", "/dev/null")

	if stderr, err = cmd.StderrPipe(); err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}

	evm.Copy(out, stderr) // stderr is used for traces as json
	err = cmd.Wait()
	duration, slow := evm.stats.TraceDone(t0)

	// If cuevm exits with 1 on stateroot errors, uncomment to ignore:
	//if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
	//	err = nil
	//}

	return &tracingResult{
		Slow:     slow,
		ExecTime: duration,
		Cmd:      cmd.String(),
	}, err
}

func (vm *CuEVM) Close() {
}

func (evm *CuEVM) Copy(out io.Writer, input io.Reader) {
	buf := bufferPool.Get().([]byte)
	//lint:ignore SA6002: argument should be pointer-like to avoid allocations.
	defer bufferPool.Put(buf)
	var cuevmState cuevmState
	scanner := bufio.NewScanner(input)
	scanner.Buffer(buf, 32*1024*1024)

	for scanner.Scan() {
		data := scanner.Bytes()

		if bytes.Contains(data, []byte("accounts")) {
			if cuevmState.Accounts == nil || len(cuevmState.Accounts) == 0 {
				if err := json.Unmarshal(data, &cuevmState); err != nil {
					fmt.Printf("Error unmarshalling stateRoot: %v\n", err)
					continue
				}

				if err := cuevmState.ComputeStateRoot(); err != nil {
					fmt.Printf("Error computing stateRoot: %+v\n", err)
					continue
				}
			}
		}
		var elem opLog
		if err := json.Unmarshal(data, &elem); err != nil {
			fmt.Printf("cuevm err: %v, line\n\t%v\n", err, string(data))
			continue
		}

		// Drop all STOP opcodes as geth does
		if elem.Op == 0x0 {
			continue
		}
		jsondata := CustomMarshal(&elem)
		if _, err := out.Write(append(jsondata, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
	}
	stateRoot := evmStateRoot{StateRoot: cuevmState.StateRoot}
	root, _ := json.Marshal(stateRoot)
	if _, err := out.Write(append(root, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to output: %v\n", err)
		return
	}
}

func (evm *CuEVM) Stats() []any {
	return evm.stats.Stats()
}
