// Copyright 2019 Martin Holst Swende
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
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"os"
	"path/filepath"

	"github.com/holiman/goevmlab/common"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Martin Holst Swende"
	app.Usage = "Fuzzer targeting SSTORE and SLOAD"
	app.Flags = append(app.Flags, common.VmFlags...)
	app.Action = startFuzzer
	return app
}

var app = initApp()

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startFuzzer(c *cli.Context) error {

	if c.NArg() != 1 {
		return fmt.Errorf("input state test file needed")
	}
	return common.RunOneTest(c.Args().First(), c)
}
