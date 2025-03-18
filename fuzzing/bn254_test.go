package fuzzing

import (
	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"testing"
)

func TestBn254(t *testing.T) {
	if have, want := len(newBnAdd()), 128; have != want {
		t.Errorf("Generated input wrong, have %d want %d", have, want)
	}
	if have, want := len(newBnScalarMul()), 96; have != want {
		t.Errorf("Generated input wrong, have %d want %d", have, want)
	}
	// 160 * K
	// k slices each of them being a byte concatenation of encoding of a
	// G1 point (128 bytes) and encoding of a scalar value (32 bytes)
	if have, multiple := len(newBnPairing()), 192; have%multiple != 0 {
		t.Errorf("Generated input wrong, have %d want multiple of %d", have, multiple)
	}
}

func TestBn254ErrorTypes(t *testing.T) {
	var ok = 0
	for i := 0; i < 1000; i++ {
		err := decodeBn254G1(makeBadBn254G1())
		if err == nil {
			ok++
			//}else{
			//	fmt.Printf("err: %v\n", err)
		}
	}
	// 792 out of 1000 were ok
	t.Logf("G1 Ok : %d", ok)

	ok = 0
	for i := 0; i < 1000; i++ {
		err := decodeBn254G2(makeBadBn254G2())
		if err == nil {
			ok++
			//}else{
			//	fmt.Printf("err: %v\n", err)
		}
	}
	// 814 out of 1000 were ok
	t.Logf("G2 Ok : %d", ok)
}

func decodeBn254G1(blob []byte) error {
	p := new(bn256.G1)
	_, err := p.Unmarshal(blob)
	return err
}

func decodeBn254G2(blob []byte) error {
	p := new(bn256.G2)
	_, err := p.Unmarshal(blob)
	return err
}
