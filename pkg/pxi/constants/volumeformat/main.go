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
package volumeformat

import "fmt"

type VolumeFormat uint8

const (
	Raw VolumeFormat = iota
	QCOW2
	VMDK
)

func (v VolumeFormat) String() string {
	switch v {
	case Raw:
		return "raw"
	case QCOW2:
		return "qcow2"
	case VMDK:
		return "vmdk"
	default:
		panic(fmt.Sprintf("unknown volume format: %d", v))
	}
}

func (v VolumeFormat) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *VolumeFormat) UnmarshalText(text []byte) error {
	switch s := string(text); s {
	case "raw":
		*v = Raw
	case "qcow2":
		*v = QCOW2
	case "vmdk":
		*v = VMDK
	default:
		return fmt.Errorf("unknown volume format: %q", s)
	}
	return nil
}
