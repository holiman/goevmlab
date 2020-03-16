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
	"bytes"
	"io"
	"os"
	"testing"
)

// TestGethVsParityStatetestOutput in this test, we simulate two external
// vms, using printouts from actual evm binaries. The two outputs should
// not produce any differences
func TestVMsOutput(t *testing.T) {

	// Some vms (Parity) require both stdout and stderr (since the stateroot
	// and the actual opcodes go to different outputs).
	type testCase struct {
		vm    Evm
		file1 string
		file2 string
	}
	var cases = []testCase{
		{NewGethEVM(""), "testdata/statetest1_geth_stderr.jsonl", ""},
		{NewParityVM(""), "testdata/statetest1_parity_stderr.jsonl", "testdata/statetest1_parity_stdout.jsonl"},
		{NewNethermindVM(""), "testdata/statetest1_nethermind_stderr.jsonl", ""},
		{NewAlethVM(""), "testdata/statetest1_testeth_stdout.jsonl", ""},
	}
	var readers []io.Reader
	var vms []Evm
	for _, vm := range cases {
		parsedOutput := bytes.NewBuffer(nil)
		for _, f := range []string{vm.file1, vm.file2} {
			if len(f) == 0 {
				break
			}
			rawOutput, err := os.Open(f)
			if err != nil {
				t.Fatal(err)
			}
			defer rawOutput.Close()
			vm.vm.Copy(parsedOutput, rawOutput)
		}
		readers = append(readers, bytes.NewReader(parsedOutput.Bytes()))
		vms = append(vms, vm.vm)
	}
	eq := CompareFiles(vms, readers)
	if !eq {
		t.Errorf("Expected equality, didn't get it")
	}
}
