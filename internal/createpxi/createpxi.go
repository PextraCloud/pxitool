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
	"fmt"
	"os"
	"path"

	"github.com/PextraCloud/pxitool/internal/backup"
	"github.com/PextraCloud/pxitool/pkg/pxi"
)

// backupVolumes backs up all non-excluded volumes of the given instance into workingDirectory, with the specified concurrency.
func backupVolumes(instance pxi.Instance, workingDirectory string, concurrency int) ([]string, error) {
	includedLen := 0
	for _, volume := range instance.Resources.Volumes {
		if !volume.Exclude() {
			includedLen++
		}
	}
	// No volumes to back up (all excluded)
	if includedLen == 0 {
		return []string{}, nil
	}

	pathsChan := make(chan string, includedLen)
	errChan := make(chan error, includedLen)
	sem := make(chan struct{}, concurrency)

	for _, volume := range instance.Resources.Volumes {
		// Skip excluded volumes
		if volume.Exclude() {
			continue
		}

		sem <- struct{}{}
		go func() {
			defer func() { <-sem }()
			volumePath, err := backup.BackupVolume(volume.Path, volume.Id, volume.Type, workingDirectory)
			if err != nil {
				errChan <- err
				return
			}
			pathsChan <- path.Base(volumePath)
		}()
	}

	var paths []string
	for i := 0; i < includedLen; i++ {
		select {
		case path := <-pathsChan:
			paths = append(paths, path)
		case err := <-errChan:
			return nil, err
		}
	}

	return paths, nil
}

// CreatePXIOptions represents the options for creating a PXI file.
type CreatePXIOptions struct {
	Instance             pxi.Instance
	OutputPath           string
	ExcludeVolumes       []string
	Overwrite            bool
	EncryptionCipher     pxi.EncryptionCipher
	EncryptionKey        string
	CompressionAlgorithm pxi.CompressionAlgorithm
	Concurrency          int
}

func excludeVolumes(instance pxi.Instance, excludeVolumes []string) error {
	set := make(map[string]struct{})
	for _, volumeId := range excludeVolumes {
		if _, ok := set[volumeId]; ok {
			return fmt.Errorf("duplicate volume ID %s in exclude list", volumeId)
		}
		set[volumeId] = struct{}{}
	}
	// Mark volumes for exclusion based on excludeVolumes list
	for i := range instance.Resources.Volumes {
		if _, ok := set[instance.Resources.Volumes[i].Id]; ok {
			instance.Resources.Volumes[i].SetExclude(true)
		}
	}

	return nil
}

// CreatePXI creates a PXI file based on the provided options, writing it to the specified output path.
func CreatePXI(options CreatePXIOptions) error {
	if err := excludeVolumes(options.Instance, options.ExcludeVolumes); err != nil {
		return err
	}

	if options.Concurrency == 0 {
		options.Concurrency = 2
	} else if options.Concurrency < 1 {
		options.Concurrency = 1
	}

	workingDirectory, err := preparePXI(options.Instance)
	if err != nil {
		return fmt.Errorf("failed to prepare PXI: %w", err)
	}
	defer os.RemoveAll(workingDirectory)

	files, err := backupVolumes(options.Instance, workingDirectory, options.Concurrency)
	if err != nil {
		return fmt.Errorf("failed to back up volumes: %w", err)
	}
	// Add manifest file to list of files to be included in PXI archive
	files = append(files, pxi.ManifestFileName)

	err = createPXI(options, workingDirectory, files)
	if err != nil {
		return fmt.Errorf("failed to create PXI: %w", err)
	}
	return nil
}
