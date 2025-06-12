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
package createpxi

import (
	"fmt"
	"os"

	"github.com/PextraCloud/pxitool/internal/encryption"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/encr"
)

func getEncryptedWriter(file *os.File) (*encryption.EncryptedWriter, error) {
	key, salt, err := encryption.CreateEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption key: %v", err)
	}

	nonce, err := encryption.GenerateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Write ENCR chunk
	encrChunk, err := encr.New(nonce, make([]byte, 16), salt)
	if err != nil {
		return nil, fmt.Errorf("failed to create ENCR chunk: %v", err)
	}
	if err = writeChunk(file, &encrChunk.Chunk); err != nil {
		return nil, fmt.Errorf("failed to write ENCR chunk: %v", err)
	}

	// Create encrypted writer for subsequent chunks
	encryptedWriter, err := encryption.NewWriter(file, key, nonce[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypted writer: %v", err)
	}

	return encryptedWriter, nil
}
