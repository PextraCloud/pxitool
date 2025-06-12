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
package instancetype

import "fmt"

type InstanceType uint8

const (
	Docker InstanceType = iota
	LXC
	QEMU
)

func (it InstanceType) String() string {
	switch it {
	case Docker:
		return "Docker"
	case LXC:
		return "LXC"
	case QEMU:
		return "QEMU"
	default:
		panic(fmt.Sprintf("Unknown InstanceType: %d", it))
	}
}
