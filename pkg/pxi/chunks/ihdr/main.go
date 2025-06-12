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
package ihdr

import (
	"fmt"

	"github.com/PextraCloud/pxitool/pkg/pxi/chunk"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/compressiontype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/encryptiontype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/instancetype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/pxiversion"
)

const (
	DataLength = 16 // Length of the IHDR data in bytes
)

type Data struct {
	PXIVersion      pxiversion.PXIVersion
	InstanceType    instancetype.InstanceType
	CompressionType compressiontype.CompressionType
	EncryptionType  encryptiontype.EncryptionType
	Reserved        [4]byte // Reserved for future use, must be zeroed
}

type IHDR struct {
	chunk.Chunk
}

// Creates a new IHDR chunk with the specified parameters.
func New(pxiVersion pxiversion.PXIVersion, instanceType instancetype.InstanceType, compressionType compressiontype.CompressionType, encryptionType encryptiontype.EncryptionType) *IHDR {
	c := &IHDR{
		Chunk: chunk.Chunk{
			Length:    DataLength,
			ChunkType: chunk.ChunkTypeIHDR,
			Data:      make([]byte, DataLength),
		},
	}

	c.Data[0] = uint8(pxiVersion)
	c.Data[1] = uint8(instanceType)
	c.Data[2] = uint8(compressionType)
	c.Data[3] = uint8(encryptionType)
	// Zero out reserved bytes
	copy(c.Data[4:8], make([]byte, 4))

	c.CRC32()
	return c
}

func GetDataStruct(data []byte) (*Data, error) {
	dataLen := len(data)
	if dataLen != DataLength {
		return nil, fmt.Errorf("invalid IHDR data length: expected %d, got %d", DataLength, dataLen)
	}

	pxiVersion := pxiversion.PXIVersion(data[0])
	instanceType := instancetype.InstanceType(data[1])
	compressionType := compressiontype.CompressionType(data[2])
	encryptionType := encryptiontype.EncryptionType(data[3])

	reserved := [4]byte{}
	copy(reserved[:], data[4:8])
	if err := verifyReservedBytes(reserved); err != nil {
		return nil, err
	}

	return &Data{
		PXIVersion:      pxiVersion,
		InstanceType:    instanceType,
		CompressionType: compressionType,
		EncryptionType:  encryptionType,
		Reserved:        reserved,
	}, nil
}

func verifyReservedBytes(reserved [4]byte) error {
	for _, b := range reserved {
		if b != 0 {
			return fmt.Errorf("reserved bytes must be zero, found: %x", reserved)
		}
	}
	return nil
}
