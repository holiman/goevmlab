// Copyright 2022 Martin Holst Swende
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

// Package fuzzing contains various fuzzers and utilities for generating testcases.
package fuzzing

// fillers is a mapping of names to functions that can fill a statetest.
var fillers = map[string]func(*GstMaker, string){
	"ecrecover":    fillEcRecover,
	"naive":        fillNaive,
	"blake":        fillBlake,
	"bls":          fillBls,
	"bn254":        fillBn254,
	"precompiles":  fillPrecompileTest,
	"simpleops":    fillSimple,
	"memops":       fillMemOps,
	"sstore_sload": fillSstore,
	"tstore_tload": fillTstore,
	"auth":         fill7702,
	"kzg":          fillPointEvaluation4844,
}

func Factory(name, fork string) func() *GstMaker {
	if filler, ok := fillers[name]; ok {
		return func() *GstMaker {
			gst := BasicStateTest(fork)
			filler(gst, fork)
			return gst
		}
	}
	return nil
}

// FactoryNames returns the names of the available factories
func FactoryNames() []string {
	var names []string
	for k := range fillers {
		names = append(names, k)
	}
	return names
}
