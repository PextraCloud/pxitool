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
package encr

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/PextraCloud/pxitool/pkg/pxi/chunk"
)

const (
	NonceLength = 12 // Length of the nonce in bytes
)

var (
	ErrInvalidAEADLength = errors.New("Invalid AEAD length, must be between 16 and 65535 bytes")
	ErrInvalidSaltLength = errors.New("Invalid salt length, must be between 0 and 65535 bytes")
)

type Data struct {
	Nonce   [NonceLength]byte
	AEADLen uint16 // Length of AEAD in bytes
	SaltLen uint16 // Length of Salt in bytes
	AEAD    []byte // Authenticated tag data (AEAD)
	Salt    []byte // Optional salt for key derivation
}

type ENCR struct {
	chunk.Chunk
}

// Creates a new ENCR chunk with the specified parameters.
func New(nonce [NonceLength]byte, aead []byte, salt []byte) (*ENCR, error) {
	aeadLen := len(aead)
	if aeadLen < 16 || aeadLen > 65535 {
		return nil, ErrInvalidAEADLength
	}

	saltLen := len(salt)
	if saltLen > 65535 {
		return nil, ErrInvalidSaltLength
	}

	dataLen := uint64(NonceLength + aeadLen + saltLen + 2 + 2) // +2 for AEAD length uint16, +2 for salt length uint16
	c := &ENCR{
		Chunk: chunk.Chunk{
			Length:    dataLen,
			ChunkType: chunk.ChunkTypeENCR,
			Data:      make([]byte, dataLen),
		},
	}

	// Copy nonce to the first 12 bytes of Data
	copy(c.Data[:NonceLength], nonce[:])
	// Write AEAD length (uint16) at position 12
	binary.BigEndian.PutUint16(c.Data[NonceLength:NonceLength+2], uint16(aeadLen))
	// Write Salt length (uint16) at position 12 + AEAD length
	binary.BigEndian.PutUint16(c.Data[NonceLength+2:NonceLength+4], uint16(saltLen))
	// Copy AEAD data after the nonce and lengths
	copy(c.Data[NonceLength+4:NonceLength+4+aeadLen], aead)
	// Copy Salt data after the AEAD data
	copy(c.Data[NonceLength+4+aeadLen:], salt)

	c.CRC32()
	return c, nil
}

func GetDataStruct(data []byte) (*Data, error) {
	dataLen := len(data)
	if dataLen < NonceLength+4 { // 2+2 for AEAD and Salt lengths
		return nil, fmt.Errorf("data too short for ENCR chunk: expected at least %d bytes, got %d bytes", NonceLength+4, dataLen)
	}

	var d Data
	copy(d.Nonce[:], data[:NonceLength])
	d.AEADLen = binary.BigEndian.Uint16(data[NonceLength : NonceLength+2])
	d.SaltLen = binary.BigEndian.Uint16(data[NonceLength+2 : NonceLength+4])

	if d.AEADLen < 16 || d.AEADLen > 65535 {
		return nil, ErrInvalidAEADLength
	}
	if d.SaltLen > 65535 {
		return nil, ErrInvalidSaltLength
	}

	if dataLen < int(NonceLength+4+d.AEADLen+d.SaltLen) {
		return nil, fmt.Errorf("data too short for ENCR chunk: expected at least %d bytes, got %d bytes", NonceLength+4+d.AEADLen+d.SaltLen, dataLen)
	}

	d.AEAD = data[NonceLength+4 : NonceLength+4+d.AEADLen]
	d.Salt = data[NonceLength+4+d.AEADLen : NonceLength+4+d.AEADLen+d.SaltLen]

	return &d, nil
}
