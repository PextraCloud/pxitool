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
	"strings"
	"testing"
	"time"
)

// validInstance is the fully-valid baseline (V1); subtests mutate a copy.
func validInstance() Instance {
	return Instance{
		Id:         "inst-1",
		NodeId:     "node-1",
		InternalId: "int-1",
		Type:       "qemu",
		Name:       "instance-1",
		Metadata:   InstanceMetadata{Type: "qemu"},
		Resources: InstanceResources{
			Cpu:         InstanceCpu{Sockets: 1, Cores: 2, Threads: 1},
			MemoryBytes: 64 * 1024 * 1024, // 64 MiB
			Volumes:     []InstanceVolume{validVolume("vol-1")},
		},
		Creation: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func validVolume(id string) InstanceVolume {
	return InstanceVolume{Id: id, Type: VolumeTypeDirectory, StoragePoolId: "pool-1", Path: "/var/lib/vm/" + id, SizeBytes: 1 * 1024 * 1024 * 1024} // 1 GiB
}

func TestInstanceValid(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Instance)
		want   string // substring of error, or "" if expecting valid
	}{
		{"minimal zero volumes", nil, ""},
		{"two volumes", func(i *Instance) {
			i.Resources.Volumes = []InstanceVolume{validVolume("vol-a"), validVolume("vol-b")}
		}, ""},
		{"empty id", func(i *Instance) { i.Id = "" }, "instance ID cannot be empty"},
		{"empty node id", func(i *Instance) { i.NodeId = "" }, "node ID cannot be empty"},
		{"empty internal id", func(i *Instance) { i.InternalId = "" }, "internal ID cannot be empty"},
		{"empty name", func(i *Instance) { i.Name = "" }, "instance name cannot be empty"},
		{"unknown metadata type", func(i *Instance) { i.Metadata.Type = "bogus" }, "unknown instance type: bogus"},
		{"empty metadata type", func(i *Instance) { i.Metadata.Type = "" }, "unknown instance type"},
		{"lxc metadata type", func(i *Instance) { i.Metadata.Type = "lxc" }, ""},
		{"sockets zero", func(i *Instance) { i.Resources.Cpu.Sockets = 0 }, "invalid sockets: 0"},
		{"sockets five", func(i *Instance) { i.Resources.Cpu.Sockets = 5 }, "invalid sockets: 5"},
		{"cores zero", func(i *Instance) { i.Resources.Cpu.Cores = 0 }, "invalid cores: 0"},
		{"threads zero", func(i *Instance) { i.Resources.Cpu.Threads = 0 }, "invalid threads: 0"},
		{"memory below 64MiB", func(i *Instance) { i.Resources.MemoryBytes = i.Resources.MemoryBytes - 1 }, "invalid memory size"},
		{"memory exactly 64MiB", func(i *Instance) { i.Resources.MemoryBytes = 64 * 1024 * 1024 }, ""},
		{"duplicate volume IDs", func(i *Instance) {
			i.Resources.Volumes = []InstanceVolume{validVolume("dup"), validVolume("dup")}
		}, "duplicate volume ID: dup"},
		{"volume missing id", func(i *Instance) {
			v := validVolume("v")
			v.Id = ""
			i.Resources.Volumes = []InstanceVolume{v}
		}, "volume ID cannot be empty"},
		{"volume missing path", func(i *Instance) {
			v := validVolume("v")
			v.Path = ""
			i.Resources.Volumes = []InstanceVolume{v}
		}, "volume path for volume v"},
		{"volume missing type", func(i *Instance) {
			v := validVolume("v")
			v.Type = ""
			i.Resources.Volumes = []InstanceVolume{v}
		}, "volume type for volume v"},
		{"volume missing storage pool", func(i *Instance) {
			v := validVolume("v")
			v.StoragePoolId = ""
			i.Resources.Volumes = []InstanceVolume{v}
		}, "storage pool ID for volume v"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ins := validInstance()
			if tc.mutate != nil {
				tc.mutate(&ins)
			}
			err := ins.Valid()
			if tc.want == "" {
				if err != nil {
					t.Fatalf("want valid, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestManifestValid(t *testing.T) {
	base := validInstance()

	t.Run("valid v2", func(t *testing.T) {
		if err := (&Manifest{Version: ManifestVersion2, Instance: base}).Valid(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})

	t.Run("unsupported version 1", func(t *testing.T) {
		err := (&Manifest{Version: 1, Instance: base}).Valid()
		if err == nil || !strings.Contains(err.Error(), "unsupported manifest version: 1 (expected 2)") {
			t.Fatalf("want version error, got %v", err)
		}
	})

	t.Run("unsupported version 3", func(t *testing.T) {
		err := (&Manifest{Version: 3, Instance: base}).Valid()
		if err == nil || !strings.Contains(err.Error(), "unsupported manifest version: 3 (expected 2)") {
			t.Fatalf("want version error, got %v", err)
		}
	})

	t.Run("invalid instance", func(t *testing.T) {
		base.Name = ""
		err := (&Manifest{Version: ManifestVersion2, Instance: base}).Valid()
		if err == nil || !strings.Contains(err.Error(), "invalid instance: instance name cannot be empty") {
			t.Fatalf("want wrapped instance error, got %v", err)
		}
	})
}

func TestCompressionAlgorithmValid(t *testing.T) {
	for _, tc := range []struct {
		alg  CompressionAlgorithm
		want bool
	}{
		{CompressionAlgorithmZstd, true},
		{CompressionAlgorithmGzip, true},
		{"lzo", false},
		{"ZSTD", false},
		{"", false},
	} {
		t.Run(string(tc.alg), func(t *testing.T) {
			if got := tc.alg.Valid(); got != tc.want {
				t.Fatalf("Valid() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestEncryptionCipherValid covers E1 for M5 (aes256 is the supported set).
func TestEncryptionCipherValid(t *testing.T) {
	for _, tc := range []struct {
		c    EncryptionCipher
		want bool
	}{
		{EncryptionCipherAES256, true},
		{"aes128", false},
		{"AES256", false},
		{"", false},
	} {
		t.Run(string(tc.c), func(t *testing.T) {
			if got := tc.c.Valid(); got != tc.want {
				t.Fatalf("Valid() = %v, want %v", got, tc.want)
			}
		})
	}
}
