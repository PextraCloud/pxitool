/*
Copyright 2026 Pextra Inc.

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
package pxi

type EncryptionCipher string

const (
	// EncryptionCipherAES256 represents the AES-256 encryption cipher.
	EncryptionCipherAES256 EncryptionCipher = "aes256"
)

// EncryptionCiphers is a list of supported encryption ciphers.
var EncryptionCiphers = []EncryptionCipher{EncryptionCipherAES256}

// EncryptionCiphersStrings is a list of supported encryption ciphers as []string.
var EncryptionCiphersStrings = []string{EncryptionCipherAES256.String()}

// Valid checks if the encryption cipher is valid.
func (c EncryptionCipher) Valid() bool {
	switch c {
	case EncryptionCipherAES256:
		return true
	default:
		return false
	}
}

func (c EncryptionCipher) String() string {
	return string(c)
}
