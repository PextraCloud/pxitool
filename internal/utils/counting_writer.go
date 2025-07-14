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

// Wrapper over io.Writer that counts number of written bytes and saves the first 4 bytes of the written data.
type CountingWriter struct {
	W      io.Writer
	count  int64
	first4 [4]byte
	set4   bool
	mu     sync.Mutex
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{W: w}
}

func (cw *CountingWriter) Write(p []byte) (int, error) {
	n, err := cw.W.Write(p)
	if n > 0 {
		cw.mu.Lock()
		defer cw.mu.Unlock()
		cw.count += int64(n)
		// Save the first 4 bytes if not already set
		if !cw.set4 {
			copy(cw.first4[:], p[:min(n, 4)])
			cw.set4 = true
		}
	}

	return n, err
}

func (cw *CountingWriter) Count() int64 {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	return cw.count
}

func (cw *CountingWriter) First4() [4]byte {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	return cw.first4
}
