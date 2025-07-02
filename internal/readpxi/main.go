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

	"github.com/PextraCloud/pxitool/pkg/pxi/constants/encryptiontype"
)

func openFile(path string) (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	return file, nil
}

func closeFile(file io.Reader) error {
	if closer, ok := file.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
	}
	return nil
}

// Reads a PXI file and returns its chunks.
func ReadChunks(path string) (*PXIChunks, error) {
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

	svols, err := readSVOLUntilIEND(reader)
	if err != nil {
		return nil, err
	}
	result.SVOL = svols

	if err := closeFile(file); err != nil {
		return nil, err
	}
	return result, nil
}

// Reads a PXI file and skips encrypted chunks, returning its chunks.
func ReadChunksSkipEncrypted(path string) (*PXIChunks, error) {
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

		return result, nil
	}

	confData, err := readCONF(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read CONF chunk: %w", err)
	}
	result.CONF = confData

	svols, err := readSVOLUntilIEND(reader)
	if err != nil {
		return nil, err
	}
	result.SVOL = svols

	if err := closeFile(file); err != nil {
		return nil, err
	}
	return result, nil
}

// Reads a PXI file and returns information about it
func GetInfo(path string, skipEncrypted bool) (*ReadPXIOutput, error) {
	var chunks *PXIChunks
	var err error
	if skipEncrypted {
		chunks, err = ReadChunksSkipEncrypted(path)
	} else {
		chunks, err = ReadChunks(path)
	}
	if err != nil {
		return nil, err
	}

	output := &ReadPXIOutput{
		PXIVersion:      chunks.IHDR.PXIVersion,
		InstanceType:    chunks.IHDR.InstanceType,
		CompressionType: chunks.IHDR.CompressionType,
		EncryptionType:  chunks.IHDR.EncryptionType,
		Config:          nil,
		Volumes:         nil,
	}
	// Only if skipEncrypted is false
	if chunks.CONF != nil {
		output.Config = &chunks.CONF.Config
		output.Volumes = chunks.CONF.Config.Volumes
	}

	return output, nil
}
