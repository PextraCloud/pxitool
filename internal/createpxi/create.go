/*
Copyright 2026 Pextra Inc.

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
package createpxi

import (
	"github.com/PextraCloud/pxitool/internal/dar"
	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func createPXI(options CreatePXIOptions, workingDirectory string, files []string) error {
	darCipher := ""
	switch options.EncryptionCipher {
	case pxi.EncryptionCipherAES256:
		darCipher = "aes"
	}
	darCompression := ""
	switch options.CompressionAlgorithm {
	case pxi.CompressionAlgorithmZstd:
		darCompression = "zstd"
	case pxi.CompressionAlgorithmGzip:
		darCompression = "gzip"
	}

	opts := dar.CreateOptions{
		OutputBasename: options.OutputPath,
		RootDir:        workingDirectory,
		Overwrite:      options.Overwrite,
		Files:          files,
		Cipher:         darCipher,
		Key:            options.EncryptionKey,
		Compression:    darCompression,
	}
	return dar.Create(opts)
}
