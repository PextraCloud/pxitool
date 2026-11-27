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
	"os"
	"path"

	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func createWorkingDirectory() (string, error) {
	dir, err := os.MkdirTemp("", "pxi-create-*")
	if err != nil {
		return "", err
	}
	return dir, nil
}

// preparePXI prepares the PXI by creating a working directory and writing the manifest file.
//
// The working directory is not automatically cleaned up; the caller is responsible for removing it after use.
func preparePXI(instance pxi.Instance) (string, error) {
	// Create working directory
	workingDirectory, err := createWorkingDirectory()
	if err != nil {
		return "", err
	}

	// Write manifest
	manifest := getManifest(instance)
	manifestPath := path.Join(workingDirectory, pxi.ManifestFileName)
	if err := writeManifest(manifest, manifestPath); err != nil {
		return "", err
	}

	return workingDirectory, nil
}
