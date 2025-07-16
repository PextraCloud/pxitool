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
package cmd

import (
	"os"

	"github.com/PextraCloud/pxitool/internal/readpxi"
	"github.com/PextraCloud/pxitool/internal/restorepxi"
	"github.com/PextraCloud/pxitool/pkg/log"
	"github.com/spf13/cobra"
)

var restorePaths map[string]string
var restoreOutputFile string

func init() {
	rootCmd.AddCommand(restoreCmd)

	restoreCmd.Flags().StringToStringVarP(&restorePaths, "paths", "p", nil, "A map of volume IDs to restore paths. Format: 'vol-xxx=path1,vol-yyy=path2,...'. Use 'rootfs' for the LXC rootfs volume ID.")
	restoreCmd.MarkFlagRequired("paths")

	restoreCmd.Flags().StringVarP(&restoreOutputFile, "config-output", "o", "", "Path to the output file where the configuration will be saved after restoration. This file will contain the restored configuration of the PXI file.")
	restoreCmd.MarkFlagRequired("config-output")
	restoreCmd.MarkFlagFilename("config-output", "json")
}

var restoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Args:  cobra.ExactArgs(1),
	Short: "Restore a Pextra Image",
	Long:  `Restore a Pextra Image (PXI) file to specified paths for each volume, and save the configuration to an output file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if restoreOutputFile == "" {
			log.Error("Config output file must be specified using --config-output flag.")
			os.Exit(1)
		}

		// Validate restorePaths
		for volumeId, restorePath := range restorePaths {
			if restorePath == "" {
				log.Error("Restore path for '%s' is empty. Please specify a valid restore path.", volumeId)
				os.Exit(1)
			}
			if volumeId == "" {
				log.Error("Volume ID '%s' is empty. Please specify a valid volume ID (or 'rootfs' for the LXC rootfs).", volumeId)
				os.Exit(1)
			}
		}

		inputFileName := args[0]
		result, err := readpxi.ReadChunks(inputFileName)
		if err != nil {
			log.Error("Error reading PXI file: %v", err)
			os.Exit(1)
		}

		log.Info("Restoring PXI file: %s", inputFileName)
		if err := restorepxi.Restore(result, restorePaths, restoreOutputFile); err != nil {
			log.Error("Error restoring PXI file: %v", err)
			os.Exit(1)
		}
		log.Info("PXI file restored successfully to specified paths.")
	},
}
