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
	"encoding/json"
	"fmt"
	"os"

	"github.com/PextraCloud/pxitool/internal/version"
	"github.com/PextraCloud/pxitool/pkg/pxi"
)

// getManifest creates a pxi.Manifest struct based on the provided pxi.Instance.
func getManifest(instance pxi.Instance) pxi.Manifest {
	v := version.Version()
	manifest := pxi.Manifest{
		Version:   pxi.ManifestVersion2,
		CreatedBy: v,
		Instance:  instance,
	}
	return manifest
}

// writeManifest writes the marshalled pxi.Manifest to path.
func writeManifest(manifest pxi.Manifest, path string) error {
	if err := manifest.Valid(); err != nil {
		return fmt.Errorf("invalid manifest %+v: %w", manifest, err)
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest to JSON: %w", err)
	}
	if err := os.WriteFile(path, manifestBytes, 0644); err != nil {
		return fmt.Errorf("failed to write manifest to %s: %w", path, err)
	}
	return nil
}
