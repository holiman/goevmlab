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
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/common"
	"github.com/urfave/cli/v2"
)

var app = initApp()

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Executes one test against several vms"
	app.Flags = append(app.Flags, common.VMFlags...)
	app.Flags = append(app.Flags, common.SkipTraceFlag)
	app.Flags = append(app.Flags, common.ThreadFlag)
	app.Flags = append(app.Flags, common.LocationFlag)
	app.Flags = append(app.Flags, common.VerbosityFlag)
	app.Action = startFuzzer
	return app
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startFuzzer(c *cli.Context) error {
	loglevel := slog.Level(c.Int(common.VerbosityFlag.Name))
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, loglevel, true)))

	if c.NArg() != 1 {
		return fmt.Errorf("file (or regexp) needed")
	}
	files, err := filepath.Glob(c.Args().First())
	if err != nil {
		return err
	}
	var nextFile atomic.Int64
	return common.ExecuteFuzzer(c, true, func(_, _ int) (string, error) {
		index := int(nextFile.Add(1)) - 1
		if index < len(files) {
			return files[index], nil
		}
		return "", io.EOF
	}, false)
}
