// This code is a slightly modified version of the mutator used in the go-fuzz-project.
// Copyright 2015 go-fuzz project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in https://github.com/dvyukov/go-fuzz/blob/master/LICENSE.

package fuzzing

import (
	"encoding/binary"
	"math/rand"
	"strconv"
)

var (
	interesting8  = []int8{-128, -1, 0, 1, 16, 32, 64, 100, 127}
	interesting16 = []int16{-32768, -129, 128, 255, 256, 512, 1000, 1024, 4096, 32767}
	interesting32 = []int32{-2147483648, -100663046, -32769, 32768, 65535, 65536, 100663045, 2147483647}
)

func init() {
	for _, v := range interesting8 {
		interesting16 = append(interesting16, int16(v))
	}
	for _, v := range interesting16 {
		interesting32 = append(interesting32, int32(v))
	}
}

type MutationConfig struct {
	bin          bool
	corpus       [][]byte
	MaxInputSize int
	// Literals
	lits    [][]byte
	strLits []string
}

func Mutate(input []byte, config MutationConfig) []byte {
	res := make([]byte, len(input))
	copy(res, input)
	iterations := rand.Intn(10)
	upper := 20
	if config.bin {
		// Last 4 transformations are not suitable for binary data
		upper = 16
	}
	for iter := 0; iter < iterations; iter++ {
		switch rand.Intn(upper) {
		case 0:
			res = removeRange(res)
		case 1:
			res = insertRange(res)
		case 2:
			res = duplicateRange(res)
		case 3:
			res = copyRange(res)
		case 4:
			res = bitFlip(res)
		case 5:
			res = randomBit(res)
		case 6:
			res = swapBit(res)
		case 7:
			res = addSubByte(res)
		case 8:
			res = addSubUint16(res)
		case 9:
			res = addSubUint32(res)
		case 10:
			res = addSubUint64(res)
		case 11:
			res = replaceInterestingByte(res)
		case 12:
			res = replaceInterestingUint16(res)
		case 13:
			res = replaceInterestingUint32(res)
		case 14:
			res = spliceInput(res, config.corpus)
		case 15:
			res = insertInput(res, config.corpus)
		case 16:
			res = replaceAscii(res)
		case 17:
			res = replaceMultiAscii(res)
		case 18:
			res = insertLiteral(res, config.lits, config.strLits)
		case 19:
			res = replaceLiteral(res, config.lits, config.strLits)
		}
	}
	if len(res) > config.MaxInputSize {
		res = res[:config.MaxInputSize]
	}
	return res
}

// chooseLen chooses length of range mutation.
// It gives preference to shorter ranges.
func chooseLen(n int) int {
	switch x := rand.Intn(100); {
	case x < 90:
		return rand.Intn(min(8, n)) + 1
	case x < 99:
		return rand.Intn(min(32, n)) + 1
	default:
		return rand.Intn(n) + 1
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func randbig() int64 {
	return int64(rand.Uint32() >> 2)
}

func randByteOrder() binary.ByteOrder {
	if rand.Intn(2) == 1 {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func reverse(data []byte) []byte {
	tmp := make([]byte, len(data))
	for i, v := range data {
		tmp[len(data)-i-1] = v
	}
	return tmp
}

// Transformation functions

func removeRange(res []byte) []byte {
	// Remove a range of bytes.
	if len(res) <= 1 {
		return res
	}
	pos0 := rand.Intn(len(res))
	pos1 := pos0 + chooseLen(len(res)-pos0)
	copy(res[pos0:], res[pos1:])
	return res[:len(res)-(pos1-pos0)]
}

func insertRange(res []byte) []byte {
	// Insert a range of random bytes.
	pos := rand.Intn(len(res) + 1)
	n := chooseLen(10)
	for i := 0; i < n; i++ {
		res = append(res, 0)
	}
	copy(res[pos+n:], res[pos:])
	for i := 0; i < n; i++ {
		res[pos+i] = byte(rand.Intn(256))
	}
	return res
}

func duplicateRange(res []byte) []byte {
	// Duplicate a range of bytes.
	if len(res) <= 1 {
		return res
	}
	src := rand.Intn(len(res))
	dst := rand.Intn(len(res))
	for dst == src {
		dst = rand.Intn(len(res))
	}
	n := chooseLen(len(res) - src)
	tmp := make([]byte, n)
	copy(tmp, res[src:])
	for i := 0; i < n; i++ {
		res = append(res, 0)
	}
	copy(res[dst+n:], res[dst:])
	for i := 0; i < n; i++ {
		res[dst+i] = tmp[i]
	}
	return res
}

func copyRange(res []byte) []byte {
	// Copy a range of bytes.
	if len(res) <= 1 {
		return res
	}
	src := rand.Intn(len(res))
	dst := rand.Intn(len(res))
	for dst == src {
		dst = rand.Intn(len(res))
	}
	n := chooseLen(len(res) - src)
	if dst > len(res) || src+n > len(res) {
		println(len(res), dst, src, n)
	}
	copy(res[dst:], res[src:src+n])
	return res
}

func bitFlip(res []byte) []byte {
	// Bit flip. Spooky!
	if len(res) == 0 {
		return res
	}
	pos := rand.Intn(len(res))
	res[pos] ^= 1 << uint(rand.Intn(8))
	return res
}

func randomBit(res []byte) []byte {
	// Set a byte to a random value.
	if len(res) == 0 {
		return res
	}
	pos := rand.Intn(len(res))
	res[pos] ^= byte(rand.Intn(255)) + 1
	return res
}

func swapBit(res []byte) []byte {
	// Swap 2 bytes.
	if len(res) <= 1 {
		return res
	}
	src := rand.Intn(len(res))
	dst := rand.Intn(len(res))
	for dst == src {
		dst = rand.Intn(len(res))
	}
	res[src], res[dst] = res[dst], res[src]
	return res
}

func addSubByte(res []byte) []byte {
	// Add/subtract from a byte.
	if len(res) == 0 {
		return res
	}
	pos := rand.Intn(len(res))
	v := byte(rand.Intn(35) + 1)
	if rand.Intn(2) == 1 {
		res[pos] += v
	} else {
		res[pos] -= v
	}
	return res
}

func addSubUint16(res []byte) []byte {
	// Add/subtract from a uint16.
	if len(res) < 2 {
		return res
	}
	pos := rand.Intn(len(res) - 1)
	buf := res[pos:]
	v := uint16(rand.Intn(35) + 1)
	if rand.Intn(2) == 1 {
		v = 0 - v
	}
	enc := randByteOrder()
	enc.PutUint16(buf, enc.Uint16(buf)+v)
	return res
}

func addSubUint32(res []byte) []byte {
	// Add/subtract from a uint32.
	if len(res) < 4 {
		return res
	}
	pos := rand.Intn(len(res) - 3)
	buf := res[pos:]
	v := uint32(rand.Intn(35) + 1)
	if rand.Intn(2) == 1 {
		v = 0 - v
	}
	enc := randByteOrder()
	enc.PutUint32(buf, enc.Uint32(buf)+v)
	return res
}

func addSubUint64(res []byte) []byte {
	// Add/subtract from a uint64.
	if len(res) < 8 {
		return res
	}
	pos := rand.Intn(len(res) - 7)
	buf := res[pos:]
	v := uint64(rand.Intn(35) + 1)
	if rand.Intn(2) == 1 {
		v = 0 - v
	}
	enc := randByteOrder()
	enc.PutUint64(buf, enc.Uint64(buf)+v)
	return res
}

func replaceInterestingByte(res []byte) []byte {
	// Replace a byte with an interesting value.
	if len(res) == 0 {
		return res
	}
	pos := rand.Intn(len(res))
	res[pos] = byte(interesting8[rand.Intn(len(interesting8))])
	return res
}

func replaceInterestingUint32(res []byte) []byte {
	// Replace an uint32 with an interesting value.
	if len(res) < 4 {
		return res
	}
	pos := rand.Intn(len(res) - 3)
	buf := res[pos:]
	v := uint32(interesting32[rand.Intn(len(interesting32))])
	randByteOrder().PutUint32(buf, v)
	return res
}

func replaceInterestingUint16(res []byte) []byte {
	// Replace an uint16 with an interesting value.
	if len(res) < 2 {
		return res
	}
	pos := rand.Intn(len(res) - 1)
	buf := res[pos:]
	v := uint16(interesting16[rand.Intn(len(interesting16))])
	randByteOrder().PutUint16(buf, v)
	return res
}

func replaceAscii(res []byte) []byte {
	// Replace an ascii digit with another digit.
	var digits []int
	for i, v := range res {
		if v >= '0' && v <= '9' {
			digits = append(digits, i)
		}
	}
	if len(digits) == 0 {
		return res
	}
	pos := rand.Intn(len(digits))
	was := res[digits[pos]]
	now := was
	for was == now {
		now = byte(rand.Intn(10)) + '0'
	}
	res[digits[pos]] = now
	return res
}

func replaceMultiAscii(res []byte) []byte {
	// Replace a multi-byte ASCII number with another number.
	type arange struct {
		start int
		end   int
	}
	var numbers []arange
	start := -1
	for i, v := range res {
		if (v >= '0' && v <= '9') || (start == -1 && v == '-') {
			if start == -1 {
				start = i
			} else if i == len(res)-1 {
				// At final byte of input.
				if i-start > 0 {
					numbers = append(numbers, arange{start, i + 1})
				}
			}
		} else {
			if start != -1 && i-start > 1 {
				numbers = append(numbers, arange{start, i})
				start = -1
			}
		}
	}
	if len(numbers) == 0 {
		return res
	}
	var v int64
	switch rand.Intn(4) {
	case 0:
		v = int64(rand.Intn(1000))
	case 1:
		v = randbig()
	case 2:
		v = randbig() * randbig()
	case 3:
		v = -randbig()
	}
	r := numbers[rand.Intn(len(numbers))]
	if res[r.start] == '-' {
		// If we started with a negative number, invert the sign of v.
		// The idea here is that negative numbers will mostly stay negative;
		// we only generate a negative (positive) replacement 1/4th of the time.
		v *= -1
	}
	str := strconv.FormatInt(v, 10)
	tmp := make([]byte, len(res)-(r.end-r.start)+len(str))
	copy(tmp, res[:r.start])
	copy(tmp[r.start:], str)
	copy(tmp[r.start+len(str):], res[r.end:])
	res = tmp
	return res
}

func spliceInput(res []byte, corpus [][]byte) []byte {
	// Splice another input.
	if len(res) < 4 || len(corpus) < 2 {
		return res
	}
	other := corpus[rand.Intn(len(corpus))]
	if len(other) < 4 || &res[0] == &other[0] {
		return res
	}
	// Find common prefix and suffix.
	idx0 := 0
	for idx0 < len(res) && idx0 < len(other) && res[idx0] == other[idx0] {
		idx0++
	}
	idx1 := 0
	for idx1 < len(res) && idx1 < len(other) && res[len(res)-idx1-1] == other[len(other)-idx1-1] {
		idx1++
	}
	// If diffing parts are too small, there is no sense in splicing, rely on byte flipping.
	diff := min(len(res)-idx0-idx1, len(other)-idx0-idx1)
	if diff < 4 {
		return res
	}
	copy(res[idx0:idx0+rand.Intn(diff-2)+1], other[idx0:])
	return res
}

func insertInput(res []byte, corpus [][]byte) []byte {
	// Insert a part of another input.
	if len(res) < 4 || len(corpus) < 2 {
		return res
	}
	other := corpus[rand.Intn(len(corpus))]
	if len(other) < 4 || &res[0] == &other[0] {
		return res
	}
	pos0 := rand.Intn(len(res) + 1)
	pos1 := rand.Intn(len(other) - 2)
	n := chooseLen(len(other)-pos1-2) + 2
	for i := 0; i < n; i++ {
		res = append(res, 0)
	}
	copy(res[pos0+n:], res[pos0:])
	for i := 0; i < n; i++ {
		res[pos0+i] = other[pos1+i]
	}
	return res
}

func insertLiteral(res []byte, intLits [][]byte, strLits []string) []byte {
	// Insert a literal.
	// TODO: encode int literals in big-endian, base-128, etc.
	if len(intLits) == 0 && len(strLits) == 0 {
		return res
	}
	var lit []byte
	if len(strLits) != 0 && rand.Intn(2) == 1 {
		lit = []byte(strLits[rand.Intn(len(strLits))])
	} else {
		lit = intLits[rand.Intn(len(intLits))]
		if rand.Intn(3) == 0 {
			lit = reverse(lit)
		}
	}
	pos := rand.Intn(len(res) + 1)
	for i := 0; i < len(lit); i++ {
		res = append(res, 0)
	}
	copy(res[pos+len(lit):], res[pos:])
	copy(res[pos:], lit)
	return res
}

func replaceLiteral(res []byte, intLits [][]byte, strLits []string) []byte {
	// Replace with literal.
	if len(intLits) == 0 && len(strLits) == 0 {
		return res
	}
	var lit []byte
	if len(strLits) != 0 && rand.Intn(2) == 1 {
		lit = []byte(strLits[rand.Intn(len(strLits))])
	} else {
		lit = intLits[rand.Intn(len(intLits))]
		if rand.Intn(3) == 0 {
			lit = reverse(lit)
		}
	}
	if len(lit) >= len(res) {
		return res
	}
	pos := rand.Intn(len(res) - len(lit))
	copy(res[pos:], lit)
	return res
}
