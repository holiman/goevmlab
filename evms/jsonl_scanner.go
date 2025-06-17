// Copyright 2024 Martin Holst Swende
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

package evms

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// NewJsonlScanner creates new jsonl scanner. Callers must call Release()
// after usage is done, for optimal pool usage.
func NewJsonlScanner(clientName string, input io.Reader, errOut io.Writer) *jsonlScanner {
	buf := bufferPool.Get().([]byte)
	scanner := bufio.NewScanner(input)
	scanner.Buffer(buf, 32*1024*1024)
	return &jsonlScanner{
		buf,
		clientName,
		scanner,
		errOut,
	}
}

// jsonlScanner is a scanner which can be used to parse jsonl inputs. Any non-jsonl
// errors are emitted to stdout (but parsing continue).
type jsonlScanner struct {
	buf        []byte
	clientName string
	scanner    *bufio.Scanner
	out        io.Writer
}

// Release releases the underlying scanner buffer to the pool
func (s *jsonlScanner) Release() {
	//lint:ignore SA6002: argument should be pointer-like to avoid allocations.
	bufferPool.Put(s.buf)
}

// Next parses the next line into elem. If an error is returned, then the underlying
// scanner is exhausted -- other errors are just printed to stdout, but the parsing
// continues.
func (s *jsonlScanner) Next(elem any) (err error) {
	for {
		if !s.scanner.Scan() {
			return io.EOF
		}
		data := s.scanner.Bytes()
		if len(data) == 0 {
			continue
		}
		if len(data) > 0 && data[0] == '#' {
			// Output preceded by # is ignored, but can be used for debugging, e.g.
			// to check that the generated tests cover the intended surface.
			fmt.Fprintf(s.out, "%v: %v\n", s.clientName, string(data))
			continue
		}
		var nonJSON []string
		err = json.Unmarshal(data, elem)
		for ; err != nil; err = json.Unmarshal(data, elem) {
			if len(nonJSON) == 0 { // Add first error
				title := fmt.Sprintf("%v error: %v", s.clientName, err.Error())
				nonJSON = append(nonJSON, title)
			}
			nonJSON = append(nonJSON, fmt.Sprintf("  | %v", string(data)))
			if !s.scanner.Scan() {
				break
			}
			data = s.scanner.Bytes()
		}
		if len(nonJSON) > 0 {
			fmt.Fprintln(s.out, strings.Join(nonJSON, "\n"))
		}
		return err
	}
}
