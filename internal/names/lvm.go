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

// LVM_DEV_PREFIX is one acceptable prefix of an LVM logical volume path (of the form /dev/<vg>/<lv>).
const LVM_DEV_PREFIX = "/dev/"

// LVM_MAPPER_PREFIX is another acceptable prefix of an LVM logical volume path (of the form /dev/mapper/<vg>-<lv>).
const LVM_MAPPER_PREFIX = "/dev/mapper/"

// LVM_MAX_NAME_LEN is the maximum length of an LVM logical volume name (of the form <lv>).
const LVM_MAX_NAME_LEN = 127

// ExtractLVMNames extracts the volume group and logical volume names from an LVM logical volume path (of the form
// /dev/<vg>/<lv> or /dev/mapper/<vg>-<lv>).
func ExtractLVMNames(p string) (vg, lv string, ok bool) {
	if strings.HasPrefix(p, LVM_MAPPER_PREFIX) {
		// /dev/mapper/<vg>-<lv>
		// Every '-' in the name is replaced with '--'. Splitting on '-' does not work, we must use the lvm tool to parse the name. Algorithm: left-to-right, upon encountering a '-', check if the next character is also a '-'. If so, skip it and continue. If not, split here.
		var vg, lv strings.Builder
		current := &vg

		p = strings.TrimPrefix(p, LVM_MAPPER_PREFIX)
		for i := 0; i < len(p); i++ {
			// Append to current name
			if p[i] != '-' {
				current.WriteByte(p[i])
				continue
			}

			// If a dash, and next character is dash, write a single dash and skip the next character
			if i+1 < len(p) && p[i+1] == '-' {
				current.WriteByte('-')
				i++
				continue
			}

			// Separator dash encountered, switch to next name. We must be in the vg name, otherwise we have two single
			// dash cases, which is invalid.
			if current == &lv {
				return "", "", false
			}
			current = &lv
		}
		vgName, lvName := vg.String(), lv.String()
		if vgName == "" || lvName == "" {
			return "", "", false
		}
		return vgName, lvName, true
	} else if strings.HasPrefix(p, LVM_DEV_PREFIX) {
		// /dev/<vg>/<lv>
		parts := strings.SplitN(strings.TrimPrefix(p, LVM_DEV_PREFIX), "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.Contains(parts[1], "/") {
			return "", "", false
		}
		return parts[0], parts[1], true
	}
	return "", "", false
}
