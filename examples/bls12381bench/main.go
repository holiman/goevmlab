package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core"
	"github.com/holiman/goevmlab/fuzzing"
	"github.com/holiman/goevmlab/ops"
	"github.com/holiman/goevmlab/program"
	"github.com/holiman/uint256"
	"github.com/urfave/cli/v2"
	"io"
	"math/big"
	"os"
	"path/filepath"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Jared Wasinger"}}
	app.Usage = "Generator for bls precompile benchmarks"
	return app
}

var (
	app            = initApp()
	precompileFlag = &cli.StringFlag{
		Name:  "precompile",
		Value: "",
		Usage: "which bls precompile to benchmark",
	}
	inputCountFlag = &cli.IntFlag{
		Name:  "input-count",
		Value: 1,
		Usage: "number of inputs to use (for pairing and msm precompiles)",
	}
	benchIterCountFlag = &cli.IntFlag{
		Name:  "iter-count",
		Value: 2850,
		Usage: "number of times to call the target benchmark contract",
	}
	evaluateCommand = &cli.Command{
		Action:      evaluate,
		Name:        "evaluate",
		Usage:       "evaluate the test using the built-in go-ethereum base",
		Description: `Evaluate the test using the built-in go-ethereum library.`,
	}
	genGethBenchmarkCommand = &cli.Command{
		Action:      genGethBenchmark,
		Name:        "genbench",
		Usage:       "...",
		Description: "...",
	}
)

func init() {
	app.Flags = []cli.Flag{
		precompileFlag,
		inputCountFlag,
		benchIterCountFlag,
	}
	app.Commands = []*cli.Command{
		evaluateCommand,
		genGethBenchmarkCommand,
	}
}

type benchmarkEntry struct {
	Input       string
	Expected    string
	Name        string
	Gas         int
	NoBenchmark bool
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func precompileNameToAddress(name string) common.Address {
	switch name {
	case "g1add":
		return common.BytesToAddress([]byte{0x0b})
	case "g1mul":
		return common.BytesToAddress([]byte{0x0c})
	case "g1msm":
		return common.BytesToAddress([]byte{0x0d})
	case "g2add":
		return common.BytesToAddress([]byte{0x0e})
	case "g2mul":
		return common.BytesToAddress([]byte{0x0f})
	case "g2msm":
		return common.BytesToAddress([]byte{0x10})
	case "pairing":
		return common.BytesToAddress([]byte{0x11})
	case "mapfp":
		return common.BytesToAddress([]byte{0x12})
	case "mapfp2":
		return common.BytesToAddress([]byte{0x13})
	default:
		panic(fmt.Sprintf("invalid precompile selection", name))
	}
}

func genG1MSMBenchmarks() {
	var benchmarks []benchmarkEntry
	precompileAddr := precompileNameToAddress("g1msm")
	for i := 1; i < 32; i++ {
		benchmarks = append(benchmarks, generateNativeBench(rand.Reader, precompileAddr, i))
	}
	largeSizes := []int{64, 128, 256, 512, 1024, 2048, 4877}
	for _, size := range largeSizes {
		benchmarks = append(benchmarks, generateNativeBench(rand.Reader, precompileAddr, size))
	}

	f, err := os.OpenFile("g1msm-benchmarks.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 777)
	if err != nil {
		panic(err)
	}

	enc, err := json.Marshal(benchmarks)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(enc)
	if err != nil {
		panic(err)
	}
}

func genG2MSMBenchmarks() {
	var benchmarks []benchmarkEntry
	precompileAddr := precompileNameToAddress("g2msm")
	for i := 1; i < 32; i++ {
		benchmarks = append(benchmarks, generateNativeBench(rand.Reader, precompileAddr, i))
	}
	largeSizes := []int{64, 128, 256, 512, 1024}
	for _, size := range largeSizes {
		benchmarks = append(benchmarks, generateNativeBench(rand.Reader, precompileAddr, size))
	}

	f, err := os.OpenFile("g2msm-benchmarks.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 777)
	if err != nil {
		panic(err)
	}

	enc, err := json.Marshal(benchmarks)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(enc)
	if err != nil {
		panic(err)
	}
}

func genPairingBenchmarks() {
	var benchmarks []benchmarkEntry
	precompileAddr := precompileNameToAddress("pairing")
	for i := 1; i < 9; i++ {
		benchmarks = append(benchmarks, generateNativeBench(rand.Reader, precompileAddr, i))
	}

	f, err := os.OpenFile("pairing-benchmarks.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 777)
	if err != nil {
		panic(err)
	}

	enc, err := json.Marshal(benchmarks)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(enc)
	if err != nil {
		panic(err)
	}
}

func genMapFp2Benchmarks() {
	var benchmarks []benchmarkEntry
	precompileAddr := precompileNameToAddress("mapfp2")
	benchmarks = append(benchmarks, generateNativeBench(rand.Reader, precompileAddr, 1))

	f, err := os.OpenFile("mapfp2-benchmarks.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 777)
	if err != nil {
		panic(err)
	}

	enc, err := json.Marshal(benchmarks)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(enc)
	if err != nil {
		panic(err)
	}
}

func genGethBenchmark(ctx *cli.Context) error {
	genG1MSMBenchmarks()
	genG2MSMBenchmarks()
	genPairingBenchmarks()
	genMapFp2Benchmarks()
	return nil
}

func evaluate(ctx *cli.Context) error {
	var (
		precompileName = ctx.String(precompileFlag.Name)
		precompile     = precompileNameToAddress(precompileName)
		inputCount     = ctx.Int(inputCountFlag.Name)
		iterCount      = ctx.Int(benchIterCountFlag.Name)
	)
	alloc := generateAlloc(iterCount, false)
	input := generateBenchInputs(rand.Reader, precompile, inputCount)
	benchName := fmt.Sprintf("bench-%s-%d", precompileName, inputCount)
	if err := convertToStateTest(benchName, "Prague", alloc, 1_000_000_000, benchContractAddr, input); err != nil {
		return err
	}
	noopAlloc := generateAlloc(iterCount, true)
	noopBenchName := fmt.Sprintf("noop-%s-%d", precompileName, inputCount)
	if err := convertToStateTest(noopBenchName, "Prague", noopAlloc, 1_000_000_000, benchContractAddr, input); err != nil {
		return err
	}
	return nil
}

func generateBenchCode(iterCount int, isNoop bool) *program.Program {
	benchCode := program.NewProgram()
	benchCode.CalldataLoad(0)

	benchCode.Op(ops.DUP1)
	benchCode.Push(uint256.MustFromHex("0xffffffffffffffffffffffffffffffff00000000000000000000000000000000"))
	benchCode.Op(ops.AND)
	benchCode.Push(0x80)
	benchCode.Op(ops.SHR)
	benchCode.Op(ops.SWAP1)
	// stack: calldata[0], input_size

	benchCode.Push(uint256.MustFromHex("0xffffffffffffffffffffffffffffffff"))
	benchCode.Op(ops.AND)
	//stack: output_size, input_size

	// mem[0:input_size+output_size] <- calldatacopy(calldata[32:32+input_size+output_size])
	benchCode.Op(ops.DUP1)
	benchCode.Op(ops.DUP3)
	benchCode.Op(ops.ADD)
	benchCode.Push(0x34)
	benchCode.Push(0)
	benchCode.Op(ops.CALLDATACOPY)

	benchCode.Push(0x20)
	benchCode.Op(ops.CALLDATALOAD)
	benchCode.Push(0x60)
	benchCode.Op(ops.SHR)
	//stack: precompile_address, output_size, input_size

	benchCode.Push(uint256.MustFromHex("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00"))
	benchCode.Op(ops.DUP4)

	for i := 0; i < iterCount; i++ {
		if !isNoop {
			benchCode.Op(ops.DUP4)
			benchCode.Op(ops.DUP2)
			benchCode.Op(ops.DUP7)
			benchCode.Push(0)
			benchCode.Op(ops.DUP7)
			benchCode.Op(ops.GASLIMIT)
			benchCode.Op(ops.STATICCALL)
		} else {
			benchCode.Push(0)
			benchCode.Op(ops.DUP1)
			benchCode.Op(ops.DUP1)
			benchCode.Op(ops.DUP1)
			benchCode.Push(4)
			benchCode.Op(ops.GASLIMIT)
			benchCode.Op(ops.STATICCALL)
		}
		benchCode.Op(ops.POP)
	}
	benchCode.Op(ops.POP)
	// stack: loop counter, precompile address, output size, input size
	/*
		benchCode.Op(ops.DUP4)
		benchCode.Op(ops.DUP4)
		benchCode.Op(ops.SWAP1)
		benchCode.Op(ops.RETURN)
	*/
	return benchCode
}
func randomG1Point(input io.Reader) *bls12381.G1Affine {
	// sample a random scalar
	s := randomScalar(input)

	// compute a random point
	pt := new(bls12381.G1Affine)
	_, _, g1Gen, _ := bls12381.Generators()
	pt.ScalarMultiplication(&g1Gen, s)

	return pt
}

func randomG2Point(input io.Reader) *bls12381.G2Affine {
	// sample a random scalar
	s := randomScalar(input)

	// compute a random point
	pt := new(bls12381.G2Affine)
	_, _, _, g2Gen := bls12381.Generators()
	pt.ScalarMultiplication(&g2Gen, s)
	return pt
}

func randomScalar(r io.Reader) (k *big.Int) {
	k, _ = rand.Int(r, math.MaxBig256)
	return k
}

func marshalFr(elem *fr.Element) []byte {
	elemBytes := elem.Bytes()
	return elemBytes[:]
}

// marshal 32 bit scalar
func marshalScalar(k *big.Int) (res []byte) {
	kBytes := k.Bytes()
	res = make([]byte, 32)
	copy(res[32-len(kBytes):], kBytes)
	return res
}

func randomFp(r io.Reader) fp.Element {
	randFp, _ := rand.Int(r, fp.Modulus())
	scalar := new(fp.Element)
	scalar.SetBytes(randFp.Bytes())
	return *scalar
}

func randomFr(r io.Reader) fr.Element {
	randFr, _ := rand.Int(r, fr.Modulus())
	scalar := new(fr.Element)
	scalar.SetBytes(randFr.Bytes())
	return *scalar
}

// returns fr with all ones:
func highArityFr() fr.Element {
	scalar := new(fr.Element)
	_, err := scalar.SetString("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	if err != nil {
		panic(err)
	}
	return *scalar
}

func marshalFp(elem fp.Element) []byte {
	res := make([]byte, 64)
	copy(res[16:], elem.Marshal())
	return res
}

func marshalG1Point(pt *bls12381.G1Affine) []byte {
	resX := marshalFp(pt.X)
	resY := marshalFp(pt.Y)

	return append(resX, resY...)
}

func marshalG2Point(pt *bls12381.G2Affine) []byte {
	resX_0 := marshalFp(pt.X.A0)
	resX_1 := marshalFp(pt.X.A1)
	resY_0 := marshalFp(pt.Y.A0)
	resY_1 := marshalFp(pt.Y.A1)
	res := append(resX_0, resX_1...)
	res = append(res, resY_0...)
	res = append(res, resY_1...)
	return res
}

func genG1MSMInputs(r io.Reader, inputCount int) (points []bls12381.G1Affine, scalars []fr.Element, encoded []byte) {
	for i := 0; i < inputCount; i++ {
		point := randomG1Point(r)
		scalar := randomFr(r) //highArityFr()

		points = append(points, *point)
		scalars = append(scalars, scalar)

		encoded = append(encoded, marshalG1Point(point)...)
		encoded = append(encoded, marshalFr(&scalar)...)
	}
	return points, scalars, encoded
}

func genG2MSMInputs(r io.Reader, inputCount int) (points []bls12381.G2Affine, scalars []fr.Element, encoded []byte) {
	for i := 0; i < inputCount; i++ {
		point := randomG2Point(r)
		scalar := highArityFr()

		points = append(points, *point)
		scalars = append(scalars, scalar)
		encoded = append(encoded, marshalG2Point(point)...)
		encoded = append(encoded, marshalFr(&scalar)...)
	}
	return points, scalars, encoded
}

func genPairingInputs(r io.Reader, inputCount int) (g1Points []bls12381.G1Affine, g2Points []bls12381.G2Affine, encoded []byte) {
	for i := 0; i < inputCount; i++ {
		g1Point := randomG1Point(r)
		g2Point := randomG2Point(r)
		g1Points = append(g1Points, *g1Point)
		g2Points = append(g2Points, *g2Point)

		encoded = append(encoded, marshalG1Point(g1Point)...)
		encoded = append(encoded, marshalG2Point(g2Point)...)
	}
	return g1Points, g2Points, encoded
}

func encodeU128(val uint64) []byte {
	res := make([]byte, 16)
	binary.BigEndian.PutUint64(res[8:16], val)
	return res
}

func generateNativeBench(r io.Reader, precompile common.Address, inputCount int) benchmarkEntry {
	var input []byte
	var output []byte
	var precompileName string

	switch precompile {
	case common.BytesToAddress([]byte{0x0b}):
		precompileName = "g1add"
		_, _, g1Gen, _ := bls12381.Generators()
		pt1 := new(bls12381.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(2))
		pt2 := g1Gen
		input = append(input, marshalG1Point(pt1)...)
		input = append(input, marshalG1Point(&pt2)...)
		output = marshalG1Point(new(bls12381.G1Affine).Add(pt1, &pt2))
	case common.BytesToAddress([]byte{0x0c}): // g1 mul
		precompileName = "g1mul"
		_, _, g1Gen, _ := bls12381.Generators()
		highArityScalar := new(big.Int)
		highArityScalar.SetString("50597600879605352240557443896859274688352069811191692694697732254669473040618", 10)
		input = append(input, marshalG1Point(&g1Gen)...)
		input = append(input, marshalScalar(highArityScalar)...)

		output = marshalG1Point(g1Gen.ScalarMultiplication(&g1Gen, highArityScalar))
	case common.BytesToAddress([]byte{0x0d}): // g1 msm
		precompileName = "g1msm"
		points, scalars, encInput := genG1MSMInputs(r, inputCount)
		outputPt, err := new(bls12381.G1Affine).MultiExp(points, scalars, ecc.MultiExpConfig{})
		if err != nil {
			panic(err)
		}
		output = marshalG1Point(outputPt)
		input = append(input, encInput...)
	case common.BytesToAddress([]byte{0x0e}): // g2 add
		precompileName = "g2add"
		_, _, _, g2Gen := bls12381.Generators()
		pt1 := new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, big.NewInt(2))
		pt2 := g2Gen
		input = append(input, marshalG2Point(pt1)...)
		input = append(input, marshalG2Point(&pt2)...)
		output = marshalG2Point(new(bls12381.G2Affine).Add(pt1, &pt2))
	case common.BytesToAddress([]byte{0x0f}): // g2 mul
		precompileName = "g2mul"
		_, _, _, g2Gen := bls12381.Generators()
		highArityScalar := new(big.Int)
		highArityScalar.SetString("50597600879605352240557443896859274688352069811191692694697732254669473040618", 10)
		input = append(input, marshalG2Point(&g2Gen)...)
		input = append(input, marshalScalar(highArityScalar)...)
		output = marshalG2Point(new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, highArityScalar))
	case common.BytesToAddress([]byte{0x10}): // g2 msm
		precompileName = "g2msm"
		points, scalars, encodedInput := genG2MSMInputs(r, inputCount)
		outputPt, err := new(bls12381.G2Affine).MultiExp(points, scalars, ecc.MultiExpConfig{})
		if err != nil {
			panic(err)
		}
		output = marshalG2Point(outputPt)
		input = append(input, encodedInput...)
	case common.BytesToAddress([]byte{0x11}): // pairing check
		precompileName = "pairing"
		g1Points, g2Points, encInput := genPairingInputs(r, inputCount)
		ok, err := bls12381.PairingCheck(g1Points, g2Points)
		if err != nil {
			panic(err)
		}
		output = make([]byte, 32)
		if ok {
			output[31] = 1
		}
		input = append(input, encInput...)
	case common.BytesToAddress([]byte{0x12}): // MapFp
		precompileName = "mapfp"
		elem := randomFp(r)
		outputPt := bls12381.MapToG1(elem)
		output = marshalG1Point(&outputPt)
		input = append(input, marshalFp(elem)...)
	case common.BytesToAddress([]byte{0x13}): // MapFp2
		precompileName = "mapfp2"
		elem := bls12381.E2{A0: randomFp(r), A1: randomFp(r)}
		outputPt := bls12381.MapToG2(elem)
		output = marshalG2Point(&outputPt)

		input = append(input, marshalFp(elem.A0)...)
		input = append(input, marshalFp(elem.A1)...)
	}

	return benchmarkEntry{
		Input:       fmt.Sprintf("%x", input),
		Expected:    fmt.Sprintf("%x", output),
		Name:        fmt.Sprintf("%s-%d-jwasinger", precompileName, inputCount),
		NoBenchmark: false,
		Gas:         0, // placeholder to be manually filled in later
	}
}

func generateBenchInputs(r io.Reader, precompile common.Address, inputCount int) []byte {
	var res []byte
	var precompileInput []byte
	var precompileOutput []byte

	switch precompile {
	case common.BytesToAddress([]byte{0x0b}):
		// g1 add
		res = append(res, encodeU128(2*128)...) // input size
		res = append(res, encodeU128(128)...)   // output size
		_, _, g1Gen, _ := bls12381.Generators()
		pt1 := new(bls12381.G1Affine).ScalarMultiplication(&g1Gen, big.NewInt(2))
		pt2 := g1Gen
		precompileInput = append(precompileInput, marshalG1Point(pt1)...)
		precompileInput = append(precompileInput, marshalG1Point(&pt2)...)
		precompileOutput = marshalG1Point(new(bls12381.G1Affine).Add(pt1, &pt2))
	case common.BytesToAddress([]byte{0x0c}): // g1 mul
		res = append(res, encodeU128(128+32)...) // input size
		res = append(res, encodeU128(128)...)    // output size
		_, _, g1Gen, _ := bls12381.Generators()
		highArityScalar := new(big.Int)
		highArityScalar.SetString("50597600879605352240557443896859274688352069811191692694697732254669473040618", 10)
		precompileInput = append(precompileInput, marshalG1Point(&g1Gen)...)
		precompileInput = append(precompileInput, marshalScalar(highArityScalar)...)

		precompileOutput = marshalG1Point(g1Gen.ScalarMultiplication(&g1Gen, highArityScalar))
	case common.BytesToAddress([]byte{0x0d}): // g1 msm
		res = append(res, encodeU128(uint64(inputCount)*(128+32))...) // input size
		res = append(res, encodeU128(128)...)                         // output size

		points, scalars, encInput := genG1MSMInputs(r, inputCount)
		output, err := new(bls12381.G1Affine).MultiExp(points, scalars, ecc.MultiExpConfig{})
		if err != nil {
			panic(err)
		}
		precompileOutput = marshalG1Point(output)
		precompileInput = append(precompileInput, encInput...)
	case common.BytesToAddress([]byte{0x0e}): // g2 add
		res = append(res, encodeU128(2*256)...) // input size
		res = append(res, encodeU128(256)...)   // output size
		_, _, _, g2Gen := bls12381.Generators()
		pt1 := new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, big.NewInt(2))
		pt2 := g2Gen
		precompileInput = append(precompileInput, marshalG2Point(pt1)...)
		precompileInput = append(precompileInput, marshalG2Point(&pt2)...)
		precompileOutput = marshalG2Point(new(bls12381.G2Affine).Add(pt1, &pt2))
	case common.BytesToAddress([]byte{0x0f}): // g2 mul
		res = append(res, encodeU128(256+32)...) // input size
		res = append(res, encodeU128(256)...)    // output size
		_, _, _, g2Gen := bls12381.Generators()
		highArityScalar := new(big.Int)
		highArityScalar.SetString("50597600879605352240557443896859274688352069811191692694697732254669473040618", 10)
		precompileInput = append(precompileInput, marshalG2Point(&g2Gen)...)
		precompileInput = append(precompileInput, marshalScalar(highArityScalar)...)
		precompileOutput = marshalG2Point(new(bls12381.G2Affine).ScalarMultiplication(&g2Gen, highArityScalar))
	case common.BytesToAddress([]byte{0x10}): // g2 msm
		res = append(res, encodeU128(uint64(inputCount)*(256+32))...) // input size
		res = append(res, encodeU128(128)...)                         // output size
		points, scalars, encodedInput := genG2MSMInputs(r, inputCount)
		output, err := new(bls12381.G2Affine).MultiExp(points, scalars, ecc.MultiExpConfig{})
		if err != nil {
			panic(err)
		}
		precompileOutput = marshalG2Point(output)
		precompileInput = append(precompileInput, encodedInput...)
	case common.BytesToAddress([]byte{0x11}): // pairing check
		res = append(res, encodeU128(uint64(inputCount)*(256+128))...) // input size
		res = append(res, encodeU128(32)...)                           // output size
		g1Points, g2Points, encInput := genPairingInputs(r, inputCount)
		ok, err := bls12381.PairingCheck(g1Points, g2Points)
		if err != nil {
			panic(err)
		}
		precompileOutput = make([]byte, 32)
		if ok {
			precompileOutput[31] = 1
		}
		precompileInput = append(precompileInput, encInput...)
	case common.BytesToAddress([]byte{0x12}): // MapFp
		res = append(res, encodeU128(64)...)  // input size
		res = append(res, encodeU128(128)...) // output size
		elem := randomFp(r)
		output := bls12381.MapToG1(elem)
		precompileOutput = marshalG1Point(&output)
		precompileInput = append(precompileInput, marshalFp(elem)...)
	case common.BytesToAddress([]byte{0x13}): // MapFp2
		res = append(res, encodeU128(128)...) // input size
		res = append(res, encodeU128(256)...) // output size
		elem := bls12381.E2{A0: randomFp(r), A1: randomFp(r)}
		output := bls12381.MapToG2(elem)
		precompileOutput = marshalG2Point(&output)

		precompileInput = append(precompileInput, marshalFp(elem.A0)...)
		precompileInput = append(precompileInput, marshalFp(elem.A1)...)
	}
	res = append(res, precompile.Bytes()...)
	res = append(res, precompileInput...)
	// TODO: check output explicitly
	//res = append(res, precompileOutput...)
	return res
}

var benchContractAddr = common.HexToAddress("0xdeadbeef")

func generateAlloc(benchIterCount int, isNoop bool) core.GenesisAlloc {
	var alloc core.GenesisAlloc
	benchCode := generateBenchCode(benchIterCount, isNoop).Bytecode()
	alloc = make(core.GenesisAlloc)
	alloc[benchContractAddr] = core.GenesisAccount{
		Code: benchCode,
	}
	return alloc
}

// convertToStateTest is a utility to turn stuff into sharable state tests.
func convertToStateTest(name, fork string, alloc core.GenesisAlloc, gasLimit uint64,
	target common.Address, txData []byte) error {

	mkr := fuzzing.BasicStateTest(fork)
	// convert the genesisAlloc
	var fuzzGenesisAlloc = make(fuzzing.GenesisAlloc)
	for k, v := range alloc {
		fuzzAcc := fuzzing.GenesisAccount{
			Code:       v.Code,
			Storage:    v.Storage,
			Balance:    v.Balance,
			Nonce:      v.Nonce,
			PrivateKey: v.PrivateKey,
		}
		if fuzzAcc.Balance == nil {
			fuzzAcc.Balance = new(big.Int)
		}
		if fuzzAcc.Storage == nil {
			fuzzAcc.Storage = make(map[common.Hash]common.Hash)
		}
		fuzzGenesisAlloc[k] = fuzzAcc
	}
	// Also add the sender
	var sender = common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b")
	if _, ok := fuzzGenesisAlloc[sender]; !ok {
		maxBalance := new(big.Int)
		maxBalance.SetUint64(math.MaxUint64)
		fuzzGenesisAlloc[sender] = fuzzing.GenesisAccount{
			Balance: maxBalance,
			Nonce:   0,
			Storage: make(map[common.Hash]common.Hash),
		}
	}

	tx := &fuzzing.StTransaction{
		GasLimit:   []uint64{gasLimit},
		Nonce:      0,
		Value:      []string{"0x0"},
		Data:       []string{fmt.Sprintf("0x%x", txData)},
		GasPrice:   big.NewInt(0x10),
		Sender:     sender,
		PrivateKey: hexutil.MustDecode("0x45a915e4d060149eb4365960e6a7a45f334393093061116b197e3240065ff2d8"),
		To:         target.Hex(),
	}
	mkr.SetTx(tx)
	mkr.SetPre(&fuzzGenesisAlloc)
	if err := mkr.Fill(io.Discard); err != nil {
		return err
	}
	gst := mkr.ToGeneralStateTest(name)
	dat, _ := json.MarshalIndent(gst, "", " ")
	fname := fmt.Sprintf("benchmarks/%v.json", name)
	if err := os.WriteFile(fname, dat, 0777); err != nil {
		return err
	}
	fmt.Printf("Wrote file %v\n", fname)
	return nil
}
