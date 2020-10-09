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
	"strings"

	"github.com/holiman/goevmlab/common"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = "Martin Holst Swende"
	app.Usage = "Tests execution speed on list of statetests"
	app.Flags = append(app.Flags, common.VmFlags...)
	app.Action = startTests
	return app
}

var app = initApp()

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startTests(c *cli.Context) error {

	if c.NArg() != 1 {
		return fmt.Errorf("input state test directory needed")
	}
	dir := c.Args()[0]
	finfo, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !finfo.IsDir() {
		return fmt.Errorf("%v is not a directory", dir)
	}
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, "json") {
			return nil
		}
		if err != nil {
			return err
		}
		slow, err := common.TestSpeed(path, c)
		if err != nil {
			return err
		}
		if !slow {
			fmt.Printf("deleting %v\n", path)
			return os.Remove(path)
		}
		return nil

	})
}
