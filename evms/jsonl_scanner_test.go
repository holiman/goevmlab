package evms

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"os"
	"strings"
	"testing"
)

func TestJsonlScanner(t *testing.T) {
	input, err := os.ReadFile("./testdata/json_scanner.input")
	if err != nil {
		t.Fatal(err)
	}
	stderr := new(strings.Builder)
	stdout := new(strings.Builder)
	scanner := NewJsonlScanner("test", bytes.NewReader(input), stderr)
	var elem logger.StructLog
	for err := scanner.Next(&elem); err == nil; err = scanner.Next(&elem) {
		fmt.Fprintln(stdout, elem.OpName())
	}

	{
		want, err := os.ReadFile("./testdata/json_scanner.stdout")
		if err != nil {
			t.Fatal(err)
		}
		have := []byte(stdout.String())
		if !bytes.Equal(have, want) {
			t.Fatalf("Unexpected output on std-output: have\n%v\nwant\n%v\n", string(have), string(want))
		}
	}
	{
		want, err := os.ReadFile("./testdata/json_scanner.stderr")
		if err != nil {
			t.Fatal(err)
		}
		have := []byte(stderr.String())
		if !bytes.Equal(have, want) {
			t.Fatalf("Unexpected output on error-output: have\n%v\nwant\n%v\n", string(have), string(want))
		}
	}
}
