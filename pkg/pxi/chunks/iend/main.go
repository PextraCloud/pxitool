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
package iend

import (
	"fmt"

	"github.com/PextraCloud/pxitool/pkg/pxi/chunk"
)

const (
	DataLength = 0 // Length of the IEND data in bytes
)

type Data struct{}

type IEND struct {
	chunk.Chunk
}

// Creates a new IEND chunk.
func New() *IEND {
	c := &IEND{
		Chunk: chunk.Chunk{
			Length:    0,
			ChunkType: chunk.ChunkTypeIEND,
		},
	}

	c.CRC32()
	return c
}

func GetDataStruct(data *[]byte) (*Data, error) {
	if len(*data) != DataLength {
		return nil, fmt.Errorf("invalid IEND data length: expected %d, got %d", DataLength, len(*data))
	}

	return &Data{}, nil
}
