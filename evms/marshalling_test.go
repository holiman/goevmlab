package evms

import (
	"encoding/json"
	"testing"

	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/holiman/uint256"
)

// Test that marshalling is valid json
func TestMarshalling(t *testing.T) {
	log := new(logger.StructLog)
	for i := 0; i < 10; i++ {
		el := uint256.NewInt(uint64(i))
		log.Stack = append(log.Stack, *el)
	}
	if out := FastMarshal(log); !json.Valid(out) {
		t.Fatalf("invalid json: %v", string(out))
	}
}

func BenchmarkMarshalling(b *testing.B) {

	log := new(logger.StructLog)
	for i := 0; i < 10; i++ {
		el := uint256.NewInt(uint64(i))
		log.Stack = append(log.Stack, *el)
	}
	var outp1 []byte
	b.Run("json", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			outp1, _ = json.Marshal(log)
		}
	})
	var outp2 []byte
	b.Run("fast", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			outp2 = FastMarshal(log)
		}
	})
	b.Log(string(outp1))
	b.Log(string(outp2))
}
