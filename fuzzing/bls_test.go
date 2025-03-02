package fuzzing

import (
	"errors"
	"fmt"
	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
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
	if have, multiple := len(newG1MSM()), 160; have%multiple != 0 {
		t.Errorf("Generated input wrong, have %d want multiple of %d", have, multiple)
	}
	// 288 * K
	// k slices each of them being a byte concatenation of encoding of
	// G2 point (256 bytes) and encoding of a scalar value (32 bytes).
	if have, multiple := len(newG2MSM()), 288; have%multiple != 0 {
		t.Errorf("Generated input wrong, have %d want multiple of %d", have, multiple)
	}
	if have, multiple := len(newPairing()), 384; have%multiple != 0 {
		t.Errorf("Generated input wrong, have %d want multiple of %d", have, multiple)
	}
}

// decodeBLS12381FieldElement decodes BLS12-381 elliptic curve field element.
// Removes top 16 bytes of 64 byte input.
func decodeBLS12381FieldElement(in []byte) (fp.Element, error) {
	if len(in) != 64 {
		return fp.Element{}, errors.New("invalid field element length")
	}
	// check top bytes
	for i := 0; i < 16; i++ {
		if in[i] != byte(0x00) {
			return fp.Element{}, errors.New("bad top bytes")
		}
	}
	var res [48]byte
	copy(res[:], in[16:])

	return fp.BigEndian.Element(&res)
}

func decodePointG1(in []byte) (*bls12381.G1Affine, error) {
	if len(in) != 128 {
		return nil, errors.New("invalid g1 point length")
	}
	// decode x
	x, err := decodeBLS12381FieldElement(in[:64])
	if err != nil {
		return nil, err
	}
	// decode y
	y, err := decodeBLS12381FieldElement(in[64:])
	if err != nil {
		return nil, err
	}
	elem := bls12381.G1Affine{X: x, Y: y}
	if !elem.IsOnCurve() {
		return nil, errors.New("invalid point: not on curve")
	}

	return &elem, nil
}

// decodePointG2 given encoded (x, y) coordinates in 256 bytes returns a valid G2 Point.
func decodePointG2(in []byte) (*bls12381.G2Affine, error) {
	if len(in) != 256 {
		return nil, errors.New("invalid g2 point length")
	}
	x0, err := decodeBLS12381FieldElement(in[:64])
	if err != nil {
		return nil, err
	}
	x1, err := decodeBLS12381FieldElement(in[64:128])
	if err != nil {
		return nil, err
	}
	y0, err := decodeBLS12381FieldElement(in[128:192])
	if err != nil {
		return nil, err
	}
	y1, err := decodeBLS12381FieldElement(in[192:])
	if err != nil {
		return nil, err
	}

	p := bls12381.G2Affine{X: bls12381.E2{A0: x0, A1: x1}, Y: bls12381.E2{A0: y0, A1: y1}}
	if !p.IsOnCurve() {
		return nil, errors.New("invalid point: not on curve")
	}
	return &p, err

}

func TestErrorTypes(t *testing.T) {
	for i := 0; i < 1000; i++ {
		input := makeBadG1()
		_, err := decodePointG1(input)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}
	for i := 0; i < 100; i++ {
		input := makeBadG2()
		_, err := decodePointG2(input)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}
}
