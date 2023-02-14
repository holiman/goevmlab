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

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/holiman/goevmlab/evms"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var binMu sync.Mutex

// We need to decompress the evm binaries prior to execution
func setupBinaries(t *testing.T) {
	binMu.Lock()
	defer binMu.Unlock()

	err := filepath.Walk("../binaries", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".zip") {
			// ignore
			return nil
		}
		destPath := strings.TrimSuffix(path, ".zip")
		if _, err := os.Stat(destPath); err == nil {
			// Already exists, skip
			return nil
		}
		var src io.Reader
		fmt.Printf("Uncompressing %v into %v \n", path, destPath)
		{ // Decompress
			zr, err := zip.OpenReader(path)
			if err != nil {
				return fmt.Errorf("failed opening reader: %v", err)
			}
			f, err := zr.File[0].Open()
			if err != nil {
				return fmt.Errorf("failed opening file inside archive: %v", err)
			}
			defer f.Close()
			src = f
		}
		dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0777)
		if err != nil {
			return err
		}
		defer dest.Close()
		size, err := io.Copy(dest, src)
		if err != nil {
			return err
		}
		fmt.Printf("Uncompressed %v OK (%d bytes)\n", path, size)
		return nil
	})
	if err != nil {
		t.Fatalf("error constructing evm binaries: %v", err)
	}
}

func testCompare(a, b evms.Evm, testfile string) error {

	wa := bytes.NewBuffer(nil)
	if _, err := a.RunStateTest(testfile, wa, false); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	wb := bytes.NewBuffer(nil)
	if _, err := b.RunStateTest(testfile, wb, false); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}
	eq, _ := evms.CompareFiles([]evms.Evm{a, b}, []io.Reader{wa, wb})
	if !eq {
		return fmt.Errorf("diffs encountered")
	}
	return nil
}
