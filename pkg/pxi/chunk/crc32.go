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
package chunk

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
)

var (
	table = crc32.MakeTable(crc32.IEEE)
)

func calculateChunkCRC32(data *[]byte) uint32 {
	if data == nil || len(*data) == 0 {
		return 0
	}

	crc := crc32.New(table)
	crc.Write(*data)
	return crc.Sum32() ^ 0xFFFFFFFF // bitwise NOT
}

// Calculates and sets the CRC32 checksum (big-endian) for the chunk data.
func (c *Chunk) CRC32() uint32 {
	if c.ChunkType == ChunkTypeSVOL {
		c.CRC = 0
		return c.CRC
	}

	if c.CRC != 0 {
		return c.CRC
	} else if c.Data == nil || len(c.Data) == 0 {
		return 0
	}

	crc := calculateChunkCRC32(&c.Data)
	c.CRC = crc
	return crc
}

func (c *Chunk) VerifyCRC32(crc uint32) error {
	if c.CRC != crc {
		return fmt.Errorf("CRC mismatch: expected %08X, got %08X", c.CRC, crc)
	}
	return nil
}
func (c *Chunk) VerifyCRC32Bytes(crc []byte) error {
	if len(crc) != 4 {
		return fmt.Errorf("invalid CRC length: expected 4 bytes, got %d", len(crc))
	}

	crc32Value := binary.BigEndian.Uint32(crc)
	if c.CRC != crc32Value {
		return fmt.Errorf("CRC mismatch: expected %08X, got %08X", c.CRC, crc32Value)
	}
	return nil
}
