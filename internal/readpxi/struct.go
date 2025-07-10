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
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/conf"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/encr"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/iend"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/ihdr"
	"github.com/PextraCloud/pxitool/pkg/pxi/chunks/svol"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/compressiontype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/encryptiontype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/instancetype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/pxiversion"
)

type PXIChunks struct {
	IHDR *ihdr.Data   // Required
	ENCR *encr.Data   // Only if encryption indicated in IHDR
	CONF *conf.Data   // Required
	SVOL []*svol.Data // 0 to n
	IEND *iend.Data   // Required
}

type ReadPXIOutput struct {
	PXIVersion      pxiversion.PXIVersion           `json:"version"`
	InstanceType    instancetype.InstanceType       `json:"instance_type"`
	CompressionType compressiontype.CompressionType `json:"compression_type"`
	EncryptionType  encryptiontype.EncryptionType   `json:"encryption_type"`
	Config          *conf.InstanceConfigGeneric     `json:"config,omitempty"` // nil if encrypted chunks are skipped
	// List of volume IDs that are present in the PXI file as SVOL chunks
	Volumes []string `json:"volumes,omitempty"` // nil if encrypted chunks are skipped
	Path    string   `json:"path"`              // Absolute path to the PXI file
}
