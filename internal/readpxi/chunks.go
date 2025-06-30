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
	"fmt"
	"io"

	"github.com/PextraCloud/pxitool/pkg/pxi/chunk"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/conf"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/encr"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/iend"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/ihdr"
)

func readIHDR(reader io.Reader) (*ihdr.Data, error) {
	var c *chunk.Chunk
	var err error

	if c, err = chunk.ParseChunk(reader); err != nil {
		return nil, err
	}
	if c.ChunkType != chunk.ChunkTypeIHDR {
		return nil, fmt.Errorf("expected IHDR chunk, got %s", c.ChunkType)
	}

	var ihdrChunk *ihdr.Data
	if ihdrChunk, err = ihdr.GetDataStruct(c.Data); err != nil {
		return nil, fmt.Errorf("error parsing IHDR chunk: %w", err)
	}
	return ihdrChunk, nil
}
func readENCR(reader io.Reader) (*encr.Data, error) {
	var c *chunk.Chunk
	var err error

	if c, err = chunk.ParseChunk(reader); err != nil {
		return nil, err
	}
	if c.ChunkType != chunk.ChunkTypeENCR {
		return nil, fmt.Errorf("expected ENCR chunk, got %s", c.ChunkType)
	}

	var encrChunk *encr.Data
	if encrChunk, err = encr.GetDataStruct(c.Data); err != nil {
		return nil, fmt.Errorf("error parsing ENCR chunk: %w", err)
	}
	return encrChunk, nil
}
func readCONF(reader io.Reader) (*conf.Data, error) {
	var c *chunk.Chunk
	var err error

	if c, err = chunk.ParseChunk(reader); err != nil {
		return nil, err
	}
	if c.ChunkType != chunk.ChunkTypeCONF {
		return nil, fmt.Errorf("expected CONF chunk, got %s", c.ChunkType)
	}

	var confChunk *conf.Data
	if confChunk, err = conf.GetDataStruct(&c.Data); err != nil {
		return nil, fmt.Errorf("error parsing CONF chunk: %w", err)
	}

	return confChunk, nil
}

func readSVOL(reader io.Reader) (*svol.Data, error) {
	var c *chunk.Chunk
	var err error

	if c, err = chunk.ParseChunk(reader); err != nil {
		return nil, err
	}
	if c.ChunkType != chunk.ChunkTypeSVOL {
		return nil, nil // Return nil if chunk type is not SVOL
	}

	var svolChunk *svol.Data
	if svolChunk, err = svol.GetDataStruct(c.Data); err != nil {
		return nil, fmt.Errorf("error parsing SVOL chunk: %w", err)
	}

	return svolChunk, nil
}
