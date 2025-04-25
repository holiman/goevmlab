// Copyright 2025 Martin Holst Swende
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

package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

func readCorpusFile(f *testing.F, path string) error {
	corpus, err := os.Open(path)
	if err != nil {
		return err
	}
	defer corpus.Close()
	f.Logf("Reading seed data from %v", path)
	scanner := bufio.NewScanner(corpus)
	scanner.Buffer(make([]byte, 1024), 10*1024*1024)
	toRemove := regexp.MustCompile(`[^0-9A-Za-z]`)
	for scanner.Scan() {
		l := scanner.Text()
		if strings.HasPrefix(l, "#") {
			continue
		}
		sanitized := toRemove.ReplaceAllString(l, "")
		input := common.FromHex(sanitized)
		if len(input) > 0 {
			f.Add(input)
		}
	}
	return scanner.Err()
}

func seedFuzzer(f *testing.F) {
	//if err := readCorpusFile(f, "testdata/eof_corpus_0.txt"); err != nil {
	//	f.Fatal(err)
	//}
	//if err := readCorpusFile(f, "testdata/eof_corpus_1.txt"); err != nil {
	//	f.Fatal(err)
	//}
}

func FuzzEofParsing(f *testing.F) {
	// Seed with corpus from execution-spec-tests
	seedFuzzer(f)
	for i := 0; ; i++ {
		fname := fmt.Sprintf("testdata/eof_corpus_%d.txt", i)
		corpus, err := os.Open(fname)
		if err != nil {
			break
		}
		f.Logf("Reading seed data from %v", fname)
		scanner := bufio.NewScanner(corpus)
		scanner.Buffer(make([]byte, 1024), 10*1024*1024)
		for scanner.Scan() {
			s := scanner.Text()
			if len(s) >= 2 && strings.HasPrefix(s, "0x") {
				s = s[2:]
			}
			b, err := hex.DecodeString(s)
			if err != nil {
				panic(err) // rotten corpus
			}
			f.Add(b)
		}
		corpus.Close()
		if err := scanner.Err(); err != nil {
			panic(err) // rotten corpus
		}
	}
	// And do the fuzzing
	f.Fuzz(func(t *testing.T, data []byte) {
		var (
			jt = vm.NewEOFInstructionSetForTesting()
			c  vm.Container
		)
		cpy := common.CopyBytes(data)
		if err := c.UnmarshalBinary(data, true); err == nil {
			c.ValidateCode(&jt, true)
			if have := c.MarshalBinary(); !bytes.Equal(have, data) {
				t.Fatal("Unmarshal-> Marshal failure!")
			}
		}
		if err := c.UnmarshalBinary(data, false); err == nil {
			c.ValidateCode(&jt, false)
			if have := c.MarshalBinary(); !bytes.Equal(have, data) {
				t.Fatal("Unmarshal-> Marshal failure!")
			}
		}
		if !bytes.Equal(cpy, data) {
			panic("data modified during unmarshalling")
		}
	})
}

// FuzzBinaries fuzzes the set of binaries in the file 'binaries.txt', which
// are assumed to be eoparse-type binaries.
func FuzzBinaries(f *testing.F) {
	var bins []string
	if binaries, err := os.ReadFile("binaries.txt"); err != nil {
		f.Fatal(err)
	} else {
		for _, x := range strings.Split(strings.TrimSpace(string(binaries)), "\n") {
			x = strings.TrimSpace(x)
			if len(x) > 0 && !strings.HasPrefix(x, "#") {
				bins = append(bins, x)
			}
		}
	}
	if len(bins) < 2 {
		fmt.Printf("Usage: comparer parser1,parser2,... \n")
		fmt.Printf("Pipe input to process")
		f.Fatal("error")
	}
	var inputs = make(chan string)
	var outputs = make(chan string)
	go func() {
		err := doCompare(bins, inputs, outputs)
		f.Log("Done")
		if err != nil {
			f.Fatalf("exec error: %v", err)
		}
	}()
	time.Sleep(3 * time.Second)
	seedFuzzer(f)

	f.Fuzz(func(t *testing.T, data []byte) {
		_ = testUnmarshal(data) // This is for coverage guidance
		inputs <- fmt.Sprintf("%#x", data)
		errStr := <-outputs
		if len(errStr) != 0 {
			t.Fatal(errStr)
		}
	})
}

// FuzzFifo fuzzes the set of for eof parsing, and also sends the fuzzing-inputs
// to a fifo
func FuzzFifo(f *testing.F) {
	// Try to dial socket
	socket := os.Getenv("EOF_FUZZ_PIPE")
	if socket == "" {
		f.Skip("env EOF_FUZZ_PIPE not set")
		return
	}
	c, err := net.Dial("unix", socket)
	if err != nil {
		f.Fatal(err)
	}
	seedFuzzer(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = testUnmarshal(data) // This is for coverage guidance
		if _, err := c.Write([]byte(fmt.Sprintf("%x\n", data))); err != nil {
			t.Fatal(err)
		}
	})
}

var jt, _ = vm.LookupInstructionSet(params.Rules{
	ChainID:          nil,
	IsHomestead:      true,
	IsEIP150:         true,
	IsEIP155:         true,
	IsEIP158:         true,
	IsEIP2929:        true,
	IsEIP4762:        false,
	IsByzantium:      true,
	IsConstantinople: true,
	IsPetersburg:     true,
	IsIstanbul:       true,
	IsBerlin:         true,
	IsLondon:         true,
	IsMerge:          true,
	IsShanghai:       true,
	IsCancun:         true,
	IsPrague:         true,
	IsOsaka:          true,
	IsVerkle:         false,
})

func testUnmarshal(blob []byte) string {
	var c vm.Container
	if err := c.UnmarshalBinary(blob, false); err != nil {
		return fmt.Sprintf("err: %v\n", err)
	}
	if err := c.ValidateCode(&jt, false); err != nil {
		return fmt.Sprintf("err: %v\n", err)
	}
	return fmt.Sprintf("OK")
}
