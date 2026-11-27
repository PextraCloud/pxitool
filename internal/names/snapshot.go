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
package names

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

// UniqueSnapshotName generates a unique name for a transient snapshot object based on the volume ID, process ID, and a
// UUID. The generated name is prefixed with "pxitool-" and truncated to the specified maximum length if necessary.
func UniqueSnapshotName(volumeID string, maxLen int) string {
	if maxLen < 1 {
		return ""
	}
	name := fmt.Sprintf("pxitool-%s-%d-%s", volumeID, os.Getpid(), uuid.New().String())
	if len(name) > maxLen {
		return name[:maxLen]
	}
	return name
}
