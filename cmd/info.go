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
	"encoding/json"
	"fmt"
	"os"

	"github.com/PextraCloud/pxitool/internal/readpxi"
	"github.com/PextraCloud/pxitool/pkg/log"
	"github.com/spf13/cobra"
)

var isInfoInJson bool
var skipEncryptedChunks bool

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().BoolVarP(&skipEncryptedChunks, "skip-encrypted", "s", false, `Skip encrypted chunks in the PXI file if they are present
This will not read the encrypted config and volumes`)
	infoCmd.Flags().BoolVarP(&isInfoInJson, "json", "j", false, "Output information in JSON format")
}

var infoCmd = &cobra.Command{
	Use:   "info [file]",
	Args:  cobra.ExactArgs(1),
	Short: "Display information about a Pextra Image",
	Long: `This command displays information about a
Pextra Image (.pxi) file. It checks the file structure,
verifies the integrity of the chunks, and outputs
information about the image`,
	Run: func(cmd *cobra.Command, args []string) {
		inputFileName := args[0]
		result, err := readpxi.GetInfo(inputFileName, skipEncryptedChunks)
		if err != nil {
			log.Error("Error reading PXI file: %v", err)
			os.Exit(1)
		}

		if isInfoInJson {
			var jsonData []byte
			if jsonData, err = json.MarshalIndent(result, "", "    "); err != nil {
				log.Error("Error serializing PXI info to JSON: %v", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		} else {
			log.Info("PXI File: %s", inputFileName)
			log.Info("Version=%s, InstanceType=%s, Compression=%s, Encryption=%s", result.PXIVersion, result.InstanceType, result.CompressionType, result.EncryptionType)
			log.Info("%d volumes in config, of which %d are present in the PXI file", len(result.Config.Volumes), len(result.Volumes))
			if result.Config != nil {
				log.Info("Config: can be viewed by passing the '--json' flag")
			} else {
				log.Info("Config: <nil> (encrypted chunks skipped)")
			}
		}
	},
}
