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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/PextraCloud/pxitool/internal/createpxi"
	"github.com/PextraCloud/pxitool/pkg/pxi"
	"github.com/spf13/cobra"
)

var createConfigPath string
var createExcludeVolumesRaw []string
var createOverwrite bool
var createEncryptionCipherString string
var createEncryptionKey string
var createCompressionAlgorithmString string
var createConcurrency int

var createCmd = &cobra.Command{
	Use:     "create [file]",
	Args:    cobra.MaximumNArgs(1),
	Example: "  # Write to a file\n  pxitool create -c /path/to/config.json /path/to/output.pxi\n\n  # Write to stdout\n  pxitool create -c /path/to/config.json -",
	Short:   "Create a new Pextra Image (PXI)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if createConcurrency < 1 {
			return fmt.Errorf("concurrency must be at least 1")
		}

		if err := createpxi.CheckEnvironment(); err != nil {
			return err
		}

		var encryptionCipher pxi.EncryptionCipher
		if createEncryptionCipherString != "" {
			encryptionCipher = pxi.EncryptionCipher(createEncryptionCipherString)
			if !encryptionCipher.Valid() {
				return fmt.Errorf("Invalid encryption cipher: %s. Supported: %s.", createEncryptionCipherString, strings.Join(pxi.EncryptionCiphersStrings, ", "))
			}
		}

		var compressionAlgorithm pxi.CompressionAlgorithm
		if createCompressionAlgorithmString != "" {
			compressionAlgorithm = pxi.CompressionAlgorithm(createCompressionAlgorithmString)
			if !compressionAlgorithm.Valid() {
				return fmt.Errorf("Invalid compression algorithm: %s. Supported: %s.", createCompressionAlgorithmString, strings.Join(pxi.CompressionAlgorithmsStrings, ", "))
			}
		}

		configBytes, err := os.ReadFile(createConfigPath)
		if err != nil {
			return fmt.Errorf("Failed to read config file: %v", err)
		}

		var configJson pxi.Instance
		if err := json.Unmarshal(configBytes, &configJson); err != nil {
			return fmt.Errorf("Failed to parse config JSON: %v", err)
		}

		// Process excluded volumes
		var createExcludeVolumes []string
		for _, volume := range createExcludeVolumesRaw {
			createExcludeVolumes = append(createExcludeVolumes, strings.TrimSpace(volume))
		}

		// Use instance name if output filename not specified
		createOutputPath := ""
		if len(args) > 0 {
			createOutputPath = args[0]
		}
		if createOutputPath == "" {
			createOutputPath = configJson.Name + ".pxi"
		}

		fmt.Fprintln(cmd.ErrOrStderr(), "Creating Pextra Image:", createOutputPath)
		options := createpxi.CreatePXIOptions{
			Instance:             configJson,
			OutputPath:           createOutputPath,
			ExcludeVolumes:       createExcludeVolumes,
			Overwrite:            createOverwrite,
			EncryptionCipher:     encryptionCipher,
			EncryptionKey:        createEncryptionKey,
			CompressionAlgorithm: compressionAlgorithm,
			Concurrency:          createConcurrency,
		}
		if err := createpxi.CreatePXI(options); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createConfigPath, "config", "c", "", "Path to the JSON configuration file for the Pextra Image")
	createCmd.MarkFlagRequired("config")
	createCmd.MarkFlagFilename("config", "json")

	createCmd.Flags().StringSliceVarP(&createExcludeVolumesRaw, "exclude", "e", nil, "Specify volumes IDs to exclude from the Pextra Image")

	createCmd.Flags().StringVarP(&createEncryptionCipherString, "cipher", "", "", fmt.Sprintf("Encryption cipher to use for the Pextra Image. If not specified, no encryption will be applied. Requires --key to be specified. Supported: %s", strings.Join(pxi.EncryptionCiphersStrings, ", ")))
	createCmd.RegisterFlagCompletionFunc("cipher", cobra.FixedCompletions(pxi.EncryptionCiphersStrings, cobra.ShellCompDirectiveNoFileComp))

	createCmd.Flags().StringVarP(&createEncryptionKey, "key", "", "", "Encryption key to use for the Pextra Image. Requires --cipher to be specified")
	createCmd.MarkFlagsRequiredTogether("cipher", "key")

	createCmd.Flags().StringVarP(&createCompressionAlgorithmString, "compression", "", "zstd", fmt.Sprintf("Compression algorithm to use for the Pextra Image. Supported: %s", strings.Join(pxi.CompressionAlgorithmsStrings, ", ")))
	createCmd.RegisterFlagCompletionFunc("compression", cobra.FixedCompletions(pxi.CompressionAlgorithmsStrings, cobra.ShellCompDirectiveNoFileComp))

	createCmd.Flags().BoolVarP(&createOverwrite, "force", "f", false, "Overwrite existing .pxi files")
	createCmd.Flags().IntVarP(&createConcurrency, "concurrency", "", 2, "Number of concurrent operations for backing up volumes")

	rootCmd.AddCommand(createCmd)
}
