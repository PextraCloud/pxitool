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
	"bufio"
	"bytes"
	"fmt"

	"github.com/PextraCloud/pxitool/pkg/pxi/chunk"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/volumeformat"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/volumetype"
)

type Data struct {
	VolumeType     volumetype.VolumeType
	VolumeFormat   volumeformat.VolumeFormat
	VolumeIDLength uint8
	VolumeID       string
	Reserved       [4]byte // Reserved for future use, must be zeroed
	VolumeData     *bufio.Reader
}

type SVOL struct {
	chunk.Chunk
}

// Creates a new SVOL chunk with the specified parameters.
func New(volumeType volumetype.VolumeType, volumeId string) *SVOL {
	volumeIdLen := len(volumeId)
	dataLen := 1 + 1 + 1 + volumeIdLen + 4 // 1 byte for volume type, 1 byte for volume format, 1 byte for volume ID length, len(volumeId) bytes for volume ID, 4 reserved bytes
	c := &SVOL{
		Chunk: chunk.Chunk{
			Length:    uint64(dataLen),
			ChunkType: chunk.ChunkTypeSVOL,
			Data:      make([]byte, dataLen),
		},
	}

	// Set volume type
	c.Data[0] = uint8(volumeType)
	// Default to volumeformat.Raw, gets overridden later
	c.Data[1] = uint8(volumeformat.Raw)
	c.Data[2] = uint8(volumeIdLen)

	volumeIdBytes := []byte(volumeId)
	copy(c.Data[3:3+len(volumeIdBytes)], volumeIdBytes)
	// Zero out reserved bytes
	copy(c.Data[3+len(volumeIdBytes):3+len(volumeIdBytes)+4], []byte{0, 0, 0, 0})

	c.CRC32()
	return c
}

func IncrementLength(c *SVOL, length uint64) {
	c.Length += length
}

func SetVolumeFormat(c *SVOL, format volumeformat.VolumeFormat) {
	c.Data[1] = uint8(format)
}

// TODO: use io.Reader
func GetDataStruct(data []byte) (*Data, error) {
	if len(data) < 7 {
		return nil, fmt.Errorf("data too short for SVOL chunk: %d bytes", len(data))
	}

	volumeType := data[0]
	volumeFormat := data[1]
	volumeIdLen := data[2]

	volumeId := string(data[3 : 3+volumeIdLen])
	reserved := [4]byte{}
	copy(reserved[:], data[3+volumeIdLen:])

	if err := verifyReservedBytes(reserved); err != nil {
		return nil, err
	}

	// 1 byte volume type, 1 byte volume format, 1 byte volume ID length, volumeIdLen bytes for volume ID, 4 reserved bytes, 4 bytes CRC32 (zeroed in this case)
	offset := 1 + 1 + 1 + volumeIdLen + 4 + 4
	return &Data{
		VolumeType:     volumetype.VolumeType(volumeType),
		VolumeFormat:   volumeformat.VolumeFormat(volumeFormat),
		VolumeIDLength: volumeIdLen,
		VolumeID:       volumeId,
		Reserved:       reserved,
		VolumeData:     bufio.NewReader(bytes.NewReader(data[offset:])),
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
