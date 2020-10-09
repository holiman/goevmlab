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
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"

	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/fuzzing"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Martin Holst Swende"
	app.Usage = "Generator for bls12-381 (state-)tests"
	app.Flags = append(app.Flags, common.VmFlags...)
	app.Flags = append(app.Flags,
		common.ThreadFlag,
		common.LocationFlag,
	)
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
	generator := func() *fuzzing.GstMaker {
		base, code := fuzzing.GenerateBLS()
		target := base.GetDestination()
		base.SetCode(target, code)
		return base
	}
	return common.ExecuteFuzzer(c, generator, "blstest")
}
