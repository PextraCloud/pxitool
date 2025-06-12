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
package conf

import (
	"encoding/json"
	"fmt"

	"github.com/PextraCloud/pxitool/pkg/pxi/chunk"
)

type Data struct {
	Config InstanceConfigGeneric
}

type CONF struct {
	chunk.Chunk
}

// Creates a new CONF chunk with the specified parameters.
func New(config *InstanceConfigGeneric) (*CONF, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	c := &CONF{
		Chunk: chunk.Chunk{
			Length:    uint64(len(jsonBytes)),
			Data:      jsonBytes,
			ChunkType: chunk.ChunkTypeCONF,
		},
	}

	c.CRC32()
	return c, nil
}

func GetDataStruct(data *[]byte) (*Data, error) {
	if len(*data) < 1 {
		return nil, fmt.Errorf("data length too short for CONF chunk: %d bytes", len(*data))
	}

	var dataGeneric InstanceConfigGeneric
	var err error

	if err = dataGeneric.UnmarshalConfJSON(*data); err != nil {
		return nil, fmt.Errorf("failed to parse CONF chunk data: %w", err)
	}
	return &Data{
		Config: dataGeneric,
	}, nil
}
