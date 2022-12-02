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
	"fmt"
	"io"
	"os"
	"testing"
)

// TestVMsOutput in this test, we simulate several external
// vms, using printouts from actual evm binaries. The parsed outputs should
// not produce any differences.
func TestVMsOutput(t *testing.T) {
	testVmsOutput(t, "testdata/statetest1.json")
	testVmsOutput(t, "testdata/statetest_filled.json")
	testVmsOutput(t, "testdata/00016209-naivefuzz-0.json")
	testVmsOutput(t, "testdata/00000006-naivefuzz-0.json")
}

func testVmsOutput(t *testing.T, testfile string) {
	type testCase struct {
		vm     Evm
		stdout string
		stderr string
	}
	var cases = []testCase{
		{NewBesuVM("", ""), fmt.Sprintf("%v.besu.stdout.txt", testfile), ""},
		{NewBesuBatchVM("", ""), fmt.Sprintf("%v.besu.stdout.txt", testfile), ""},
		{NewNethermindVM("", ""), "", fmt.Sprintf("%v.nethermind.stderr.txt", testfile)},
		{NewErigonVM("", ""), "", fmt.Sprintf("%v.erigon.stderr.txt", testfile)},
		{NewGethEVM("", ""), "", fmt.Sprintf("%v.geth.stderr.txt", testfile)},
	}
	var readers []io.Reader
	var vms []Evm
	for _, tc := range cases {
		parsedOutput := bytes.NewBuffer(nil)
		// Read the stdout and stderr outputs
		for _, f := range []string{tc.stdout, tc.stderr} {
			if len(f) == 0 {
				continue
			}
			rawOutput, err := os.Open(f)
			if err != nil {
				t.Fatal(err)
			}
			// And feed via the vm-specific copy-shim
			tc.vm.Copy(parsedOutput, rawOutput)
			rawOutput.Close()
		}
		readers = append(readers, bytes.NewReader(parsedOutput.Bytes()))
		vms = append(vms, tc.vm)
	}
	if eq, _ := CompareFiles(vms, readers); !eq {
		t.Errorf("Expected equality, didn't get it, file: %v", testfile)
	}
}

// TestStateRootOnly checks if the functionality to extract raw stateroot works
func TestStateRootOnly(t *testing.T) {
	t.Skip("Test is machine-specific due to bundled binaries")
	vms := []Evm{
		NewGethEVM("../binaries/evm", ""),
		NewNethermindVM("../binaries/nethtest", ""),
	}
	for _, vm := range vms {
		got, _, err := vm.GetStateRoot("./testdata/statetest1.json")
		if err != nil {
			t.Errorf("got error: %v", err)
		} else if exp := "0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134"; got != exp {
			t.Errorf("Wrong root, got '%v' exp '%v'", got, exp)
		}
		// A filled statetest
		// It would be good to get this working too, but not as important
		//{
		//	got, err := g.GetStateRoot("./testdata/statetest_filled.json")
		//	if err != nil {
		//		t.Errorf("got error: %v", err)
		//	}else if exp := "0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134"; got != exp {
		//		t.Errorf("Wrong root, got '%v' exp '%v'", got, exp)
		//	}
		//}
	}
}
