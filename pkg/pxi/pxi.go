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
package pxi

import (
	"fmt"
	"time"
)

// ManifestFileName is the name of the manifest file inside a PXI archive.
const ManifestFileName = "pxi.json"

// ManifestVersion represents the version of the PXI manifest format.
type ManifestVersion int

const (
	// ManifestVersion2 is the second version of the PXI manifest format.
	ManifestVersion2 ManifestVersion = 2
	// ManifestVersionLatest is the latest supported manifest version.
	ManifestVersionLatest ManifestVersion = ManifestVersion2
)

// InstanceCpu represents the CPU configuration of an instance.
type InstanceCpu struct {
	Sockets int `json:"sockets"`
	Cores   int `json:"cores"`
	Threads int `json:"threads"`
}

// Valid checks if the InstanceCpu struct has valid values.
func (c *InstanceCpu) Valid() error {
	if c.Sockets < 1 || c.Sockets > 4 {
		return fmt.Errorf("invalid sockets: %d (must be between 1 and 4)", c.Sockets)
	}
	if c.Cores < 1 {
		return fmt.Errorf("invalid cores: %d (must be at least 1)", c.Cores)
	}
	if c.Threads < 1 {
		return fmt.Errorf("invalid threads: %d (must be at least 1)", c.Threads)
	}
	return nil
}

// InstanceVolume represents a single storage volume belonging to the instance.
type InstanceVolume struct {
	Id            string     `json:"id"`
	Type          VolumeType `json:"type"`
	StoragePoolId string     `json:"storage_pool_id"`
	Path          string     `json:"path"`
	SizeBytes     uint64     `json:"size_bytes"`
	Exclude       bool       `json:"exclude,omitempty"`
}

// Valid checks if the InstanceVolume struct has valid values.
func (v *InstanceVolume) Valid() error {
	if v.Id == "" {
		return fmt.Errorf("volume ID cannot be empty")
	}
	if v.Path == "" {
		return fmt.Errorf("volume path for volume %s cannot be empty", v.Id)
	}
	if !v.Type.Valid() {
		return fmt.Errorf("volume type for volume %s is invalid", v.Id)
	}
	if v.StoragePoolId == "" {
		return fmt.Errorf("storage pool ID for volume %s cannot be empty", v.Id)
	}
	if v.SizeBytes == 0 {
		return fmt.Errorf("volume size for volume %s cannot be zero", v.Id)
	}
	// TODO: validate type

	return nil
}

// InstanceResources represents the resources allocated to an instance.
type InstanceResources struct {
	Cpu         InstanceCpu      `json:"cpu"`
	MemoryBytes uint64           `json:"memory_bytes"`
	Volumes     []InstanceVolume `json:"volumes"`
}

// Valid checks if the InstanceResources struct has valid values.
func (r *InstanceResources) Valid() error {
	if err := r.Cpu.Valid(); err != nil {
		return fmt.Errorf("invalid CPU configuration: %w", err)
	}

	const minBytes = 64 * 1024 * 1024 // 64MiB
	if r.MemoryBytes < minBytes {
		return fmt.Errorf("invalid memory size: %d bytes (must be at least %d bytes)", r.MemoryBytes, minBytes)
	}

	if r.Volumes == nil {
		return fmt.Errorf("volumes cannot be nil")
	}
	// Validate volumes and check for duplicate IDs
	set := make(map[string]struct{})
	for _, volume := range r.Volumes {
		if err := volume.Valid(); err != nil {
			return fmt.Errorf("invalid volume: %w", err)
		}
		if _, exists := set[volume.Id]; exists {
			return fmt.Errorf("duplicate volume ID: %s", volume.Id)
		}
		set[volume.Id] = struct{}{}
	}
	return nil
}

// InstanceMetadata represents the metadata of an instance, which varies based on the instance's type.
type InstanceMetadata struct {
	Type string         `json:"_type"` // "qemu", "lxc" TODO
	Data map[string]any `json:"data"`
}

// Valid checks if the InstanceMetadata struct has valid values.
func (m *InstanceMetadata) Valid() error {
	switch m.Type {
	case "qemu", "lxc":
		return nil
	default:
		return fmt.Errorf("unknown instance type: %s", m.Type)
	}
}

// Instance represents a virtual machine or container instance with its configuration and metadata.
type Instance struct {
	Id          string            `json:"id"`
	NodeId      string            `json:"node_id"`
	InternalId  string            `json:"internal_id"`
	Type        string            `json:"type"` // "qemu", "lxc" TODO
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Metadata    InstanceMetadata  `json:"metadata"`
	Resources   InstanceResources `json:"resources"`
	Autostart   bool              `json:"autostart"`
	BootOrder   int               `json:"boot_order"`
	Creation    time.Time         `json:"creation"`
}

// Valid checks if the Instance struct has valid values.
func (i *Instance) Valid() error {
	if i.Id == "" {
		return fmt.Errorf("instance ID cannot be empty")
	}
	if i.NodeId == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if i.InternalId == "" {
		return fmt.Errorf("internal ID cannot be empty")
	}
	if i.Name == "" {
		return fmt.Errorf("instance name cannot be empty")
	}
	if err := i.Metadata.Valid(); err != nil {
		return fmt.Errorf("invalid metadata: %w", err)
	}
	if err := i.Resources.Valid(); err != nil {
		return fmt.Errorf("invalid resources: %w", err)
	}

	return nil
}

// Manifest represents the manifest of a PXI archive.
type Manifest struct {
	// Version is the version of the manifest format.
	Version ManifestVersion `json:"version"`
	// CreatedBy is the name and version of the tools used to create the manifest. Optional.
	//
	// Example:
	//   "pxitool/1.0.0 (dar 2.8.6, libdar 2.8.6)"
	CreatedBy string `json:"created_by,omitempty"`
	// Instance is the instance configuration and metadata.
	Instance Instance `json:"instance"`
}

// Valid checks if the Manifest struct has valid values.
func (m *Manifest) Valid() error {
	if m.Version != ManifestVersion2 {
		return fmt.Errorf("unsupported manifest version: %d (expected %d)", m.Version, ManifestVersion2)
	}
	if err := m.Instance.Valid(); err != nil {
		return fmt.Errorf("invalid instance: %w", err)
	}
	return nil
}

// String returns a human-readable representation of the Manifest struct.
func (m *Manifest) String() string {
	return fmt.Sprintf("Version: %d\nCreatedBy: %s\nInstance: %+v", m.Version, m.CreatedBy, m.Instance)
}
