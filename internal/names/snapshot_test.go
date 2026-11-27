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
	"strings"
	"testing"
)

func TestUniqueSnapshotName(t *testing.T) {
	t.Run("prefix and full for large maxLen", func(t *testing.T) {
		name := UniqueSnapshotName("vid", 1000)
		if !strings.HasPrefix(name, "pxitool-vid-") {
			t.Errorf("name %q does not start with pxitool-vid-", name)
		}
		if !strings.Contains(name, fmt.Sprintf("-%d-", os.Getpid())) {
			t.Errorf("name %q does not contain pid", name)
		}
	})

	t.Run("truncated to maxLen", func(t *testing.T) {
		name := UniqueSnapshotName("vid", 20)
		if len(name) != 20 {
			t.Errorf("len(name) = %d, want 20", len(name))
		}
	})

	t.Run("truncated with maxLen smaller than prefix", func(t *testing.T) {
		name := UniqueSnapshotName("vid", 4)
		if len(name) != 4 {
			t.Errorf("len(name) = %d, want 4", len(name))
		}
	})

	t.Run("maxLen of zero returns empty", func(t *testing.T) {
		if name := UniqueSnapshotName("vid", 0); name != "" {
			t.Errorf("UniqueSnapshotName(vid, 0) = %q, want empty", name)
		}
	})

	t.Run("negative maxLen returns empty", func(t *testing.T) {
		if name := UniqueSnapshotName("vid", -1); name != "" {
			t.Errorf("UniqueSnapshotName(vid, -1) = %q, want empty", name)
		}
	})

	t.Run("names are unique", func(t *testing.T) {
		if UniqueSnapshotName("vid", 100) == UniqueSnapshotName("vid", 100) {
			t.Error("two generated names are equal")
		}
	})
}
