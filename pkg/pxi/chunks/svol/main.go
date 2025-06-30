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
package svol

import (
	"fmt"
	"io"

	"github.com/PextraCloud/pxitool/pkg/pxi/chunk"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/volumetype"
)

type Data struct {
	VolumeType volumetype.VolumeType
	VolumeID   string
	Reserved   [4]byte // Reserved for future use, must be zeroed
	VolumeData *io.Reader
}

type SVOL struct {
	chunk.Chunk
	VolumeDataReader *io.Reader // Reader for volume data
}

// Creates a new SVOL chunk with the specified parameters.
func New(volumeType volumetype.VolumeType, volumeId string) *SVOL {
	dataLen := 1 + len(volumeId) + 4 // 1 byte for volume type uint8, len(volumeId) bytes for volume ID, 4 reserved bytes
	c := &SVOL{
		Chunk: chunk.Chunk{
			Length:    uint64(dataLen),
			ChunkType: chunk.ChunkTypeSVOL,
			Data:      make([]byte, dataLen),
		},
	}

	// Set volume type
	c.Data[0] = byte(volumeType)
	// Set volume ID
	volumeIdBytes := []byte(volumeId)
	copy(c.Data[1:1+len(volumeIdBytes)], volumeIdBytes)
	// Zero out reserved bytes
	copy(c.Data[1+len(volumeIdBytes):1+len(volumeIdBytes)+4], []byte{})

	c.CRC32()
	return c
}

func IncrementLength(c *SVOL, length uint64) {
	c.Length += length
}

// TODO: use io.Reader
func GetDataStruct(data []byte) (*Data, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("data too short for SVOL chunk: %d bytes", len(data))
	}

	volumeType := volumetype.VolumeType(data[0])
	volumeIdLen := len(data) - 5 // 1 byte for volume type, 4 reserved bytes
	if volumeIdLen < 1 {
		return nil, fmt.Errorf("invalid volume ID length: %d", volumeIdLen)
	}

	volumeId := string(data[1 : 1+volumeIdLen-4])
	reserved := [4]byte{}
	copy(reserved[:], data[1+volumeIdLen-4:])

	if err := verifyReservedBytes(reserved); err != nil {
		return nil, err
	}
	return &Data{
		VolumeType: volumeType,
		VolumeID:   volumeId,
		Reserved:   reserved,
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
