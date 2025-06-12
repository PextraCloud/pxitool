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
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Generic structure for a PXI chunk.
type Chunk struct {
	ChunkType [4]byte
	Length    uint64
	Data      []byte
	CRC       uint32
}

var (
	ErrChunkTooShort = errors.New("chunk data is too short")

	ChunkTypeIHDR = [4]byte{0x49, 0x48, 0x44, 0x52} // "IHDR"
	ChunkTypeENCR = [4]byte{0x45, 0x4E, 0x43, 0x52} // "ENCR"
	ChunkTypeIEND = [4]byte{0x49, 0x45, 0x4E, 0x44} // "IEND"
	ChunkTypeCONF = [4]byte{0x43, 0x4F, 0x4E, 0x46} // "CONF"
	ChunkTypeSVOL = [4]byte{0x53, 0x56, 0x4F, 0x4C} // "SVOL"
)

const (
	ChunkOverhead = 16 // 8 bytes for Length, 4 bytes for ChunkType, 4 bytes for CRC
)

// Converts the chunk to a byte slice.
// The first 8 bytes are the length, followed by the
// 4-byte chunk type, the data, and finally, a
// 4-byte CRC checksum.
func (c *Chunk) Bytes() []byte {
	if len(c.Data) != int(c.Length) {
		panic("Data length does not match chunk Length")
	}
	buf := bytes.NewBuffer(nil)

	length := make([]byte, 8)
	binary.BigEndian.PutUint64(length, c.Length)
	crc := make([]byte, 4)
	binary.BigEndian.PutUint32(crc, c.CRC)

	buf.Write(length)
	buf.Write(c.ChunkType[:])
	buf.Write(c.Data)
	buf.Write(crc)

	return buf.Bytes()
}

// Debug
func printChunk(c *Chunk) {
	if c == nil {
		return
	}

	fmt.Printf("----- chunk %s -----\n", c.ChunkType)
	fmt.Printf("Length: %d bytes\n", c.Length)
	if len(c.Data) > 0 {
		fmt.Printf("Data: %x\n", c.Data)
	} else {
		fmt.Println("Data: <empty>")
	}
	fmt.Printf("CRC32: %08x\n", c.CRC)
	fmt.Println("----------------------")
}

// Parses a chunk from the provided io.Reader.
// TODO: do not read everything into memory
func ParseChunk(r io.Reader) (*Chunk, error) {
	var length uint64
	var chunkType [4]byte
	var crc uint32

	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	if _, err := io.ReadFull(r, chunkType[:]); err != nil {
		return nil, err
	}

	fmt.Printf("Reading chunk type: %s, length: %d bytes\n", chunkType, length)
	chunkData := make([]byte, length)
	if _, err := io.ReadFull(r, chunkData); err != nil {
		return nil, err
	}
	if err := binary.Read(r, binary.BigEndian, &crc); err != nil {
		return nil, err
	}

	c := &Chunk{
		ChunkType: chunkType,
		Length:    length,
		Data:      chunkData,
	}
	c.CRC32()
	if err := c.VerifyCRC32(crc); err != nil {
		return nil, err
	}

	printChunk(c)
	return c, nil
}
