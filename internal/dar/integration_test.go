//go:build integration

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
package dar

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// rawDar runs dar directly; it is the source of truth for the package.
func rawDar(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out, err := exec.Command("dar", args...).CombinedOutput()
	return string(out), err
}

func writeFixture(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func findArchiveSlices(t *testing.T, basename string) []string {
	t.Helper()
	matches, err := filepath.Glob(basename + ".*.dar")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatalf("expected an archive slice created for basename %s", basename)
	}
	return matches
}

func TestCreate(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "alpha", "b.txt": "bravo"})

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	findArchiveSlices(t, out)

	if _, err := rawDar(t, "-N", "-Q", "-t", out); err != nil {
		t.Fatalf("raw dar -t: %v", err)
	}

	output, err := rawDar(t, "-N", "-Q", "-l", out)
	if err != nil {
		t.Fatalf("raw dar -l: %v", err)
	}
	if !strings.Contains(output, "a.txt") {
		t.Errorf("archive missing a.txt:\n%s", output)
	}
	if !strings.Contains(output, "b.txt") {
		t.Errorf("archive missing b.txt:\n%s", output)
	}
}

func TestCreateDryRun(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "alpha"})

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src, DryRun: true}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if matches, _ := filepath.Glob(out + ".*.dar"); len(matches) != 0 {
		t.Fatalf("expected no archive, found %v", matches)
	}

	if output, err := rawDar(t, "-N", "-Q", "-l", out); err == nil {
		t.Fatalf("expected error for non-existent archive, got output:\n%s", output)
	}
}

func TestCreateFiles(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "alpha", "b.txt": "bravo"})

	out := filepath.Join(dir, "out.dar")
	opts := CreateOptions{OutputBasename: out, RootDir: src, Files: []string{"a.txt"}}
	if err := Create(opts); err != nil {
		t.Fatalf("Create: %v", err)
	}

	output, err := rawDar(t, "-N", "-Q", "-l", out)
	if err != nil {
		t.Fatalf("raw dar -l: %v", err)
	}
	if !strings.Contains(output, "a.txt") {
		t.Errorf("archive missing a.txt:\n%s", output)
	}
	if strings.Contains(output, "b.txt") {
		t.Errorf("archive should not contain b.txt:\n%s", output)
	}
}

func TestCreateCipher(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "secret-data"})

	out := filepath.Join(dir, "enc.dar")
	opts := CreateOptions{OutputBasename: out, RootDir: src, Cipher: "aes", Key: "secret"}
	if err := Create(opts); err != nil {
		t.Fatalf("Create: %v", err)
	}
	findArchiveSlices(t, out)

	header, err := ListHeader(ListHeaderOptions{InputBasename: out})
	if err != nil {
		t.Fatalf("ListHeader: %v", err)
	}
	if !header.IsEncrypted {
		t.Error("expected encrypted archive")
	}

	raw, _ := rawDar(t, "-N", "-Q", "-l", out, "-aheader")
	if !regexp.MustCompile(`(?i)aes`).MatchString(raw) {
		t.Errorf("raw header output missing aes:\n%s", raw)
	}
}

func TestCreateCompression(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "alpha"})

	out := filepath.Join(dir, "zstd.dar")
	opts := CreateOptions{OutputBasename: out, RootDir: src, Compression: "zstd"}
	if err := Create(opts); err != nil {
		t.Fatalf("Create: %v", err)
	}
	findArchiveSlices(t, out)

	raw, _ := rawDar(t, "-N", "-Q", "-l", out, "-aheader")
	if !strings.Contains(raw, "zstd") {
		t.Errorf("raw header output missing zstd:\n%s", raw)
	}

	output, err := rawDar(t, "-N", "-Q", "-l", out)
	if err != nil {
		t.Fatalf("raw dar -l: %v", err)
	}
	if !strings.Contains(output, "a.txt") {
		t.Errorf("archive missing a.txt:\n%s", output)
	}
}

func TestCreateOverwrite(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "alpha"})

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err == nil {
		t.Error("expected error when overwriting without -w")
	}
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src, Overwrite: true}); err != nil {
		t.Errorf("Create with overwrite: %v", err)
	}
}

func TestCreateStdout(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "alpha"})

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	err = Create(CreateOptions{OutputBasename: "-", RootDir: src})
	w.Close()
	os.Stdout = old
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	archive, _ := io.ReadAll(r)
	if len(archive) == 0 {
		t.Error("expected archive on stdout")
	}
}

func TestCreateInvalidCipher(t *testing.T) {
	if err := Create(CreateOptions{OutputBasename: "out.dar", RootDir: t.TempDir(), Key: "secret"}); err == nil {
		t.Error("expected key-without-cipher error")
	}
	if err := Create(CreateOptions{OutputBasename: "out.dar", RootDir: t.TempDir(), Cipher: "aes"}); err == nil {
		t.Error("expected cipher-without-key error")
	}
}

func TestParseHeaderMissingCompression(t *testing.T) {
	if _, err := parseHeader("Symmetric key encryption used : none\n"); err == nil {
		t.Error("expected compression-missing error")
	}
}

func TestExtract(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFixture(t, src, map[string]string{"a.txt": "alpha", "b.txt": "bravo"})
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := Extract(ExtractOptions{InputBasename: out, RootDir: dst}); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatalf("read restored a.txt: %v", err)
	}
	if string(got) != "alpha" {
		t.Errorf("a.txt = %q, want alpha", got)
	}
}

func TestExtractFiles(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFixture(t, src, map[string]string{"a.txt": "alpha", "b.txt": "bravo"})
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	err := Extract(ExtractOptions{InputBasename: out, RootDir: dst, Files: []string{"a.txt"}})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.txt")); err != nil {
		t.Errorf("expected a.txt restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "b.txt")); !os.IsNotExist(err) {
		t.Errorf("expected b.txt absent, stat err = %v", err)
	}
}

func TestExtractDryRun(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFixture(t, src, map[string]string{"a.txt": "alpha"})
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := Extract(ExtractOptions{InputBasename: out, RootDir: dst, DryRun: true}); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.txt")); !os.IsNotExist(err) {
		t.Errorf("dry run should not restore files, stat err = %v", err)
	}
}

func TestExtractOverwrite(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFixture(t, src, map[string]string{"a.txt": "alpha"})
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, "a.txt"), []byte("OLD"), 0644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := Extract(ExtractOptions{InputBasename: out, RootDir: dst}); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	got, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(got) != "OLD" {
		t.Errorf("expected existing file kept without -w, got %q", got)
	}

	if err := Extract(ExtractOptions{InputBasename: out, RootDir: dst, Overwrite: true}); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	got, _ = os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(got) != "alpha" {
		t.Errorf("got %q, want alpha", got)
	}
}

func TestExtractCipher(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	writeFixture(t, src, map[string]string{"a.txt": "secret-data"})
	if err := os.MkdirAll(dst, 0755); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "enc.dar")
	opts := CreateOptions{OutputBasename: out, RootDir: src, Cipher: "aes", Key: "secret"}
	if err := Create(opts); err != nil {
		t.Fatalf("Create: %v", err)
	}

	err := Extract(ExtractOptions{InputBasename: out, RootDir: dst, Cipher: "aes", Key: "secret"})
	if err != nil {
		t.Fatalf("Extract: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatalf("read restored a.txt: %v", err)
	}
	if string(got) != "secret-data" {
		t.Errorf("a.txt = %q, want secret-data", got)
	}

	dst2 := filepath.Join(dir, "dst2")
	if err := os.MkdirAll(dst2, 0755); err != nil {
		t.Fatal(err)
	}
	err = Extract(ExtractOptions{InputBasename: out, RootDir: dst2, Cipher: "aes", Key: "wrong"})
	if err == nil {
		t.Error("expected error with wrong key")
	}
}

func TestListHeader(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "alpha"})

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	header, err := ListHeader(ListHeaderOptions{InputBasename: out})
	if err != nil {
		t.Fatalf("ListHeader: %v", err)
	}
	if header.IsEncrypted {
		t.Error("expected unencrypted archive")
	}
	if header.CompressionAlgorithm == "" {
		t.Error("expected compression algorithm")
	}

	raw, _ := rawDar(t, "-N", "-Q", "-l", out, "-aheader")
	if !strings.Contains(raw, header.CompressionAlgorithm) {
		t.Errorf("raw header does not contain parsed compression algorithm %q:\n%s", header.CompressionAlgorithm, raw)
	}
}

func TestListHeaderInvalid(t *testing.T) {
	p := filepath.Join(t.TempDir(), "not-a-dar.txt")
	if err := os.WriteFile(p, []byte("not a dar archive"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ListHeader(ListHeaderOptions{InputBasename: p}); err == nil {
		t.Error("expected error for non-dar input")
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	writeFixture(t, src, map[string]string{"a.txt": "alpha", "b.txt": "bravo"})

	out := filepath.Join(dir, "out.dar")
	if err := Create(CreateOptions{OutputBasename: out, RootDir: src}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	names, err := List(ListOptions{InputBasename: out})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(names) == 0 {
		t.Fatal("expected names from List")
	}

	if !slices.Contains(names, "a.txt") {
		t.Errorf("archive missing a.txt: %v", names)
	}
	if !slices.Contains(names, "b.txt") {
		t.Errorf("archive missing b.txt: %v", names)
	}

	raw, err := rawDar(t, "-N", "-Q", "--list-format=xml", "-l", out)
	if err != nil {
		t.Fatalf("raw dar -l: %v", err)
	}

	if !strings.Contains(raw, "<File name=\"a.txt\"") {
		t.Errorf("raw dar output missing a.txt:\n%s", raw)
	}
	if !strings.Contains(raw, "<File name=\"b.txt\"") {
		t.Errorf("raw dar output missing b.txt:\n%s", raw)
	}
}

func TestVersion(t *testing.T) {
	v, err := Version()
	if err != nil {
		t.Fatalf("Version: %v", err)
	}

	raw, _ := rawDar(t, "--version")
	wantDar := strings.TrimSpace(strings.Join(darVersionRegex.FindAllString(raw, -1), "\n"))
	wantLibDar := strings.TrimSpace(strings.Join(libdarVersionRegex.FindAllString(raw, -1), "\n"))
	if v.DarVersion != wantDar {
		t.Errorf("DarVersion = %q, want %q from raw dar --version", v.DarVersion, wantDar)
	}
	if v.LibDarVersion != wantLibDar {
		t.Errorf("LibDarVersion = %q, want %q from raw dar --version", v.LibDarVersion, wantLibDar)
	}
	if !regexp.MustCompile(`\d+\.\d+\.\d+`).MatchString(v.LibDarVersion) {
		t.Errorf("LibDarVersion missing semver: %q", v.LibDarVersion)
	}
}

func TestVersionUnavailable(t *testing.T) {
	t.Setenv("PATH", "")
	if _, err := Version(); err == nil {
		t.Error("expected error when dar is unavailable")
	}
}
