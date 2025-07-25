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
package encryption

import (
	"crypto/cipher"
	"io"
)

const (
	NonceSize = 12
)

type EncryptedWriter struct {
	w        io.Writer
	aesgcm   cipher.AEAD
	nonce    []byte
	counter  uint64
	buf      []byte
	blockPos int
}

type DecryptedReader struct {
	r       io.Reader
	aesgcm  cipher.AEAD
	nonce   []byte
	counter uint64
	buf     []byte
	pos     int
	end     int
}
