/*
Copyright 2025 Pextra Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"io"
	"sync"
)

// Wrapper over io.Writer that also counts number of written bytes.
type CountingWriter struct {
	W     io.Writer
	count int64
	mu    sync.Mutex
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{W: w}
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.W.Write(p)
	cw.mu.Lock()
	cw.count += int64(n)
	cw.mu.Unlock()
	return n, err
}

func (cw *CountingWriter) Count() int64 {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	return cw.count
}
