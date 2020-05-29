// Copyright 2020 Martin Holst Swende, Marius van der Wijden
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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
	"gopkg.in/urfave/cli.v1"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Martin Holst Swende"
	app.Usage = "Generator for blake (state-)tests"
	return app
}

var (
	app = initApp()
)

func init() {
	app.Flags = []cli.Flag{
		common.PrefixFlag,
		common.LocationFlag,
		common.CountFlag,
	}
	app.Commands = []cli.Command{
		generateCommand,
	}
}

var generateCommand = cli.Command{
	Action:      generate,
	Name:        "generate",
	Usage:       "generate tests",
	ArgsUsage:   "<number of tests to generate>",
	Description: `The generate generates the tests.`,
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func createTests(location, prefix string, limit int) error {
	if err := fuzzing.ReadBLSCorpus(); err != nil {
		return err
	}
	base := fuzzing.GenerateBLS()
	target := base.GetDestination()
	fmt.Printf("target: %v\n", target.Hex())

	for i := 0; i < limit; i++ {

		testName := fmt.Sprintf("%v-bls12381-test-%d", prefix, i)
		fileName := fmt.Sprintf("%v.json", testName)
		p := path.Join(location, fileName)
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		// Generate new code
		code := fuzzing.RandCallBLS()
		base.SetCode(target, code)

		// Get new state root and logs hash
		if err := base.Fill(nil); err != nil {
			f.Close()
			return err
		}

		test := base.ToGeneralStateTest(testName)
		// Write to tfile
		encoder := json.NewEncoder(f)
		encoder.SetIndent("", " ")
		if err = encoder.Encode(test); err != nil {
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

func generate(ctx *cli.Context) error {

	var prefix = ""
	if ctx.GlobalIsSet(common.PrefixFlag.Name) {
		prefix = ctx.GlobalString(common.PrefixFlag.Name)
	}
	var location = ""
	if ctx.GlobalIsSet(common.LocationFlag.Name) {
		location = ctx.GlobalString(common.LocationFlag.Name)
		if err := os.MkdirAll(location, 0755); err != nil {
			return fmt.Errorf("could not create %v: %v", location, err)
		}
	}
	var count = 0
	if ctx.GlobalIsSet(common.CountFlag.Name) {
		count = ctx.GlobalInt(common.CountFlag.Name)
	}
	return createTests(location, prefix, count)
}
