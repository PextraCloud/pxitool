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
	"strings"
)

// ZFS_ZVOL_PREFIX is the block device prefix of a ZFS zvol.
const ZFS_ZVOL_PREFIX = "/dev/zvol/"

// ZFS_MAX_NAME_LEN is the maximum length of a ZFS zvol (of the form <pool>/<dataset>).
const ZFS_MAX_NAME_LEN = 254

// ExtractZFSName extracts the ZFS zvol name (of the form <pool>/<dataset>) from a device path of the form
// /dev/zvol/<pool>/<dataset>.
func ExtractZFSName(p string) (string, bool) {
	name := strings.TrimPrefix(p, ZFS_ZVOL_PREFIX)
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", false
	}
	full := fmt.Sprintf("%s/%s", parts[0], parts[1])
	if len(full) > ZFS_MAX_NAME_LEN {
		return "", false
	}
	return full, true
}
