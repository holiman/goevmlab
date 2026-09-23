// Copyright 2026 Martin Holst Swende
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
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// A trace as emitted by geth's evm blocktest --fuzz --trace (stderr), with
// the (proposed) end marker.
const gethBlockTrace = `{"pc":0,"op":96,"gas":"0x14c08","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":2,"op":96,"gas":"0x14c05","gasCost":"0x3","memSize":0,"stack":["0x1"],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":4,"op":85,"gas":"0x14c02","gasCost":"0x2f44","memSize":0,"stack":["0x1","0x0"],"depth":1,"refund":0,"opName":"SSTORE","error":"out of gas"}
{"pc":4,"op":85,"gas":"0x14c02","gasCost":"0x2f44","memSize":0,"stack":["0x1","0x0"],"depth":1,"refund":0,"opName":"SSTORE","error":"out of gas"}
{"output":"","gasUsed":"0x14c08","error":"out of gas"}
{"pc":0,"op":96,"gas":"0x14c08","gasCost":"0x3","memSize":0,"stack":[],"depth":1,"refund":0,"opName":"PUSH1"}
{"pc":2,"op":0,"gas":"0x14c05","gasCost":"0x0","memSize":0,"stack":["0x1"],"depth":1,"refund":0,"opName":"STOP"}
{"output":"","gasUsed":"0x5208"}
{"stateRoot":"0xfc6b324ef9527cabee7bb2596662998ce9df900a5efc96929c0fb5a942d4728b","pass":true}
`

// A trace as emitted by nethtest -b -t (stderr), with the testEnd marker.
const nethBlockTrace = `{"pc":0,"op":96,"gas":"0x14c08","gasCost":"0x3","memSize":0,"stack":[],"depth":1}
{"pc":2,"op":96,"gas":"0x14c05","gasCost":"0x3","memSize":0,"stack":["0x1"],"depth":1}
{"pc":4,"op":85,"gas":"0x14c02","gasCost":"0x2f44","memSize":0,"stack":["0x1","0x0"],"depth":1,"error":"out of gas"}
{"output":"","gasUsed":"0x14c08"}
{"pc":0,"op":96,"gas":"0x14c08","gasCost":"0x3","memSize":0,"stack":[],"depth":1}
{"pc":2,"op":0,"gas":"0x14c05","gasCost":"0x0","memSize":0,"stack":["0x1"],"depth":1}
{"output":"","gasUsed":"0x5208"}
{"testEnd":{"name":"test","pass":true,"fork":"Amsterdam","v":1,"gasUsed":"0x19e10","txs":2,"blocks":1,"root":"0xfc6b324ef9527cabee7bb2596662998ce9df900a5efc96929c0fb5a942d4728b"}}
`

func TestBlockTestTraceNormalization(t *testing.T) {
	var gethOut, nethOut bytes.Buffer
	gethEnd := copyBlockTestTrace("geth", &gethOut, strings.NewReader(gethBlockTrace), true)
	nethEnd := copyBlockTestTrace("neth", &nethOut, strings.NewReader(nethBlockTrace), false)
	if gethEnd == nil || nethEnd == nil {
		t.Fatalf("missing end marker: geth %v, neth %v", gethEnd, nethEnd)
	}
	if *gethEnd != *nethEnd {
		t.Fatalf("end markers differ: %+v vs %+v", gethEnd, nethEnd)
	}
	if !gethEnd.Pass || gethEnd.StateRoot != "0xfc6b324ef9527cabee7bb2596662998ce9df900a5efc96929c0fb5a942d4728b" {
		t.Fatalf("unexpected end marker %+v", gethEnd)
	}
	writeBlockTestEnd(&gethOut, gethEnd)
	writeBlockTestEnd(&nethOut, nethEnd)
	if gethOut.String() != nethOut.String() {
		t.Fatalf("outputs differ:\n%v\n----\n%v", gethOut.String(), nethOut.String())
	}
	// 3 + 1 opcodes (STOPs dropped, duplicate merged), plus terminator
	if lines := strings.Count(gethOut.String(), "\n"); lines != 5 {
		t.Fatalf("expected 5 lines, got %d:\n%v", lines, gethOut.String())
	}
	if !strings.HasSuffix(gethOut.String(), `{"stateRoot":"0xfc6b324ef9527cabee7bb2596662998ce9df900a5efc96929c0fb5a942d4728b","pass":true}`+"\n") {
		t.Fatalf("bad terminator:\n%v", gethOut.String())
	}
}

func TestBlockTestTraceNoMarker(t *testing.T) {
	// Without a marker (older geth), nil is returned, but the trace is
	// still written.
	var out bytes.Buffer
	input := strings.Join(strings.Split(gethBlockTrace, "\n")[:8], "\n")
	if end := copyBlockTestTrace("geth", &out, strings.NewReader(input), true); end != nil {
		t.Fatalf("expected no end marker, got %+v", end)
	}
	if lines := strings.Count(out.String(), "\n"); lines != 4 {
		t.Fatalf("expected 4 lines, got %d:\n%v", lines, out.String())
	}
}

func TestBlockTestTraceFailure(t *testing.T) {
	input := `{"pc":0,"op":96,"gas":"0x14c08","gasCost":"0x3","memSize":0,"stack":[],"depth":1}
{"testEnd":{"name":"test","pass":false,"fork":"Amsterdam","v":1,"root":"0xfc6b324ef9527cabee7bb2596662998ce9df900a5efc96929c0fb5a942d4728b"}}
`
	end := copyBlockTestTrace("neth", &bytes.Buffer{}, strings.NewReader(input), false)
	if end == nil || end.Pass || end.StateRoot != "" {
		t.Fatalf("unexpected end marker %+v", end)
	}
}

func TestBlockTestResults(t *testing.T) {
	// Two result arrays back-to-back, as in batch mode
	input := `[
  {
    "name": "a",
    "pass": true,
    "stateRoot": "0xfc6b324ef9527cabee7bb2596662998ce9df900a5efc96929c0fb5a942d4728b",
    "fork": "Amsterdam"
  }
]
[
  {
    "name": "b",
    "pass": false,
    "fork": "Amsterdam",
    "error": "access list hash mismatch"
  }
]
`
	dec := json.NewDecoder(strings.NewReader(input))
	end, err := decodeBlockTestResults(dec)
	if err != nil {
		t.Fatal(err)
	}
	if !end.Pass || end.StateRoot != "0xfc6b324ef9527cabee7bb2596662998ce9df900a5efc96929c0fb5a942d4728b" {
		t.Fatalf("unexpected result %+v", end)
	}
	end, err = decodeBlockTestResults(dec)
	if err != nil {
		t.Fatal(err)
	}
	if end.Pass || end.StateRoot != "" {
		t.Fatalf("unexpected result %+v", end)
	}
	if _, err = decodeBlockTestResults(dec); err == nil {
		t.Fatal("expected error at end of stream")
	}
}

func TestFixtureHeadRoot(t *testing.T) {
	root, err := fixtureHeadRoot("testdata/blockcases/blocktest1.json", false)
	if err != nil {
		t.Fatal(err)
	}
	eroot, err := fixtureHeadRoot("testdata/blockcases/blocktest1.engine.json", true)
	if err != nil {
		t.Fatal(err)
	}
	if root != eroot || len(root) != 66 {
		t.Fatalf("roots differ or malformed: %v %v", root, eroot)
	}
}

// TestBlockTestsFromEnv runs the reference blocktests on the binaries given
// via environment variables (GETH_BIN, NETH_BIN), if set, and checks that the
// outputs of all of them are identical.
func TestBlockTestsFromEnv(t *testing.T) {
	var vms []Evm
	if p := os.Getenv("GETH_BIN"); p != "" {
		vms = append(vms, NewGethEVM(p, "geth"), NewGethBatchVM(p, "gethbatch"))
	}
	if p := os.Getenv("NETH_BIN"); p != "" {
		vms = append(vms, NewNethermindVM(p, "neth"), NewNethermindBatchVM(p, "nethbatch"))
	}
	if len(vms) == 0 {
		t.Skip("no binaries configured (GETH_BIN, NETH_BIN)")
	}
	defer func() {
		for _, vm := range vms {
			vm.Close()
		}
	}()
	for _, tc := range []struct {
		file string
		mode BlockTestMode
	}{
		{"testdata/blockcases/blocktest1.json", BlockTestMode{Trace: true}},
		{"testdata/blockcases/blocktest1.json", BlockTestMode{Trace: false}},
		{"testdata/blockcases/blocktest1.engine.json", BlockTestMode{Engine: true}},
	} {
		t.Run(tc.mode.String(), func(t *testing.T) {
			var (
				outputs []*bytes.Buffer
				names   []string
			)
			for _, vm := range vms {
				var out bytes.Buffer
				res, err := vm.RunBlockTest(tc.file, &out, tc.mode)
				if err != nil {
					t.Logf("%v: %v (cmd %v)", vm.Name(), err, res.Cmd)
					continue
				}
				t.Logf("%v: %d bytes of output, last line %v", vm.Name(), out.Len(), lastLine(out.String()))
				outputs = append(outputs, &out)
				names = append(names, vm.Name())
			}
			for i := 1; i < len(outputs); i++ {
				if outputs[i].String() != outputs[0].String() {
					t.Errorf("output of %v differs from %v", names[i], names[0])
				}
			}
			if len(outputs) > 0 && !strings.Contains(lastLine(outputs[0].String()), `"pass":true`) {
				t.Errorf("test did not pass: %v", lastLine(outputs[0].String()))
			}
		})
	}
}

func lastLine(s string) string {
	s = strings.TrimRight(s, "\n")
	if i := strings.LastIndex(s, "\n"); i >= 0 {
		return s[i+1:]
	}
	return s
}

// TestBlockTestRecorded feeds recorded client outputs (see testdata/run.sh)
// through the blocktest output normalization, and checks them against the
// expected canonical output. Clients for which no recording exists are skipped.
func TestBlockTestRecorded(t *testing.T) {
	expected, err := os.ReadFile("testdata/blocktraces/blocktest1.json.expected.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		client string
		dedupe bool
	}{
		{"geth", true},
		{"nethermind", false},
	} {
		t.Run(tc.client, func(t *testing.T) {
			// Traced execution: stderr holds the trace and the end marker
			traced, err := os.ReadFile("testdata/blocktraces/blocktest1.json." + tc.client + ".stderr.txt")
			if err != nil {
				t.Skipf("no recording: %v", err)
			}
			var out bytes.Buffer
			end := copyBlockTestTrace(tc.client, &out, bytes.NewReader(traced), tc.dedupe)
			if end == nil {
				t.Fatal("no end marker found")
			}
			writeBlockTestEnd(&out, end)
			if out.String() != string(expected) {
				t.Errorf("canonical output differs from expected\nhave:\n%v\nwant:\n%v", out.String(), string(expected))
			}
			// Untraced execution: stdout holds the result array
			results, err := os.ReadFile("testdata/blocktraces/blocktest1.json." + tc.client + ".notrace.stdout.txt")
			if err != nil {
				t.Skipf("no recording: %v", err)
			}
			end, err = decodeBlockTestResults(json.NewDecoder(bytes.NewReader(results)))
			if err != nil {
				t.Fatal(err)
			}
			end = fixupBlockTestRoot(end, "testdata/blockcases/blocktest1.json", BlockTestMode{})
			var term bytes.Buffer
			writeBlockTestEnd(&term, end)
			if want := lastLine(string(expected)) + "\n"; term.String() != want {
				t.Errorf("terminator differs: have %v, want %v", term.String(), want)
			}
		})
	}
}
