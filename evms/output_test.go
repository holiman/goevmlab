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
	"embed"
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
		if finfo.Name() == "eofcode.json" {
			// We skip this one. Evmone refuse to run it.
			// https://github.com/holiman/goevmlab/issues/127
			continue
		}
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
		{NewBesuVM("", "besuvm"), fmt.Sprintf("%v.besu.stdout.txt", testfile), ""},
		{NewBesuBatchVM("", "besuba"), fmt.Sprintf("%v.besu.stdout.txt", testfile), ""},
		{NewNethermindVM("", "nether"), "", fmt.Sprintf("%v.nethermind.stderr.txt", testfile)},
		{NewErigonVM("", "erigon"), "", fmt.Sprintf("%v.erigon.stderr.txt", testfile)},
		{NewGethEVM("", "gethvm"), "", fmt.Sprintf("%v.geth.stderr.txt", testfile)},
		{NewNimbusEVM("", "nimbus"), "", fmt.Sprintf("%v.nimbus.stderr.txt", testfile)},
		{NewNimbusBatchVM("", "nimbusba"), "", fmt.Sprintf("%v.nimbus.stderr.txt", testfile)},
		{NewEvmoneVM("", "evmone"), "", fmt.Sprintf("%v.evmone.stderr.txt", testfile)},
		{NewRethVM("", "rethvm"), "", fmt.Sprintf("%v.revm.stderr.txt", testfile)},
		{NewEelsEVM("", "eelsvm"), "", fmt.Sprintf("%v.eels.stderr.txt", testfile)},
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
	if eq, _, data := CompareFiles(vms, readers); !eq {
		t.Log(data)
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

func TestStateRootRethVM(t *testing.T) {
	testStateRootOnly(t, NewRethVM("", ""), "revm")
}

func TestStateRootEelsVM(t *testing.T) {
	testStateRootOnly(t, NewEelsEVM("", "eels"), "eels")
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
		"eofcode.json":                   "0x53f6733a696cb3bbf77b635d96ace97f25ffee2d08d3e3d4ae1e566bfc060d6f",
	}
	for i, finfo := range finfos {
		testfile := filepath.Join("testdata", "roots", finfo.Name())
		stderr, _ := os.ReadFile(fmt.Sprintf("%v.%v.stderr.txt", testfile, name))
		stdout, _ := os.ReadFile(fmt.Sprintf("%v.%v.stdout.txt", testfile, name))
		combined := append(stderr, stdout...)
		have, err := vm.ParseStateRoot(combined)
		if err != nil {
			if finfo.Name() == "eofcode.json" {
				// We accept this failure. Evmone refuse to run it.
				// https://github.com/holiman/goevmlab/issues/127
				t.Logf("case %d, %v: got error: %v (failure accepted)", i, finfo.Name(), err)
				continue
			}
			t.Errorf("case %d, %v: got error: %v", i, finfo.Name(), err)
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

// We embed the reference cases here
//
//go:embed testdata/cases
var testcases embed.FS

// createEvmsFromEnv instantiates vms bsaed on ENV info.
func createEvmsFromEnv() []Evm {
	var vms []Evm
	if k := "GETH_BIN"; os.Getenv(k) != "" {
		vms = append(vms, NewGethEVM(os.Getenv(k), "geth"))
		vms = append(vms, NewGethBatchVM(os.Getenv(k), "gethbatch"))
	}
	if k := "NETH_BIN"; os.Getenv(k) != "" {
		vms = append(vms, NewNethermindVM(os.Getenv(k), "neth"))
		vms = append(vms, NewNethermindBatchVM(os.Getenv(k), "nethbatch"))
	}
	if k := "NIMB_BIN"; os.Getenv(k) != "" {
		vms = append(vms, NewNimbusEVM(os.Getenv(k), "nimb"))
		vms = append(vms, NewNimbusBatchVM(os.Getenv(k), "nimbbatch"))
	}
	if k := "RETH_BIN"; os.Getenv(k) != "" {
		vms = append(vms, NewRethVM(os.Getenv(k), "reth"))
	}
	if k := "ERIG_BIN"; os.Getenv(k) != "" {
		vms = append(vms, NewErigonVM(os.Getenv(k), "erigon"))
		vms = append(vms, NewErigonBatchVM(os.Getenv(k), "erigonbatch"))
	}
	if k := "BESU_BIN"; os.Getenv(k) != "" {
		vms = append(vms, NewBesuVM(os.Getenv(k), "besu"))
		vms = append(vms, NewBesuBatchVM(os.Getenv(k), "besubatch"))
	}
	if k := "EVMO_BIN"; os.Getenv(k) != "" {
		vms = append(vms, NewEvmoneVM(os.Getenv(k), "evmone"))
	}
	if k := "EELS_BIN"; os.Getenv(k) != "" {
		vms = append(vms, NewEelsEVM(os.Getenv(k), "eels"))
		vms = append(vms, NewEelsBatchVM(os.Getenv(k), "eelsbatch"))
	}
	return vms
}

// writeReferenceTestsToDisk writes the testcases from embedding to actual disk,
// so the external vms can execute them.
func writeReferenceTestsToDisk(t *testing.T) []string {
	t.Helper()
	var testfiles []string
	path := filepath.Join("testdata", "cases")
	embedded, err := testcases.ReadDir(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, finfo := range embedded {
		if finfo.Name() == "eofcode.json" {
			// We skip this one. Evmone refuse to run it.
			// https://github.com/holiman/goevmlab/issues/127
			continue
		}
		src := filepath.Join(path, finfo.Name())
		dst := filepath.Join(dir, finfo.Name())
		data, _ := testcases.ReadFile(src)
		_ = os.WriteFile(dst, data, 0777)
		testfiles = append(testfiles, dst)
		t.Logf("Copied embed:%v -> %v", src, dst)
	}
	return testfiles
}

// TestVMsFromEnv_tracing is meant to be used as a sanity-check, testing the
// tracing-functionality of an (external) evm is compatible with goevmlabm.
// The intended use for this is primarily from inside docke containers,
// where the client binaries are available.
//
// This test simply executes the reference tests on all clients, and compares the
// traces between them. The idea is to build this as a standalone binary, and bundle
// in the docker image:
//
//	go test -c ./evms
//
// And then, later on, do
//
//	./evms.test -test.run TestVMsFromEnv_tracing -test.v
func TestVMsFromEnv_tracing(t *testing.T) {
	vms := createEvmsFromEnv()
	t.Cleanup(func() {
		for _, vm := range vms {
			vm.Close()
		}
	})
	if len(vms) < 2 {
		t.Skipf("Need at least 2 vms for sanity-test, have %v, skipping", len(vms))
		return
	}
	// We need to write the testcases from embedding to actual disk, so
	// the vms can execute them
	testfiles := writeReferenceTestsToDisk(t)
	var readers = make([]io.Reader, len(vms))
	// Check the full-trace functionality
	for _, testfile := range testfiles {
		for i, vm := range vms {
			output := bytes.NewBuffer(nil)
			res, err := vm.RunStateTest(testfile, output, false)
			if err != nil {
				t.Fatal(err)
			}
			readers[i] = output
			t.Logf("Executed test, cmd: %q", res.Cmd)
		}
		equal, _, diff := CompareFiles(vms, readers)
		if !equal {
			t.Log(diff)
			t.Errorf("Difference found")
		}
	}
}

// TestVMsFromEnv_stateroot is meant to be used as a sanity-check, testing the
// ability of goevmlab to obtain a stateroot from a test execution.
// The intended use for this is primarily from inside docke containers,
// where the client binaries are available.
//
// This test simply executes the reference tests on all clients, and compares the
// stateroot against the known wanted roots. The idea is to build this as a
// standalone binary, and bundle in the docker image:
//
//	go test -c ./evms
//
// And then, later on, do
//
//	./evms.test -test.run TestVMsFromEnv_stateroot -test.v
func TestVMsFromEnv_stateroot(t *testing.T) {
	vms := createEvmsFromEnv()
	t.Cleanup(func() {
		for _, vm := range vms {
			vm.Close()
		}
	})
	testfiles := writeReferenceTestsToDisk(t)
	wants := map[string]string{
		"00000006-naivefuzz-0.json":      "0xad1024c87b5548e77c937aa50f72b6cb620d278f4dd79bae7f78f71ff75af458",
		"00003656-naivefuzz-0.json":      "0x75dc56643cc707a2e6c9a4cf7e28061e9598bd02ecac22c406365c058088d59b",
		"statetest1.json":                "0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134",
		"00016209-naivefuzz-0.json":      "0x9b732142c31ee7c3c1d28a1c5f451f555524e0bb39371d94a9698000203742fb",
		"statetest_filled.json":          "0xa2b3391f7a85bf1ad08dc541a1b99da3c591c156351391f26ec88c557ff12134",
		"stackUnderflow_nonzeroMem.json": "0x1f07fb182fd18ad9b11f8ef6cf369981e87e9f8514c803a1f2df145724f62fa4",
		"00000936-mixed-1.json":          "0xd14c10ed22a1cfb642e374be985ac581c39f3969bd59249e0405aca3beb47a47",
		"negative_refund.json":           "0xee0bbf0438796320ede24ca3c52e31f04dccbfe1fce282f79fe44e67a23351e9",
		"eofcode.json":                   "0x53f6733a696cb3bbf77b635d96ace97f25ffee2d08d3e3d4ae1e566bfc060d6f",
	}
	// Check stateroot functionality
	for _, testfile := range testfiles {
		for _, vm := range vms {
			root, cmd, err := vm.GetStateRoot(testfile)
			if err != nil {
				t.Fatal(err)
			}
			fname := filepath.Base(testfile)
			want := wants[fname]
			if want != root {
				t.Errorf("Wrong root, have %v, want %v, file %v, cmd %q", root, want, fname, cmd)
			}
			t.Logf("Executed test %v, root %v, cmd %q", fname, root, cmd)
		}
	}
}
