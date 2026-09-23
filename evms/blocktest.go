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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// ErrBlockTestUnsupported is returned by evms which cannot execute blocktests.
var ErrBlockTestUnsupported = errors.New("blocktest not supported by this evm")

// BlockTestMode configures how a blocktest is executed.
type BlockTestMode struct {
	// Engine selects the blockchain_test_engine fixture flavour, where each block
	// is delivered as an engine_newPayload with the block access list attached.
	// In this mode, clients may use the BAL-driven parallel execution path.
	// If false, the blockchain_test flavour (RLP blocks) is used, and clients
	// construct the block access list themselves.
	Engine bool
	// Trace enables opcode tracing. Tracing is not compatible with parallel
	// execution, so Engine mode is typically run without tracing.
	Trace bool
}

func (m BlockTestMode) String() string {
	flavour := "construct"
	if m.Engine {
		flavour = "consume"
	}
	if m.Trace {
		return flavour + "+trace"
	}
	return flavour
}

// blockTestEnd is the normalized terminator line for blocktest outputs. Every
// evm ends its canonical blocktest output with one such line, so that both
// the resulting state root and the verdict are part of the comparison.
type blockTestEnd struct {
	StateRoot string `json:"stateRoot"`
	Pass      bool   `json:"pass"`
}

// btTestEnd is the end marker emitted by nethermind (and possibly others) in
// the trace stream: {"testEnd":{"name":..,"pass":..,"fork":..,"root":..}}
type btTestEnd struct {
	Name string `json:"name"`
	Pass bool   `json:"pass"`
	Fork string `json:"fork"`
	Root string `json:"root"`
}

// btResult is an entry in the JSON result array printed by geth's and
// nethermind's test runners on stdout.
type btResult struct {
	Name          string `json:"name"`
	Pass          bool   `json:"pass"`
	StateRoot     string `json:"stateRoot"`
	LastBlockHash string `json:"lastBlockHash"`
	Fork          string `json:"fork"`
	Error         string `json:"error"`
}

// writeBlockTestEnd writes the terminator line to the output.
func writeBlockTestEnd(out io.Writer, end *blockTestEnd) {
	if end == nil {
		end = &blockTestEnd{}
	}
	data, _ := json.Marshal(end)
	if _, err := out.Write(append(data, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
	}
}

// decodeBlockTestResults decodes one JSON result array from the reader. The
// reader may contain several arrays back-to-back (batch mode), in which case
// the decoder must be reused.
func decodeBlockTestResults(dec *json.Decoder) (*blockTestEnd, error) {
	var results []btResult
	if err := dec.Decode(&results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("empty result array")
	}
	// A blocktest file may in theory contain several tests, but goevmlab
	// produces one per file. Combine: all must pass, root from the last.
	end := &blockTestEnd{Pass: true}
	for _, r := range results {
		if !r.Pass {
			end.Pass = false
		}
		end.StateRoot = r.StateRoot
	}
	if !end.Pass {
		end.StateRoot = ""
	}
	return end, nil
}

// copyBlockTestTrace reads a jsonl trace stream for one blocktest, writes the
// canonical opcode lines to out, and stops at the end marker. It handles both
// the geth-style marker ({"stateRoot":..,"pass":..}) and the nethermind-style
// marker ({"testEnd":{..}}). If the stream ends without a marker, nil is
// returned. The terminator line is not written; that is up to the caller.
//
// If dedupe is set, consecutive duplicate lines (same pc, depth) are merged,
// which is needed for geth output (see GethEVM.copyUntilEnd).
func copyBlockTestTrace(name string, out io.Writer, input io.Reader, dedupe bool) *blockTestEnd {
	scanner := NewJsonlScanner(name, input, os.Stderr)
	defer scanner.Release()
	var (
		end  *blockTestEnd
		prev *opLog
	)
	var yield = func(current *opLog) {
		if !dedupe {
			if current != nil {
				data := CustomMarshal(current)
				if _, err := out.Write(append(data, '\n')); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
				}
			}
			return
		}
		if prev == nil {
			prev = current
			return
		}
		data := CustomMarshal(prev)
		if _, err := out.Write(append(data, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to out: %v\n", err)
		}
		if current == nil { // final flush
			return
		}
		if prev.Pc == current.Pc && prev.Depth == current.Depth && prev.FunctionDepth == current.FunctionDepth {
			prev = nil
		} else {
			prev = current
		}
	}
	for {
		var elem opLog
		if err := scanner.Next(&elem); err != nil {
			break
		}
		if elem.TestEnd != nil {
			end = &blockTestEnd{StateRoot: elem.TestEnd.Root, Pass: elem.TestEnd.Pass}
			break
		}
		if len(elem.StateRoot1) != 0 {
			end = &blockTestEnd{StateRoot: elem.StateRoot1, Pass: true}
			if elem.Pass != nil {
				end.Pass = *elem.Pass
			}
			break
		}
		// Lines which are not opcodes (tx summaries, errors) have depth 0
		if elem.Depth == 0 {
			continue
		}
		// Drop STOPs, see GethEVM.copyUntilEnd
		if elem.Op == 0x0 {
			continue
		}
		yield(&elem)
	}
	yield(nil)
	if end != nil && !end.Pass {
		// On failure, the root is not meaningful for comparison: clients
		// report the root at different points of failure.
		end.StateRoot = ""
	}
	return end
}

// batchProc is a persistent test-runner process which reads test file paths
// from stdin, and emits either a trace (with end marker) on stderr, or a JSON
// result array on stdout, per test.
type batchProc struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	out   io.ReadCloser
	dec   *json.Decoder // only used when not tracing
	mode  BlockTestMode
}

// startBatchProc starts the process. If the mode has tracing enabled, the
// trace stream (stderr) is consumed; otherwise the result stream (stdout).
func startBatchProc(path string, args []string, mode BlockTestMode) (*batchProc, error) {
	var (
		p   = &batchProc{cmd: exec.Command(path, args...), mode: mode}
		err error
	)
	if mode.Trace {
		p.out, err = p.cmd.StderrPipe()
	} else {
		p.out, err = p.cmd.StdoutPipe()
		p.dec = json.NewDecoder(p.out)
	}
	if err != nil {
		return p, err
	}
	if p.stdin, err = p.cmd.StdinPipe(); err != nil {
		return p, err
	}
	if err = p.cmd.Start(); err != nil {
		return p, err
	}
	return p, nil
}

// run feeds one test to the process and returns the result. In traced mode,
// the given copy function is used to consume the trace.
func (p *batchProc) run(path string, copyTrace func(io.Reader) *blockTestEnd) (*blockTestEnd, error) {
	if _, err := fmt.Fprintf(p.stdin, "%v\n", path); err != nil {
		return nil, fmt.Errorf("error writing to %v: %w", p.cmd.Path, err)
	}
	if p.mode.Trace {
		end := copyTrace(p.out)
		if end == nil {
			return nil, fmt.Errorf("%v: trace stream ended without end marker", p.cmd.Path)
		}
		return end, nil
	}
	return decodeBlockTestResults(p.dec)
}

func (p *batchProc) String() string {
	if p == nil || p.cmd == nil {
		return ""
	}
	return p.cmd.String()
}

// reusable returns true if the process can be used for the given mode. A
// process started in one mode cannot execute tests in another mode.
func (p *batchProc) reusable(mode BlockTestMode) bool {
	return p != nil && p.mode == mode
}

func (p *batchProc) close() {
	if p == nil {
		return
	}
	if p.stdin != nil {
		p.stdin.Close()
	}
	if p.cmd != nil {
		_ = p.cmd.Wait()
	}
}

// fixupBlockTestRoot fills in the state root of a passing test from the test
// fixture, for runners which do not report it (nethermind reports the root of
// blocktests in the trace end marker, but not in the result array).
func fixupBlockTestRoot(end *blockTestEnd, path string, mode BlockTestMode) *blockTestEnd {
	if end == nil || !end.Pass || len(end.StateRoot) > 0 {
		return end
	}
	root, err := fixtureHeadRoot(path, mode.Engine)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read expected root from %v: %v\n", path, err)
		return end
	}
	end.StateRoot = root
	return end
}

// fixtureHeadRoot reads the state root of the last block from a fixture file.
func fixtureHeadRoot(path string, engine bool) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var tests map[string]struct {
		Blocks []struct {
			Header struct {
				StateRoot string `json:"stateRoot"`
			} `json:"blockHeader"`
		} `json:"blocks"`
		Payloads []struct {
			Params []json.RawMessage `json:"params"`
		} `json:"engineNewPayloads"`
	}
	if err := json.Unmarshal(data, &tests); err != nil {
		return "", err
	}
	var root string
	for _, test := range tests {
		if engine {
			if len(test.Payloads) == 0 || len(test.Payloads[len(test.Payloads)-1].Params) == 0 {
				return "", errors.New("no payloads in fixture")
			}
			var payload struct {
				StateRoot string `json:"stateRoot"`
			}
			if err := json.Unmarshal(test.Payloads[len(test.Payloads)-1].Params[0], &payload); err != nil {
				return "", err
			}
			root = payload.StateRoot
		} else {
			if len(test.Blocks) == 0 {
				return "", errors.New("no blocks in fixture")
			}
			root = test.Blocks[len(test.Blocks)-1].Header.StateRoot
		}
	}
	if len(root) == 0 {
		return "", errors.New("no state root in fixture")
	}
	return root, nil
}

// SupportsBlockTests returns true if the given evm implements blocktest
// execution (as opposed to returning ErrBlockTestUnsupported).
func SupportsBlockTests(evm Evm) bool {
	switch evm.(type) {
	case *GethEVM, *GethBatchVM, *NethermindVM, *NethermindBatchVM:
		return true
	}
	return false
}
