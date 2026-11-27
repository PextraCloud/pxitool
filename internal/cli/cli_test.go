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
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PextraCloud/pxitool/internal/dar"
	"github.com/PextraCloud/pxitool/internal/testutils"
)

// darStub is a shell script that simulates the behavior of the dar tool for testing purposes.
const darStub = `#!/bin/sh
for a in "$@"; do
    if [ "$a" = "--version" ]; then
        printf '  dar 3.0.0\nlibdar 2.7.2\n'
        exit 0
    fi
done
case "$*" in
  *-aheader*)
    printf 'Archive version format               : 11.3\nCompression algorithm used           : zstd\nSymmetric key encryption used        : none\n'
    exit 2
    ;;
esac
prev=
for a in "$@"; do
    if [ "$prev" = "-c" ]; then
		printf 'testdata' > "$a.1.dar"
        exit 0
    fi
	prev="$a"
done
exit 0`

func resetCLIFlags() {
	// create
	createConfigPath = ""
	createExcludeVolumesRaw = nil
	createOverwrite = false
	createEncryptionCipherString = ""
	createEncryptionKey = ""
	createCompressionAlgorithmString = ""
	createConcurrency = 1

	// info
	infoJsonOutput = false
	infoEncryptionKey = ""
	infoOnlyHeader = false

	// restore
	restoreVolumeMap = nil
	restoreConfigOutput = ""
	restoreEncryptionKey = ""
	restoreConcurrency = 1

	// root
	rootVersionFlag = false
}

type cliResult struct {
	stdout string
	stderr string
	code   int
	cmd    string
}

func runCLI(t *testing.T, workDir string, args ...string) cliResult {
	t.Helper()

	// Set up stdout and stderr capture
	out := &bytes.Buffer{}
	err := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	// Ensure args is set
	rootCmd.SetArgs(args)
	resetCLIFlags()

	// Set cwd
	if workDir != "" {
		if err := os.Chdir(workDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}
	}

	// Run and return result
	code := Execute(t.Context())
	return cliResult{stdout: out.String(), stderr: err.String(), code: code, cmd: strings.Join(args, " ")}
}

func validConfig() map[string]any {
	return map[string]any{
		"id":          "inst-1",
		"node_id":     "node-1",
		"internal_id": "int-1",
		"type":        "qemu",
		"name":        "instance-1",
		"metadata":    map[string]any{"_type": "qemu"},
		"resources": map[string]any{
			"cpu":          map[string]any{"sockets": 1, "cores": 1, "threads": 1},
			"memory_bytes": 67108864,
			"volumes":      []any{},
		},
		"creation": "2026-01-01T00:00:00Z",
	}
}

func writeConfig(t *testing.T, dir string, mutate func(map[string]any)) string {
	t.Helper()
	cfg := validConfig()
	if mutate != nil {
		mutate(cfg)
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "config.json")
	if err := os.WriteFile(p, b, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func addVolume(volume map[string]any) func(map[string]any) {
	return func(cfg map[string]any) {
		res := cfg["resources"].(map[string]any)
		res["volumes"] = []any{volume}
	}
}

func assertCode(t *testing.T, r cliResult, want int) {
	t.Helper()
	if r.code != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s", r.code, want, r.stdout, r.stderr)
	}
}

func assertStderrContains(t *testing.T, r cliResult, sub string) {
	t.Helper()
	if !strings.Contains(r.stderr, sub) {
		t.Fatalf("stderr missing %q\nstdout:\n%s\nstderr:\n%s\ncommand: %s", sub, r.stdout, r.stderr, r.cmd)
	}
}
func assertStdoutContains(t *testing.T, r cliResult, sub string) {
	t.Helper()
	if !strings.Contains(r.stdout, sub) {
		t.Fatalf("stdout missing %q\nstdout:\n%s\nstderr:\n%s\ncommand: %s", sub, r.stdout, r.stderr, r.cmd)
	}
}

func TestRootVersion(t *testing.T) {
	testutils.StubScript(t, "dar", darStub)
	for _, flag := range []string{"--version", "-v"} {
		t.Run(flag, func(t *testing.T) {
			r := runCLI(t, "", flag)
			assertCode(t, r, 0)
			assertStdoutContains(t, r, "pxitool 2.0.0 (dar 3.0.0, libdar 2.7.2)")
		})
	}
}

func TestRootVersionWithoutDar(t *testing.T) {
	r := runCLI(t, "", "--version")
	assertCode(t, r, 0)
	assertStdoutContains(t, r, "pxitool 2.0.0")
}

func TestRootHelpNoArgs(t *testing.T) {
	r := runCLI(t, "")
	assertCode(t, r, 0)
	for _, sub := range []string{"Usage:", "create", "info", "restore"} {
		assertStdoutContains(t, r, sub)
	}
}

func TestRootUnknownCommand(t *testing.T) {
	r := runCLI(t, "", "bogus")
	assertCode(t, r, 1)
	assertStderrContains(t, r, `unknown command "bogus"`)
}

func TestCreateRequiresConfigFlag(t *testing.T) {
	r := runCLI(t, "", "create")
	assertCode(t, r, 1)
	assertStderrContains(t, r, `required flag(s) "config" not set`)
}

func TestCreateConcurrencyTooLow(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, nil)
	r := runCLI(t, dir, "create", "-c", cfg, "--concurrency", "0")
	assertCode(t, r, 1)
	assertStderrContains(t, r, "concurrency must be at least 1")
}

func TestCreateInvalidCipherAndCompression(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, nil)
	for _, args := range [][]string{
		{"--cipher", "bogus", "--key", "k"},
		{"--compression", "bogus"},
	} {
		r := runCLI(t, dir, append([]string{"create", "-c", cfg}, args...)...)
		assertCode(t, r, 1)
		assertStderrContains(t, r, "Invalid")
	}
}

func TestCreateCipherWithoutKey(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, nil)
	r := runCLI(t, dir, "create", "-c", cfg, "--cipher", "aes256")
	assertCode(t, r, 1)
	assertStderrContains(t, r, "cipher specified without key")
}

func TestCreateMissingConfigFile(t *testing.T) {
	dir := t.TempDir()
	r := runCLI(t, dir, "create", "-c", filepath.Join(dir, "nope.json"))
	assertCode(t, r, 1)
	assertStderrContains(t, r, "Failed to read config file")
}

func TestCreateInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")
	if err := os.WriteFile(p, []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runCLI(t, dir, "create", "-c", p)
	assertCode(t, r, 1)
	assertStderrContains(t, r, "Failed to parse config JSON")
}

func TestCreateInvalidInstance(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, func(m map[string]any) { m["name"] = "" })
	r := runCLI(t, dir, "create", "-c", cfg)
	assertCode(t, r, 1)
	assertStderrContains(t, r, "failed to prepare PXI")
}

func TestCreateDuplicateExclude(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, addVolume(map[string]any{
		"id": "vol1", "type": "directory", "storage_pool_id": "pool-1", "path": "missing.raw", "size_bytes": 1,
	}))
	r := runCLI(t, dir, "create", "-c", cfg, "-e", "vol1", "-e", "vol1")
	assertCode(t, r, 1)
	assertStderrContains(t, r, "duplicate volume ID")
}

func TestCreateMissingVolumeSource(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, addVolume(map[string]any{
		"id": "vol1", "type": "directory", "storage_pool_id": "pool-1", "path": filepath.Join(dir, "missing.raw"), "size_bytes": 1,
	}))
	r := runCLI(t, dir, "create", "-c", cfg)
	assertCode(t, r, 1)
	assertStderrContains(t, r, "failed to back up volumes")
}

func TestCreateDefaultOutput(t *testing.T) {
	workDir := t.TempDir()
	cfg := writeConfig(t, workDir, nil)
	r := runCLI(t, workDir, "create", "-c", cfg)
	assertCode(t, r, 0)
	assertStderrContains(t, r, "Creating Pextra Image: instance-1.pxi")
	if _, err := os.Stat(filepath.Join(workDir, "instance-1.pxi.1.dar")); err != nil {
		t.Fatalf("default output missing: %v", err)
	}
}

func TestCreateHappyPathExplicitOutput(t *testing.T) {
	workDir := t.TempDir()
	cfg := writeConfig(t, workDir, nil)
	r := runCLI(t, workDir, "create", "-c", cfg, "out.pxi")
	assertCode(t, r, 0)
	assertStderrContains(t, r, "Creating Pextra Image: out.pxi")
	if _, err := os.Stat(filepath.Join(workDir, "out.pxi.1.dar")); err != nil {
		t.Fatalf("explicit output missing: %v", err)
	}
}

func TestCreateTooManyArgs(t *testing.T) {
	dir := t.TempDir()
	cfg := writeConfig(t, dir, nil)
	r := runCLI(t, dir, "create", "-c", cfg, "a.pxi", "b.pxi")
	assertCode(t, r, 1)
	assertStderrContains(t, r, "accepts at most 1 arg(s)")
}

func TestInfoRequiresOneArg(t *testing.T) {
	workDir := t.TempDir()
	r := runCLI(t, workDir, "info")
	assertCode(t, r, 1)
	assertStderrContains(t, r, "accepts 1 arg(s), received 0")
}

func TestInfoMissingTools(t *testing.T) {
	workDir := t.TempDir()
	p := filepath.Join(workDir, "some.pxi")
	if err := os.WriteFile(p, []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", "")
	r := runCLI(t, workDir, "info", p)
	assertCode(t, r, 1)
	assertStderrContains(t, r, "environment check failed")
}

func TestInfoHeader(t *testing.T) {
	testutils.StubScript(t, "dar", darStub)
	workDir := t.TempDir()
	p := filepath.Join(workDir, "some.pxi")
	if err := os.WriteFile(p, []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runCLI(t, workDir, "info", "--header", p)
	assertCode(t, r, 0)
	assertStdoutContains(t, r, "Compression Algorithm: zstd")
	assertStdoutContains(t, r, "Symmetric Encryption Algorithm: none")
}

func TestInfoHeaderJSON(t *testing.T) {
	testutils.StubScript(t, "dar", darStub)
	workDir := t.TempDir()
	p := filepath.Join(workDir, "some.pxi")
	if err := os.WriteFile(p, []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runCLI(t, workDir, "info", "--header", "--json", p)
	assertCode(t, r, 0)
	var header dar.Header
	if err := json.Unmarshal([]byte(r.stdout), &header); err != nil {
		t.Fatalf("stdout not JSON: %v\n%s", err, r.stdout)
	}
	if header.CompressionAlgorithm != "zstd" || header.SymmetricEncryptionAlgorithm != "none" || header.IsEncrypted {
		t.Fatalf("unexpected header: %+v", header)
	}
}

func TestInfoReadFailure(t *testing.T) {
	testutils.StubScript(t, "dar", darStub)
	workDir := t.TempDir()
	p := filepath.Join(workDir, "some.pxi")
	if err := os.WriteFile(p, []byte("stub"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runCLI(t, workDir, "info", p)
	assertCode(t, r, 1)
	assertStderrContains(t, r, "failed to read manifest")
}

func TestRestoreRequiresConfigOutput(t *testing.T) {
	workDir := t.TempDir()
	r := runCLI(t, workDir, "restore", "some.pxi")
	assertCode(t, r, 1)
	assertStderrContains(t, r, `required flag(s) "config-output" not set`)
}

func TestRestoreConcurrencyTooLow(t *testing.T) {
	workDir := t.TempDir()
	r := runCLI(t, workDir, "restore", "some.pxi", "-o", "out.json", "--concurrency", "0")
	assertCode(t, r, 1)
	assertStderrContains(t, r, "Concurrency must be at least 1")
}

func TestRestoreMissingFile(t *testing.T) {
	testutils.StubScript(t, "dar", darStub)
	workDir := t.TempDir()
	r := runCLI(t, workDir, "restore", filepath.Join(workDir, "missing.pxi"), "-o", "out.json")
	assertCode(t, r, 1)
	assertStderrContains(t, r, "failed to prepare to restore PXI")
	if _, err := os.Stat(filepath.Join(workDir, "out.json")); !os.IsNotExist(err) {
		t.Fatalf("config output should not be written on failure")
	}
}
