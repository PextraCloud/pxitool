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
package readpxi

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/PextraCloud/pxitool/internal/dar"
	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func createWorkingDirectory() (string, error) {
	dir, err := os.MkdirTemp("", "pxi-read-*")
	if err != nil {
		return "", err
	}
	return dir, nil
}

func readPXI(inputBasename string, workingDirectory string, files []string, encryptionKey string) error {
	/*darCipher := "" TODO add options, and make common with createpxi2
	switch options.EncryptionCipher {
	case pxi2.EncryptionCipherAES256:
		darCipher = "aes"
	}*/

	opts := dar.ExtractOptions{
		InputBasename: inputBasename,
		RootDir:       workingDirectory,
		Overwrite:     false,
		Files:         files,
		// Cipher:        darCipher,
		Key: encryptionKey,
	}
	return dar.Extract(opts)
}

func readManifest(workingDirectory string) (*pxi.Manifest, error) {
	manifestPath := path.Join(workingDirectory, pxi.ManifestFileName)
	file, err := os.Open(manifestPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var manifest pxi.Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return nil, err
	}
	if err := manifest.Valid(); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	return &manifest, nil
}

// ReadPXI reads the PXI file at the given path and returns the contents of its manifest file.
func ReadPXI(path string, encryptionKey string) (*pxi.Manifest, error) {
	workingDirectory, err := createWorkingDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to prepare to read PXI: %w", err)
	}
	defer os.RemoveAll(workingDirectory)

	files := []string{pxi.ManifestFileName}
	err = readPXI(path, workingDirectory, files, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read PXI: %w", err)
	}

	manifest, err := readManifest(workingDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	return manifest, nil
}

func ReadPXIHeader(inputBasename string) (*dar.Header, error) {
	options := dar.ListHeaderOptions{
		InputBasename: inputBasename,
	}
	return dar.ListHeader(options)
}
