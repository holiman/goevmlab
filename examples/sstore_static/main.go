package main

import (
	"encoding/json"
	"fmt"
	//"github.com/holiman/goevmlab/evms"
	"github.com/holiman/goevmlab/fuzzing"
	"os"
	"path"
)

func main() {

	gstMaker := fuzzing.GenerateSSToreStatic()
	gst := gstMaker.ToGeneralStateTest("foobar")
	testName := "sstore_test-3"
	fileName := fmt.Sprintf("%v.json", testName)
	p := path.Join("./", fileName)
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		fmt.Sprintf("err: %x", err)
		os.Exit(1)
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", " ")
	if err = encoder.Encode(gst); err != nil {
		f.Close()
		fmt.Sprintf("err: %x", err)
		os.Exit(1)
	}
	f.Close()
	//geth := evms.NewGethEVM("../binaries/evm")
	//parity := evms.NewParityVM("../binaries/parity-evm")
	//
	//if err := testCompare(geth, parity, p); err != nil {
	//	t.Fatal(err)
	//}

}
