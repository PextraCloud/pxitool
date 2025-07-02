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
package conf

import (
	"encoding/json"
	"fmt"

	"github.com/PextraCloud/pxitool/pkg/pxi/constants/volumetype"
)

type InstanceCPUInfo struct {
	Sockets uint8 `json:"sockets"`
	Cores   uint8 `json:"cores"`
	Threads uint8 `json:"threads"`
}

type InstanceMetadataDocker struct {
	Image   string   `json:"image"`
	Command string   `json:"command,omitempty"`
	Env     []string `json:"env,omitempty"` // Array of "key=value"
}
type InstanceMetadataLxc struct {
	Init   string `json:"init,omitempty"`
	Rootfs struct {
		StoragePoolID string `json:"storage_pool_id"`
		Image         string `json:"image"`
	} `json:"rootfs"`
}

type InstanceMetadataQemu struct {
	Arch        string `json:"arch"`
	CpuModel    string `json:"cpu_model"`
	MachineType string `json:"machine_type"`
	Firmware    string `json:"firmware"`
	Rootfs      struct {
		StoragePoolID string `json:"storage_pool_id"`
		Image         string `json:"image"`
	} `json:"rootfs"`
	RootfsPersistent bool `json:"rootfs_persistent"`
}
type InstanceMetadata struct {
	Type   string                  `json:"_type"` // "qemu", "lxc", "docker_podman"
	Docker *InstanceMetadataDocker `json:"docker,omitempty"`
	Lxc    *InstanceMetadataLxc    `json:"lxc,omitempty"`
	Qemu   *InstanceMetadataQemu   `json:"qemu,omitempty"`
}

type InstanceVolume struct {
	ID   string                `json:"id"`
	Type volumetype.VolumeType `json:"type"`
	Path string                `json:"path"`
	Size uint64                `json:"size,omitempty"` // Optional size in bytes
}

type InstanceConfig struct {
	ID          string          `json:"id,omitempty"`
	NodeID      string          `json:"node_id,omitempty"`
	Type        uint8           `json:"type"`
	InternalID  string          `json:"internal_id,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Cpu         InstanceCPUInfo `json:"cpu,omitzero"`
	Memory      uint32          `json:"memory,omitempty"` // MiB
	Metadata    struct {
		Type string         `json:"_type"`          // "qemu", "lxc", "docker_podman"
		Data map[string]any `json:"data,omitempty"` // Flexible metadata structure
	} `json:"metadata"`
	Volumes   []InstanceVolume `json:"volumes"`
	Autostart bool             `json:"autostart,omitempty"`
	BootOrder int8             `json:"boot_order,omitempty"`
	Creation  string           `json:"creation,omitempty"`
}
type InstanceConfigGeneric struct {
	InstanceConfig `json:",inline"`
	Metadata       InstanceMetadata `json:"metadata"`
}

// Metadata field is always 'metadata', schema dependent on type
// This is used for unmarshalling from JSON configuration files (from PCE API)
func (ic *InstanceConfigGeneric) UnmarshalDUnionJSON(data []byte) error {
	var metadataType struct {
		Metadata struct {
			Type string `json:"_type"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(data, &metadataType); err != nil {
		return fmt.Errorf("failed to unmarshal metadata type: %w", err)
	}

	switch metadataType.Metadata.Type {
	case "qemu":
		var tmp struct {
			InstanceConfig
			Metadata InstanceMetadataQemu `json:"metadata"`
		}
		if err := json.Unmarshal(data, &tmp); err != nil {
			return fmt.Errorf("failed to unmarshal qemu metadata: %w", err)
		}
		ic.InstanceConfig = tmp.InstanceConfig
		ic.Metadata = InstanceMetadata{
			Type: "qemu",
			Qemu: &tmp.Metadata,
		}
	case "lxc":
		var tmp struct {
			InstanceConfig
			Metadata InstanceMetadataLxc `json:"metadata"`
		}
		if err := json.Unmarshal(data, &tmp); err != nil {
			return fmt.Errorf("failed to unmarshal lxc metadata: %w", err)
		}
		ic.InstanceConfig = tmp.InstanceConfig
		ic.Metadata = InstanceMetadata{
			Type: "lxc",
			Lxc:  &tmp.Metadata,
		}
	case "docker_podman":
		var tmp struct {
			InstanceConfig
			Metadata InstanceMetadataDocker `json:"metadata"`
		}
		if err := json.Unmarshal(data, &tmp); err != nil {
			return fmt.Errorf("failed to unmarshal docker metadata: %w", err)
		}
		ic.InstanceConfig = tmp.InstanceConfig
		ic.Metadata = InstanceMetadata{
			Type:   "docker_podman",
			Docker: &tmp.Metadata,
		}
	default:
		return fmt.Errorf("unknown instance type: %s", metadataType.Metadata.Type)
	}
	return nil
}

// Metadata field name is dependent on type, schema is fixed
// This is used for unmarshalling from CONF chunks
func (ic *InstanceConfigGeneric) UnmarshalConfJSON(data []byte) error {
	var metadataType struct {
		Metadata struct {
			Type string `json:"_type"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(data, &metadataType); err != nil {
		return fmt.Errorf("failed to unmarshal metadata type: %w", err)
	}

	switch metadataType.Metadata.Type {
	case "qemu":
		var tmp struct {
			InstanceConfig
			Metadata struct {
				InstanceMetadataQemu `json:"qemu"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(data, &tmp); err != nil {
			return fmt.Errorf("failed to unmarshal qemu metadata: %w", err)
		}
		ic.InstanceConfig = tmp.InstanceConfig
		ic.Metadata = InstanceMetadata{
			Type: "qemu",
			Qemu: &tmp.Metadata.InstanceMetadataQemu,
		}
	case "lxc":
		var tmp struct {
			InstanceConfig
			Metadata struct {
				InstanceMetadataLxc `json:"lxc"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(data, &tmp); err != nil {
			return fmt.Errorf("failed to unmarshal lxc metadata: %w", err)
		}
		ic.InstanceConfig = tmp.InstanceConfig
		ic.Metadata = InstanceMetadata{
			Type: "lxc",
			Lxc:  &tmp.Metadata.InstanceMetadataLxc,
		}
	case "docker_podman":
		var tmp struct {
			InstanceConfig
			Metadata struct {
				InstanceMetadataDocker `json:"docker"`
			} `json:"metadata"`
		}
		if err := json.Unmarshal(data, &tmp); err != nil {
			return fmt.Errorf("failed to unmarshal docker metadata: %w", err)
		}
		ic.InstanceConfig = tmp.InstanceConfig
		ic.Metadata = InstanceMetadata{
			Type:   "docker_podman",
			Docker: &tmp.Metadata.InstanceMetadataDocker,
		}
	default:
		return fmt.Errorf("unknown instance type: %s", metadataType.Metadata.Type)
	}
	return nil
}
