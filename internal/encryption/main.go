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
	"crypto/rand"
	"fmt"
	"syscall"

	"github.com/PextraCloud/pxitool/internal/utils"
	"golang.org/x/crypto/argon2"
)

func promptForKey() ([]byte, error) {
	// Check environment variable first
	if envKey, found := syscall.Getenv("PXI_ENCRYPTION_KEY"); found {
		return []byte(envKey), nil
	}

	fmt.Print("Enter encryption key: ")
	password, err := utils.ReadPassword(syscall.Stdin)
	fmt.Println()
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %v", err)
	}
	return password, nil
}

func deriveEncryptionKey(password, salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		return nil, fmt.Errorf("salt cannot be empty")
	}
	if len(password) == 0 {
		return nil, fmt.Errorf("password cannot be empty")
	}

	key := argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
	return key, nil
}

func CreateEncryptionKey() ([]byte, []byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	password, err := promptForKey()
	if err != nil {
		return nil, nil, err
	}

	key, err := deriveEncryptionKey(password, salt)
	if err != nil {
		return nil, nil, err
	}

	return key, salt, nil
}

func DeriveEncryptionKeyFromSalt(salt []byte) ([]byte, error) {
	if len(salt) == 0 {
		return nil, fmt.Errorf("salt cannot be empty")
	}

	password, err := promptForKey()
	if err != nil {
		return nil, err
	}

	key, err := deriveEncryptionKey(password, salt)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func GenerateNonce() ([NonceSize]byte, error) {
	var nonce [NonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nonce, fmt.Errorf("failed to generate nonce: %v", err)
	}
	return nonce, nil
}
