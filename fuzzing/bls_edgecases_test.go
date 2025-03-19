// Copyright 2025 Martin Holst Swende and @kevaundray
// This file is part of the go-evmlab library.
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

// Most of the code in this file is written by @kevaundray
package fuzzing

import (
	"math/big"
	"testing"

	bls12381 "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

// Endomorphism for BLS12-381 (original curve)
func checkEndomorphismBLS(P bls12381.G1Affine) bool {
	var endoCoeff big.Int
	endoCoeff.SetString("793479390729215512621379701633421447060886740281060493010456487427281649075476305620758731620350", 10)

	// Apply endomorphism: phi(P) = (beta*x, y) where beta is the endomorphism coefficient
	var phiP bls12381.G1Affine
	var tmp fp.Element
	tmp.SetBigInt(&endoCoeff)
	phiP.X.Mul(&P.X, &tmp)
	phiP.Y = P.Y

	// Apply scalar multiplication: x²*P
	xGen := uint64(15132376222941642752)
	var x2P bls12381.G1Jac
	var x2 big.Int
	x2.SetUint64(xGen)
	x2.Mul(&x2, &x2)

	x2P.ScalarMultiplication(&bls12381.G1Jac{X: P.X, Y: P.Y, Z: fp.One()}, &x2)

	// Convert x²*P to affine
	var x2PAffine bls12381.G1Affine
	x2PAffine.FromJacobian(&x2P)

	// Add phi(P) + x²*P
	var result bls12381.G1Jac
	result.FromAffine(&phiP)
	var temp bls12381.G1Jac
	temp.FromAffine(&x2PAffine)
	result.AddAssign(&temp)

	// Check if the result is the point at infinity
	var resultAffine bls12381.G1Affine
	resultAffine.FromJacobian(&result)
	return resultAffine.X.IsZero() && resultAffine.Y.IsZero()
}

// Endomorphism check that takes in a twisted point but uses bls12-381 endo check
func checkEndomorphismTwisted(P twistedPoint) bool {
	// Use the BLS12-381 endomorphism check, this emulates the bug
	return checkEndomorphismBLS(bls12381.G1Affine{
		X: P.x,
		Y: P.y,
	})
}

// Check if a point is on the original BLS12-381 curve: y² = x³ + 4
func isPointOnBLS12381(x, y fp.Element) bool {
	var left, right, tmp fp.Element

	// Calculate left side: y²
	left.Square(&y)

	// Calculate right side: x³ + 4
	tmp.Square(&x)      // x²
	right.Mul(&tmp, &x) // x³
	var b fp.Element
	b.SetUint64(4)
	right.Add(&right, &b) // x³ + 4

	return left.Equal(&right)
}

func TestGenerateOffCurve(t *testing.T) {
	// Create a twisted curve (Et) with a = 0, b = 24 (instead of a = 0, b = 4 for E1)
	// For the twisted curve Et, we create points manually

	// Generate a point on the twisted curve (Et)
	pt := generatePointOnTwistedCurve()
	t.Logf("Generated point on twisted curve (Et): {x: %s, y: %s} ", pt.x.String(), pt.y.String())
	// Check if Pt satisfies the endomorphism check
	if passesTwisted := checkEndomorphismTwisted(pt); !passesTwisted {
		t.Fatalf("Pt does not pass BLS12-381 endomorphism check")
	}
	// Verify that Pt is NOT on the original BLS12-381 curve
	if isOnBLS := isPointOnBLS12381(pt.x, pt.y); isOnBLS {
		t.Fatalf("Pt is on original BLS12-381 curve")
	}
}

func BenchmarkGenerateOffCurve(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Generate a point on the twisted curve (Et)
		pt := generatePointOnTwistedCurve()
		// Check if Pt satisfies the endomorphism check
		if passesTwisted := checkEndomorphismTwisted(pt); !passesTwisted {
			b.Fatalf("Pt does not pass BLS12-381 endomorphism check")
		}
		// Verify that Pt is NOT on the original BLS12-381 curve
		if isOnBLS := isPointOnBLS12381(pt.x, pt.y); isOnBLS {
			b.Fatalf("Pt is on original BLS12-381 curve")
		}
	}
}
