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
package cli

import (
	"fmt"

	"github.com/PextraCloud/pxitool/internal/restorepxi"
	"github.com/spf13/cobra"
)

var restoreVolumeMap map[string]string
var restoreConfigOutput string
var restoreEncryptionKey string
var restoreConcurrency int

var restoreCmd = &cobra.Command{
	Use:     "restore <file>",
	Args:    cobra.ExactArgs(1),
	Example: "  # Restore a file-backed volume and an LVM volume\n  pxitool restore /path/to/backup.pxi -V vol-123=/var/lib/pxi/vol-123.qcow2 -V vol-456=/dev/vg1/vol-456 -o /path/to/config/output",
	Short:   "Restore data from a backup",
	Long: "This command restores data from a previously created backup.\n\n" +
		"The -V flag maps each volume ID to its destination on this host. A\n" +
		"pre-existing destination is REPLACED (rollback); each backend uses its\n" +
		"atomic primitive so a crash leaves either the old or the new object:\n" +
		"  directory:    file path for the restored image (same QEMU format as the backup)\n" +
		"  lvm:          /dev/<vg>/<lv> to create/replace, or <vg> to reuse the source LV name in the given VG\n" +
		"  zfs:          dataset name (e.g. tank/vm-1/root), created/replaced by zfs recv\n" +
		"  rbd:          <pool>/<image> to create/replace\n" +
		"  lxc_rootfs:   directory for the rootfs (replaced wholesale if present)",
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		if restoreConcurrency < 1 {
			return fmt.Errorf("Concurrency must be at least 1")
		}

		fmt.Fprintln(cmd.ErrOrStderr(), "Restoring from file:", filePath)
		options := restorepxi.RestorePXIOptions{
			PXIPath:       filePath,
			VolumeMap:     restoreVolumeMap,
			ConfigOutput:  restoreConfigOutput,
			EncryptionKey: restoreEncryptionKey,
			Concurrency:   restoreConcurrency,
		}
		if err := restorepxi.RestorePXI(options); err != nil {
			return fmt.Errorf("Error restoring PXI file: %v", err)
		}
		return nil
	},
}

func init() {
	restoreCmd.Flags().StringToStringVarP(&restoreVolumeMap, "volume-map", "V", nil, "Specify volume ID to restore path mapping, e.g., vol-123=/dev/vg1/vol-123")

	restoreCmd.Flags().StringVarP(&restoreConfigOutput, "config-output", "o", "", "Specify the output file for the config to restore")
	restoreCmd.Flags().StringVarP(&restoreEncryptionKey, "key", "k", "", "Encryption key for the PXI file")
	restoreCmd.Flags().IntVarP(&restoreConcurrency, "concurrency", "", 2, "Number of concurrent volume restores")
	restoreCmd.MarkFlagRequired("config-output")
	restoreCmd.MarkFlagFilename("config-output", "json")

	rootCmd.AddCommand(restoreCmd)
}
