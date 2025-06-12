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
package readpxi

import (
	"bufio"

	"github.com/PextraCloud/pxitool/internal/encryption"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/encr"
)

func getDecryptedReader(buf *bufio.Reader, encrData *encr.Data) (*encryption.DecryptedReader, error) {
	key, err := encryption.DeriveEncryptionKeyFromSalt(encrData.Salt)
	if err != nil {
		return nil, err
	}

	reader, err := encryption.NewDecryptedReader(buf, key, encrData.Nonce[:])
	if err != nil {
		return nil, err
	}
	return reader, nil
}
