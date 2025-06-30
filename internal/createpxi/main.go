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
package createpxi

import (
	"fmt"
	"io"
	"os"

	"github.com/PextraCloud/pxitool/internal/backup"
	"github.com/PextraCloud/pxitool/internal/encryption"
	"github.com/PextraCloud/pxitool/internal/utils"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/conf"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/iend"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/ihdr"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/svol"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/compressiontype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/encryptiontype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/instancetype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/pxiversion"
	"github.com/PextraCloud/pxitool/pkg/pxi/signature"
)

func Create(file *os.File, config *conf.InstanceConfigGeneric, compressionType compressiontype.CompressionType, encryptionType encryptiontype.EncryptionType, excludedVolumes []string) error {
	var writer io.Writer = file
	var err error

	// Write magic number
	if _, err = writer.Write(signature.PXISignature); err != nil {
		return fmt.Errorf("failed to write PXI signature: %v", err)
	}
	// Write IHDR chunk
	ihdrChunk := ihdr.New(pxiversion.V1, instancetype.InstanceType(config.Type), compressionType, encryptionType)
	if err = writeChunk(writer, &ihdrChunk.Chunk); err != nil {
		return err
	}

	var encryptedWriter *encryption.EncryptedWriter
	if encryptionType != encryptiontype.None {
		encryptedWriter, err = getEncryptedWriter(file)
		if err != nil {
			return fmt.Errorf("failed to get encrypted writer: %v", err)
		}

		writer = encryptedWriter
	}

	// Write CONF chunk
	var confChunk *conf.CONF
	if confChunk, err = conf.New(config); err != nil {
		return fmt.Errorf("failed to create CONF chunk: %v", err)
	}
	if err = writeChunk(writer, &confChunk.Chunk); err != nil {
		return fmt.Errorf("failed to write CONF chunk: %v", err)
	}

	// Write volumes, excluding specified ones
	volumes := utils.GetVolumePathsFromConfig(config, excludedVolumes)
	fmt.Printf("Backing up %d volumes...\n", len(volumes))
	for i, volumePath := range volumes {
		// Save before position seek
		// Save current file position to later update SVOL chunk length
		var startPos int64
		if seeker, ok := writer.(io.Seeker); ok {
			if startPos, err = seeker.Seek(0, io.SeekCurrent); err != nil {
				return fmt.Errorf("failed to get current file position: %v", err)
			}
		}

		svolChunk := svol.New(config.Volumes[i].Type, config.Volumes[i].ID)
		if err = writeChunk(writer, &svolChunk.Chunk); err != nil {
			return fmt.Errorf("failed to write SVOL chunk for volume %s: %v", volumePath, err)
		}

		fmt.Printf("Backing up volume %d/%d: %s\n", i+1, len(volumes), volumePath)
		var bytesWritten int64
		if bytesWritten, err = backup.BackupVolume(volumePath, config.Volumes[i].Type, writer); err != nil {
			return fmt.Errorf("failed to backup volume %s: %v", volumePath, err)
		}
		fmt.Printf("Volume %d/%d backed up successfully (%d bytes written).\n", i+1, len(volumes), bytesWritten)

		// Increment SVOL chunk length
		svol.IncrementLength(svolChunk, uint64(bytesWritten))

		// Change the length of the SVOL chunk in the writer (todo)
		if seeker, ok := writer.(io.Seeker); ok {
			if _, err = seeker.Seek(startPos, io.SeekStart); err != nil {
				return fmt.Errorf("failed to seek back to SVOL chunk start position: %v", err)
			}
			if err = writeChunk(writer, &svolChunk.Chunk); err != nil {
				return fmt.Errorf("failed to update SVOL chunk length: %v", err)
			}
			if _, err = seeker.Seek(0, io.SeekEnd); err != nil {
				return fmt.Errorf("failed to seek back to end of file after updating SVOL chunk: %v", err)
			}
		}
		println("Volume backup completed successfully.")
	}

	// Write IEND chunk
	iendChunk := iend.New()
	if err = writeChunk(writer, &iendChunk.Chunk); err != nil {
		return fmt.Errorf("failed to write IEND chunk: %v", err)
	}

	// Flush buffers if using encrypted/compressed writer
	if encryptedWriter != nil {
		println("Flushing encrypted writer buffers...")
		if err := encryptedWriter.Close(); err != nil {
			return fmt.Errorf("failed to close encrypted writer: %v", err)
		}
	}
	if closer, ok := writer.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("failed to close writer: %v", err)
		}
	}
	return nil
}
