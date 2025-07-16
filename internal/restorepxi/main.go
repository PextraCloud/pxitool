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
package restorepxi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/PextraCloud/pxitool/internal/readpxi"
	"github.com/PextraCloud/pxitool/pkg/log"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/conf"
)

type restorePathsType map[string]string
type svolMapType map[string]*bufio.Reader
type volumeMapType map[string]*conf.InstanceVolume

type fileWritersStruct struct {
	file *os.File
	buf  *bufio.Writer
}
type fileWritersType map[string]fileWritersStruct

func makeMaps(chunks *readpxi.PXIChunks) (svolMapType, volumeMapType, error) {
	if chunks == nil {
		return nil, nil, fmt.Errorf("chunks cannot be nil")
	}

	svolMap := make(svolMapType)
	for _, chunk := range chunks.SVOL {
		if chunk == nil {
			return nil, nil, fmt.Errorf("SVOL chunk cannot be nil")
		}
		if chunk.VolumeData == nil {
			return nil, nil, fmt.Errorf("SVOL chunk VolumeData cannot be nil")
		}
		svolMap[chunk.VolumeID] = chunk.VolumeData
	}

	config := chunks.CONF.Config
	volumeMap := make(volumeMapType)
	for _, volume := range config.Volumes {
		// Skip volumes that are in the config but not backed up
		if _, found := svolMap[volume.ID]; !found {
			log.Debug("Volume '%s' not found in SVOL chunks, skipping", volume.ID)
			continue
		}

		log.Debug("Adding volume '%s' to volumeMap", volume.ID)
		volumeMap[volume.ID] = &volume

	}
	return svolMap, volumeMap, nil
}

func getFileWriters(restorePaths restorePathsType) (fileWritersType, error) {
	writers := make(fileWritersType)
	for volumeID, restorePath := range restorePaths {
		if volumeID == "rootfs" {
			// Skip rootfs for now, as it requires special handling
			continue
		}

		file, err := os.OpenFile(restorePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
		if err != nil {
			return nil, fmt.Errorf("failed to open file '%s': %w", restorePath, err)
		}
		writers[volumeID] = fileWritersStruct{
			file: file,
			buf:  bufio.NewWriter(file),
		}
	}
	return writers, nil
}

func writeConfig(outputFileName string, config *conf.InstanceConfigGeneric) error {
	file, err := os.OpenFile(outputFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return fmt.Errorf("failed to open file '%s': %w", outputFileName, err)
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	_, err = file.Write(jsonData)
	return err
}

func restoreLXC(restorePath string, reader *bufio.Reader) error {
	dirMode := os.FileMode(0755) // Default directory mode
	if dirInfo, err := os.Stat(restorePath); err == nil {
		log.Debug("Removing existing rootfs path '%s'", restorePath)
		dirMode = dirInfo.Mode()
		if err := os.RemoveAll(restorePath); err != nil {
			return fmt.Errorf("failed to remove existing rootfs path '%s': %w", restorePath, err)
		}
	}

	log.Debug("Creating directory for rootfs at '%s'", restorePath)
	if err := os.MkdirAll(restorePath, dirMode); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", restorePath, err)
	}

	log.Debug("Restoring rootfs to path '%s'", restorePath)
	cmd := exec.Command("tar", "-x", "-C", restorePath)
	cmd.Stdin = reader

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		log.Error("stderr: %s", errBuf.String())
		return fmt.Errorf("failed to untar rootfs: %w", err)
	}

	log.Debug("Finished restoring rootfs to path '%s'", restorePath)
	return nil
}

func Restore(chunks *readpxi.PXIChunks, restorePaths restorePathsType, outputFileName string) error {
	if chunks == nil {
		return fmt.Errorf("config cannot be nil")
	}

	config := chunks.CONF.Config
	log.Debug("Restoring Pextra Image with config: %+v", config)
	if err := writeConfig(outputFileName, &config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	svolMap, volumeMap, err := makeMaps(chunks)
	if err != nil {
		return fmt.Errorf("failed to create maps from chunks: %w", err)
	}

	fileWriters, err := getFileWriters(restorePaths)
	if err != nil {
		return fmt.Errorf("failed to get file writers: %w", err)
	}

	for volumeID, restorePath := range restorePaths {
		reader, found := svolMap[volumeID]
		if !found {
			return fmt.Errorf("no SVOL chunk found for volume ID '%s'", volumeID)
		}

		// Handle special case for LXC rootfs
		if volumeID == "rootfs" {
			if !found {
				return fmt.Errorf("no SVOL chunk found for rootfs volume ID '%s'", volumeID)
			}
			if err := restoreLXC(restorePath, reader); err != nil {
				return err
			}
			continue
		}

		_, found = volumeMap[volumeID]
		if !found {
			return fmt.Errorf("volume ID '%s' was not found in the config or was not backed up", volumeID)
		}

		writer, found := fileWriters[volumeID]
		if !found {
			return fmt.Errorf("no writer found for volume ID '%s'", volumeID)
		}

		_, err = reader.WriteTo(writer.buf)
		if err != nil {
			return fmt.Errorf("failed to write volume '%s' to file: %w", volumeID, err)
		}

		if err = writer.buf.Flush(); err != nil {
			return fmt.Errorf("failed to flush writer for volume '%s': %w", volumeID, err)
		}
		log.Debug("Finished restoring volume '%s' to path '%s'", volumeID, restorePath)
	}

	// Close all file writers
	for _, writer := range fileWriters {
		if err := writer.file.Close(); err != nil {
			return fmt.Errorf("failed to close file writer: %w", err)
		}
		log.Debug("Closed file writer for volume '%s'", writer.file.Name())
	}
	return nil
}
