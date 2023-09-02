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
	"path/filepath"
	"testing"
)

// TestVMsOutput in this test, we simulate several external
// vms, using printouts from actual evm binaries. The parsed outputs should
// not produce any differences.
func TestVMsOutput(t *testing.T) {
	finfos, err := os.ReadDir(filepath.Join("testdata", "cases"))
	if err != nil {
		t.Fatal(err)
	}
	for _, finfo := range finfos {
		testVmsOutput(t, filepath.Join("testdata", "traces", finfo.Name()))
	}
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
		{NewNimbusEVM("", ""), "", fmt.Sprintf("%v.nimbus.stderr.txt", testfile)},
		{NewEvmoneVM("", ""), "", fmt.Sprintf("%v.evmone.stderr.txt", testfile)},
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

func TestStateRootGeth(t *testing.T) {
	testStateRootOnly(t, NewGethEVM("", ""), "geth")
}

func TestStateRootBesu(t *testing.T) {
	testStateRootOnly(t, NewBesuVM("", ""), "besu")
}

func TestStateRootErigon(t *testing.T) {
	testStateRootOnly(t, NewErigonVM("", ""), "erigon")
}

func TestStateRootNethermind(t *testing.T) {
	testStateRootOnly(t, NewNethermindVM("", ""), "nethermind")
}

func TestStateRootNimbus(t *testing.T) {
	testStateRootOnly(t, NewNimbusEVM("", ""), "nimbus")
}

func TestStateRootEvmone(t *testing.T) {
	testStateRootOnly(t, NewEvmoneVM("", ""), "evmone")
}

func testStateRootOnly(t *testing.T, vm Evm, name string) {

	finfos, err := os.ReadDir(filepath.Join("testdata", "cases"))
	if err != nil {
		t.Fatal(err)
	}
	wants := map[string]string{
		"00000006-naivefuzz-0.json":      "0xad1024c87b5548e77c937aa50f72b6cb620d278f4dd79bae7f78f71ff75af458",
		"00003656-naivefuzz-0.json":      "0x75dc56643cc707a2e6c9a4cf7e28061e9598bd02ecac22c406365c058088d59b",
		"statetest1.json":                "0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134",
		"00016209-naivefuzz-0.json":      "0x9b732142c31ee7c3c1d28a1c5f451f555524e0bb39371d94a9698000203742fb",
		"statetest_filled.json":          "0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134",
		"stackUnderflow_nonzeroMem.json": "0x1f07fb182fd18ad9b11f8ef6cf369981e87e9f8514c803a1f2df145724f62fa4",
		"00000936-mixed-1.json":          "0xd14c10ed22a1cfb642e374be985ac581c39f3969bd59249e0405aca3beb47a47",
		"negative_refund.json":           "0xee0bbf0438796320ede24ca3c52e31f04dccbfe1fce282f79fe44e67a23351e9",
	}
	for i, finfo := range finfos {
		testfile := filepath.Join("testdata", "roots", finfo.Name())
		stderr, _ := os.ReadFile(fmt.Sprintf("%v.%v.stderr.txt", testfile, name))
		stdout, _ := os.ReadFile(fmt.Sprintf("%v.%v.stdout.txt", testfile, name))
		combined := append(stderr, stdout...)
		have, err := vm.ParseStateRoot(combined)
		if err != nil {
			t.Fatalf("case %d, %v: got error: %v", i, finfo.Name(), err)
		}
		want, ok := wants[finfo.Name()]
		if !ok {
			t.Fatalf("Test error! A new trace (%v) has been added, but the corresponding stateroot has not been added", finfo.Name())
		}
		if have != want {
			t.Errorf("case %d, %v: have '%v' want '%v'", i, finfo.Name(), have, want)
		}
	}
}
