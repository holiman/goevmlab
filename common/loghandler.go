// Copyright 2026 Martin Holst Swende
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

package common

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/log"
)

// SetupLogging installs a terminal log handler with the given level. Log
// records below WARN level which originate from within go-ethereum itself are
// dropped: goevmlab uses go-ethereum in-process (e.g. to build blocks), which
// is very chatty at INFO level.
func SetupLogging(level slog.Level) {
	inner := log.NewTerminalHandlerWithLevel(os.Stderr, level, true)
	log.SetDefault(log.NewLogger(&filteredHandler{inner: inner}))
	log.Root().Write(level, "Set loglevel", "level", level)
}

// filteredHandler drops non-warning records emitted from go-ethereum packages.
type filteredHandler struct {
	inner slog.Handler
}

func (h *filteredHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *filteredHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level < slog.LevelWarn && r.PC != 0 {
		frames := runtime.CallersFrames([]uintptr{r.PC})
		if frame, _ := frames.Next(); strings.Contains(frame.File, "github.com/ethereum/go-ethereum") {
			return nil
		}
	}
	return h.inner.Handle(ctx, r)
}

func (h *filteredHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &filteredHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *filteredHandler) WithGroup(name string) slog.Handler {
	return &filteredHandler{inner: h.inner.WithGroup(name)}
}
