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
	"fmt"
	"io"
)

const (
	KeySize = 32
)

// Creates a writer that encrypts data using AES-256-GCM
func NewWriter(w io.Writer, key []byte, nonce []byte) (*EncryptedWriter, error) {
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

	return &EncryptedWriter{
		w:       w,
		aesgcm:  aesgcm,
		nonce:   nonce,
		counter: 0,
		// Buffer is 16KB
		buf:      make([]byte, 16*1024),
		blockPos: 0,
	}, nil
}

// Write encrypts data and writes it to the underlying writer
func (ew *EncryptedWriter) Write(p []byte) (int, error) {
	totalWritten := 0
	for len(p) > 0 {
		// Calculate how much we can write to the buffer
		remainingBuf := len(ew.buf) - ew.blockPos
		writeLen := remainingBuf
		if len(p) < writeLen {
			writeLen = len(p)
		}

		// Copy data to the buffer
		copy(ew.buf[ew.blockPos:], p[:writeLen])
		ew.blockPos += writeLen
		p = p[writeLen:]
		totalWritten += writeLen

		// If buffer is full, encrypt and write it
		if ew.blockPos == len(ew.buf) {
			if err := ew.flushBuffer(); err != nil {
				return totalWritten, err
			}
		}
	}
	return totalWritten, nil
}

// Close flushes any remaining data and finalizes the encryption
func (ew *EncryptedWriter) Close() error {
	if ew.blockPos > 0 {
		return ew.flushBuffer()
	}
	return nil
}

// flushBuffer encrypts the current buffer and writes it to the underlying writer
func (ew *EncryptedWriter) flushBuffer() error {
	if ew.blockPos == 0 {
		return nil
	}

	// Create a nonce with a counter to ensure uniqueness
	nonceWithCounter := make([]byte, NonceSize)
	copy(nonceWithCounter, ew.nonce)

	// Use the last 8 bytes for a counter
	for i := range 8 {
		nonceWithCounter[NonceSize-1-i] = byte(ew.counter >> (i * 8))
	}
	ew.counter++

	// Encrypt the data
	ciphertext := ew.aesgcm.Seal(nil, nonceWithCounter, ew.buf[:ew.blockPos], nil)

	// Write the ciphertext size and ciphertext
	lenBuf := make([]byte, 4)
	lenBuf[0] = byte(len(ciphertext) >> 24)
	lenBuf[1] = byte(len(ciphertext) >> 16)
	lenBuf[2] = byte(len(ciphertext) >> 8)
	lenBuf[3] = byte(len(ciphertext))

	if _, err := ew.w.Write(lenBuf); err != nil {
		return err
	}
	if _, err := ew.w.Write(ciphertext); err != nil {
		return err
	}

	// Reset buffer position
	ew.blockPos = 0
	return nil
}
