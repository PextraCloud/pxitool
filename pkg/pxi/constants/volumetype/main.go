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
package volumetype

import "fmt"

type VolumeType uint8

const (
	Directory VolumeType = iota
	ISCSI
	LVM
	NetFS
	RBD
	ZFS
)

func (v VolumeType) String() string {
	switch v {
	case Directory:
		return "directory"
	case ISCSI:
		return "iscsi"
	case LVM:
		return "lvm"
	case NetFS:
		return "netfs"
	case RBD:
		return "rbd"
	case ZFS:
		return "zfs"
	default:
		panic(fmt.Sprintf("unknown volume type: %d", v))
	}
}

func (v VolumeType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *VolumeType) UnmarshalText(text []byte) error {
	switch s := string(text); s {
	case "directory":
		*v = Directory
	case "iscsi":
		*v = ISCSI
	case "lvm":
		*v = LVM
	case "netfs":
		*v = NetFS
	case "rbd":
		*v = RBD
	case "zfs":
		*v = ZFS
	default:
		return fmt.Errorf("unknown volume type: %q", s)
	}
	return nil
}
