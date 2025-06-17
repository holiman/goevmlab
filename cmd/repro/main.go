// Copyright 2023 Martin Holst Swende
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
	"strconv"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/program"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/goevmlab/common"
	"github.com/holiman/goevmlab/traces"
	"github.com/urfave/cli/v2"
)

func initApp() *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Authors = []*cli.Author{{Name: "Martin Holst Swende"}}
	app.Usage = "Test-case creator"
	app.Flags = append(app.Flags, common.VMFlags...)
	app.Action = startAnalysis
	return app
}

var app = initApp()

func main() {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stderr, log.LevelInfo, true)))
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startAnalysis(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("input trace file needed")
	}
	traces, err := traces.ReadFile(c.Args().First())
	if err != nil {
		return err
	}
	log.Info("Read traces", "steps", len(traces.Ops))
	step := traces.Ops[len(traces.Ops)-1].Step()
	if c.NArg() > 1 {
		if x, err := strconv.Atoi(c.Args().Get(1)); err != nil {
			return err
		} else {
			step = uint64(x)
		}
	}
	log.Info("Read traces", "steps", len(traces.Ops), "target step", step)
	// We have a target to hone in on.
	traceLine := traces.Ops[step]
	fmt.Printf("Target line:\n\n\t%v\n\n", traceLine.Source())
	fmt.Printf("Reproing stack, at step %d: \n", traceLine.Step())
	stack := traceLine.Stack()
	p := program.New()
	for i := len(stack) - 1; i >= 0; i-- {
		v := stack[i]
		fmt.Printf("\t push %v\n", v.Hex())
		p.Push(v)
	}
	operation := vm.OpCode(traceLine.Op())
	fmt.Printf("Adding op\n\t%v\n", operation)
	p.Op(operation)
	fmt.Printf("Code: %#x\n", p.Bytes())
	return nil
}
