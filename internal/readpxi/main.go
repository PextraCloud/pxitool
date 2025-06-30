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
	"fmt"
	"io"
	"os"

	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/svol"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/encryptiontype"
)

func openFile(path string) (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	return file, nil
}

// Reads a PXI file and returns its chunks.
func Read(path string) (*PXIChunks, error) {
	file, err := openFile(path)
	if err != nil {
		return nil, err
	}

	if err := verifySignature(file); err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	result := &PXIChunks{}

	buf := bufio.NewReader(file)
	ihdrData, err := readIHDR(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read IHDR chunk: %w", err)
	}
	result.IHDR = ihdrData

	reader := io.Reader(buf)
	if ihdrData.EncryptionType != encryptiontype.None {
		encrData, err := readENCR(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read ENCR chunk: %w", err)
		}
		result.ENCR = encrData

		decReader, err := getDecryptedReader(buf, encrData)
		if err != nil {
			return nil, fmt.Errorf("failed to get decrypted reader: %w", err)
		}
		reader = decReader
	}

	confData, err := readCONF(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read CONF chunk: %w", err)
	}
	result.CONF = confData

	var svols []*svol.Data
	for {
		svol, err := readSVOL(reader)
		// IEND chunk found
		if svol == nil && err == nil {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read SVOL chunk: %w", err)
		}
		svols = append(svols, svol)
	}
	result.SVOL = svols

	if closer, ok := file.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return nil, fmt.Errorf("failed to close file: %w", err)
		}
	}
	return result, nil
}
