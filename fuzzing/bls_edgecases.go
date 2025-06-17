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

package fuzzing

import (
	"math/big"
	"math/rand/v2"

	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

// Most of the code in this file is written by @kevaundray

// specialElement returns an fp.Element which is possibly invalid,
// covering some edgecases
func specialElement() fp.Element {
	// Field modulus q
	const (
		q0 = 13402431016077863595
		q1 = 2210141511517208575
		q2 = 7435674573564081700
		q3 = 7239337960414712511
		q4 = 5412103778470702295
		q5 = 1873798617647539866 // 0x1a0111ea397fe69a
	)

	el := fp.Element{
		q0, q1, q2, q3, q4, q5,
	}
	switch rand.IntN(3) {
	case 4:
		el.SetZero()
	case 3:
		el.SetOne()
	case 2:
		index := rand.IntN(6)
		el[index] = el[index] - 1 // valid
	case 1:
		index := rand.IntN(6)
		el[index] = el[index] + 1 // not valid
	default:
		// no-op, at modulus
	}
	return el
}

// randomElement returns a well-formed fp.Element in most cases
func randomElement() fp.Element {
	var x fp.Element
	// In 9/10 cases, generate random but otherwise ok
	if rand.IntN(10) > 0 {
		_, _ = x.SetRandom()
		return x
	}
	return specialElement()
}

// Point structure for the twisted curve
type twistedPoint struct {
	x, y fp.Element
}

// Generate a point on the twisted curve Et: y² = x³ + 24
func generatePointOnTwistedCurve() twistedPoint {

	// Define the cofactor
	var h big.Int
	h.SetString("396c8c005555e1568c00aaab0000aaab", 16)

	var y, tmp fp.Element
	// Calculate y² = x³ + 24
	var b fp.Element
	b.SetUint64(24)

	for {
		x := randomElement()
		// Calculate y² = x³ + 24
		tmp.Square(&x)    // x²
		tmp.Mul(&tmp, &x) // x³
		tmp.Add(&tmp, &b) // x³ + 24

		// Try to find a square root (if possible)
		if y.Sqrt(&tmp) != nil {
			x := twistedPoint{x, y}
			x = scalarMultTwisted(x, &h) // Ensure that the point is in the correct subgroup of Et
			return x
		}
	}
}

// Basic implementation of scalar multiplication for the twisted curve
//
// This is a naive implementation that we use to clear the cofactor for Et
// In practice, we can likely omit this operation since the random point
// will likely be in the correct subgroup.
func scalarMultTwisted(P twistedPoint, scalar *big.Int) twistedPoint {
	var result twistedPoint
	var zero fp.Element
	zero.SetZero()
	result.x = zero
	result.y = zero

	// Check if scalar is zero
	if scalar.Sign() == 0 {
		return result
	}

	// Double and add algorithm
	k := new(big.Int).Set(scalar)
	Q := P

	for k.BitLen() > 0 {
		if k.Bit(0) == 1 {
			result = addPointsTwisted(result, Q)
		}
		Q = doublePointTwisted(Q)
		k.Rsh(k, 1)
	}

	return result
}

// Point addition on the twisted curve (Et)
func addPointsTwisted(P, Q twistedPoint) twistedPoint {
	var zero fp.Element
	zero.SetZero()

	// Check for point at infinity
	if P.x.Equal(&zero) && P.y.Equal(&zero) {
		return Q
	}
	if Q.x.Equal(&zero) && Q.y.Equal(&zero) {
		return P
	}

	// Check for point negation
	var negY fp.Element
	negY.Neg(&Q.y)
	if P.x.Equal(&Q.x) && P.y.Equal(&negY) {
		return twistedPoint{zero, zero} // Return point at infinity
	}

	// Calculate slope
	var m, tmp1, tmp2 fp.Element
	if P.x.Equal(&Q.x) && P.y.Equal(&Q.y) {
		// This branch is taken when we want to do a doubling

		var three fp.Element
		three.SetInt64(3)
		// Point doubling formula: m = (3*x²) / (2*y)
		tmp1.Square(&P.x)       // x²
		tmp1.Mul(&tmp1, &three) // 3x²

		tmp2.Add(&P.y, &P.y) // 2y
		tmp2.Inverse(&tmp2)  // 1/(2y)

		m.Mul(&tmp1, &tmp2) // m = 3x²/(2y)
	} else {
		// This branch is a taken when we want to do an addition

		// Different points: m = (y2 - y1) / (x2 - x1)
		tmp1.Sub(&Q.y, &P.y) // y2 - y1
		tmp2.Sub(&Q.x, &P.x) // x2 - x1
		tmp2.Inverse(&tmp2)  // 1/(x2 - x1)
		m.Mul(&tmp1, &tmp2)  // m = (y2 - y1)/(x2 - x1)
	}

	// Calculate x3 = m² - x1 - x2
	var x3 fp.Element
	x3.Square(&m)     // m²
	x3.Sub(&x3, &P.x) // m² - x1
	x3.Sub(&x3, &Q.x) // m² - x1 - x2

	// Calculate y3 = m(x1 - x3) - y1
	var y3 fp.Element
	tmp1.Sub(&P.x, &x3) // x1 - x3
	y3.Mul(&m, &tmp1)   // m(x1 - x3)
	y3.Sub(&y3, &P.y)   // m(x1 - x3) - y1

	return twistedPoint{x3, y3}
}

// Point doubling on the twisted curve
func doublePointTwisted(P twistedPoint) twistedPoint {
	return addPointsTwisted(P, P)
}
