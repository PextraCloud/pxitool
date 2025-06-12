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
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
)

// Creates a reader that decrypts data from the underlying reader using AES-GCM
func NewDecryptedReader(r io.Reader, key []byte, nonce []byte) (*DecryptedReader, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", KeySize, len(key))
	}
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce size: expected %d bytes, got %d", NonceSize, len(nonce))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &DecryptedReader{
		r:       r,
		aesgcm:  aesgcm,
		nonce:   nonce,
		counter: 0,
		buf:     nil,
		pos:     0,
		end:     0,
	}, nil
}

// Read decrypts data from the underlying reader
func (dr *DecryptedReader) Read(p []byte) (int, error) {
	if dr.pos >= dr.end {
		// Need to read and decrypt the next block
		if err := dr.readNextBlock(); err != nil {
			return 0, err
		}
		if dr.pos >= dr.end { // Still no data after reading
			return 0, io.EOF
		}
	}

	// Copy decrypted data to the output buffer
	n := copy(p, dr.buf[dr.pos:dr.end])
	dr.pos += n
	return n, nil
}

// readNextBlock reads the next encrypted block and decrypts it
func (dr *DecryptedReader) readNextBlock() error {
	// Read the size of the next ciphertext block
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(dr.r, lenBuf); err != nil {
		return err
	}

	size := binary.BigEndian.Uint32(lenBuf)
	if size == 0 {
		return io.EOF
	}

	// Read the ciphertext
	ciphertext := make([]byte, size)
	if _, err := io.ReadFull(dr.r, ciphertext); err != nil {
		return err
	}

	// Create a nonce with a counter to ensure uniqueness
	nonceWithCounter := make([]byte, NonceSize)
	copy(nonceWithCounter, dr.nonce)

	// Use the last 8 bytes for a counter
	for i := range 8 {
		nonceWithCounter[NonceSize-1-i] = byte(dr.counter >> (i * 8))
	}
	dr.counter++

	// Decrypt the ciphertext
	plaintext, err := dr.aesgcm.Open(nil, nonceWithCounter, ciphertext, nil)
	if err != nil {
		return err
	}

	// Update buffer pointers
	dr.buf = plaintext
	dr.pos = 0
	dr.end = len(plaintext)

	return nil
}
