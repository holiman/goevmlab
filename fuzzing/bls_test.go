package fuzzing

import (
	"testing"
)

func TestBls(t *testing.T) {
	if have, want := len(newG1Point()), 128; have != want {
		t.Errorf("Generated input wrong, have %d want %d", have, want)
	}
	if have, want := len(newG2Point()), 256; have != want {
		t.Errorf("Generated input wrong, have %d want %d", have, want)
	}
	if have, want := len(newFPtoG1()), 64; have != want {
		t.Errorf("Generated input wrong, have %d want %d", have, want)
	}
	if have, want := len(newFP2toG2()), 128; have != want {
		t.Errorf("Generated input wrong, have %d want %d", have, want)
	}
	// 160 * K
	// k slices each of them being a byte concatenation of encoding of a
	// G1 point (128 bytes) and encoding of a scalar value (32 bytes)
	if have, want := len(newG1MSM()), 160; have != want {
		t.Errorf("Generated input wrong, have %d want %d", have, want)
	}
	// 288 * K
	// k slices each of them being a byte concatenation of encoding of
	// G2 point (256 bytes) and encoding of a scalar value (32 bytes).
	if have, want := len(newG2MSM()), 288; have != want {
		t.Errorf("Generated input wrong, have %d want %d", have, want)
	}
}
