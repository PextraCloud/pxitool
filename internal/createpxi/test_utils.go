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
	"time"

	"github.com/PextraCloud/pxitool/pkg/pxi"
)

func testInstance(volumes ...pxi.InstanceVolume) pxi.Instance {
	if volumes == nil {
		volumes = []pxi.InstanceVolume{}
	}
	return pxi.Instance{
		Id:         "instance-1",
		NodeId:     "node-1",
		InternalId: "internal-1",
		Type:       "qemu",
		Name:       "test-instance",
		Metadata:   pxi.InstanceMetadata{Type: "qemu", Data: map[string]any{}},
		Resources: pxi.InstanceResources{
			Cpu:         pxi.InstanceCpu{Sockets: 1, Cores: 2, Threads: 4},
			MemoryBytes: 64 * 1024 * 1024,
			Volumes:     volumes,
		},
		Creation: time.Now(),
	}
}

func testVolume(id, sourcePath string) pxi.InstanceVolume {
	return pxi.InstanceVolume{
		Id:            id,
		Type:          pxi.VolumeTypeDirectory,
		StoragePoolId: "pool-1",
		Path:          sourcePath,
		SizeBytes:     1024,
	}
}