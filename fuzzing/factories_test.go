// Copyright 2024 Martin Holst Swende, Marius van der Wijden
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
package fuzzing

import (
	"testing"
	"time"
)

// TestGenerateStatetests is a sanity-check that the test-generators do not croak
func TestGenerateStatetests(t *testing.T) {
	for _, name := range FactoryNames() {
		factory := Factory(name, "Cancun")
		t0 := time.Now()
		for i := 0; i < 10; i++ {
			if res := factory(); res == nil {
				t.Fatalf("factory %v failed generating test (attempt %d)", name, i)
			}
		}
		t.Logf("Engine %v took %v", name, time.Since(t0))
	}
}
