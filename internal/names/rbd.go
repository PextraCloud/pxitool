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

import "strings"

// RBD_MAX_NAME_LEN is the maximum length of an RBD image name (of the form <pool>/<image>).
const RBD_MAX_NAME_LEN = 255

// ExtractRBDNames extracts the pool and image names from an RBD image reference (of the form <pool>/<image>).
func ExtractRBDNames(p string) (pool, image string, ok bool) {
	if strings.HasPrefix(p, "/") {
		return "", "", false
	}
	if len(p) > RBD_MAX_NAME_LEN {
		return "", "", false
	}

	parts := strings.SplitN(p, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.Contains(parts[1], "/") {
		return "", "", false
	}
	return parts[0], parts[1], true
}
