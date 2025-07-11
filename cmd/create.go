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

	"github.com/PextraCloud/pxitool/internal/createpxi"
	"github.com/PextraCloud/pxitool/internal/utils"
	"github.com/PextraCloud/pxitool/pkg/log"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/compressiontype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/encryptiontype"
	"github.com/spf13/cobra"
)

var jsonFilePath string
var outputFileName string
var forceOverwrite bool
var encryptionTypeString string
var excluded []string
var rootfsPath string

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&jsonFilePath, "config", "c", "", "Path to the JSON configuration file for the Pextra Image")
	createCmd.MarkFlagRequired("config")
	createCmd.MarkFlagFilename("config", "json")

	createCmd.Flags().StringVarP(&outputFileName, "output", "o", "", "Output file name for the Pextra Image (.pxi). Defaults to <name>.pxi from the config file")
	createCmd.MarkFlagFilename("output", "pxi")

	createCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Force overwrite of existing .pxi files without prompt")

	createCmd.Flags().StringVarP(&encryptionTypeString, "encryption", "e", "aes-256-gcm", "Encryption type to use for the Pextra Image (default: aes-256-gcm). Supported: aes-256-gcm, none. You will be prompted for a password if encryption is enabled.")

	createCmd.Flags().StringArrayVarP(&excluded, "exclude", "x", nil, "List of volume IDs to exclude from the Pextra Image. Can be specified multiple times.")

	createCmd.Flags().StringVarP(&rootfsPath, "rootfs", "r", "", "Path to the root filesystem for LXC instances (required). This option is ignored for other instance types.")
	createCmd.MarkFlagDirname("rootfs")
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Pextra Image (.pxi)",
	Long: `This command creates a new Pextra Image file,
based on a JSON configuration file and volume
data.`,
	Run: func(cmd *cobra.Command, args []string) {
		json, err := utils.GetConfChunkJSON(jsonFilePath)
		if err != nil {
			log.Error("Error reading JSON config: %v\n", err)
			os.Exit(1)
		}

		// Use instance name if output filename not specified
		if outputFileName == "" {
			outputFileName = json.Name + ".pxi"
		}

		file, err := utils.GetOutputFileHandle(outputFileName, forceOverwrite)
		if err != nil {
			log.Error("Error opening output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		var encryptionType encryptiontype.EncryptionType
		switch encryptionTypeString {
		case "aes-256-gcm":
			encryptionType = encryptiontype.AES256GCM
		case "none":
			encryptionType = encryptiontype.None
		default:
			log.Error("Unsupported encryption type: %s. Supported: aes-256-gcm, none.\n", encryptionTypeString)
			os.Exit(1)
		}

		err = createpxi.Create(file, json, compressiontype.None, encryptionType, excluded)
		if err != nil {
			os.Remove(outputFileName)
			log.Error("Error creating Pextra Image: %v\n", err)
			os.Exit(1)
		}
		log.Info("Pextra Image created successfully: %s\n", outputFileName)
	},
}
