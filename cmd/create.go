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
	"fmt"
	"os"

	"github.com/PextraCloud/pxitool/internal/createpxi"
	"github.com/PextraCloud/pxitool/internal/utils"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/compressiontype"
	"github.com/PextraCloud/pxitool/pkg/pxi/constants/encryptiontype"
	"github.com/spf13/cobra"
)

var jsonFilePath string
var outputFileName string
var forceOverwrite bool
var encryptionTypeString string

// volume id -> volume path
var volumePaths map[string]string

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVarP(&jsonFilePath, "config", "c", "", "Path to the JSON configuration file for the Pextra Image")
	createCmd.MarkFlagRequired("config")
	createCmd.MarkFlagFilename("config", "json")

	createCmd.Flags().StringVarP(&outputFileName, "output", "o", "", "Output file name for the Pextra Image (.pxi). Defaults to <name>.pxi from the config file")
	createCmd.MarkFlagFilename("output", "pxi")

	createCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Force overwrite of existing .pxi files without prompt")

	createCmd.Flags().StringVarP(&encryptionTypeString, "encryption", "e", "aes-256-gcm", "Encryption type to use for the Pextra Image (default: aes-256-gcm). Supported: aes-256-gcm, none. You will be prompted for a password if encryption is enabled.")

	createCmd.Flags().StringToStringVarP(&volumePaths, "volumes", "v", nil, "Map of volume IDs to paths (e.g., --volumes id1=/dev/zvol/pool/vol1,id2=/mnt/data/vol.qcow2)")
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
			fmt.Fprintf(os.Stderr, "Error reading JSON config: %v\n", err)
			os.Exit(1)
		}

		// Use instance name if output filename not specified
		if outputFileName == "" {
			outputFileName = json.Name + ".pxi"
		}

		file, err := utils.GetOutputFileHandle(outputFileName, forceOverwrite)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening output file: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "Unsupported encryption type: %s. Supported: aes-256-gcm, none.\n", encryptionTypeString)
			os.Exit(1)
		}

		err = createpxi.Create(file, json, compressiontype.None, encryptionType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Pextra Image: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Pextra Image created successfully: %s\n", outputFileName)
	},
}
