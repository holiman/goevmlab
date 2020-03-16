// Copyright 2019 Martin Holst Swende
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

// These tests are all commented out, since they're only useful if you want to
// see the values. Not suitable for automated testing, as they don't ever fail

/*

func TestValueGen(t *testing.T) {
	gen := ValueRandomizer()
	for i := 0; i < 100; i++ {
		fmt.Printf("%x\n", gen())
	}
}
func TestMemGen(t *testing.T) {
	memFn := MemRandomizer()
	for i := 0; i < 100; i++ {
		loc, size := memFn()
		fmt.Printf("%v %v\n", loc, size)
	}
}
func TestGasGen(t *testing.T) {
	gen := GasRandomizer()
	for i := 0; i < 100; i++ {
		fmt.Printf("%x\n", gen())
	}
}
func TestRandCall(t *testing.T) {
	addrGen := addressRandomizer([]common.Address{
		common.HexToAddress("0x1337"), common.HexToAddress("0x1338"),
	})
	memFn := MemRandomizer()
	fmt.Printf("%x\n", RandCall(GasRandomizer(), addrGen, ValueRandomizer(), memFn, memFn))
}
*/
