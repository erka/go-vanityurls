package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
  /sdk:
    repo: https://github.com/example/sdk
    vcs: git
  /contrib:
    repo: https://github.com/example/contrib
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	// Check index.html exists
	indexPath := filepath.Join(tmpDir, "index.html")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read index.html: %v", err)
	}

	indexStr := string(indexContent)
	if !strings.Contains(indexStr, "go.example.com") {
		t.Error("index.html should contain host")
	}
	if !strings.Contains(indexStr, "go.example.com/contrib") {
		t.Error("index.html should contain /contrib handler")
	}
	if !strings.Contains(indexStr, "go.example.com/sdk") {
		t.Error("index.html should contain /sdk handler")
	}

	// Check generated vanity files
	sdkPath := filepath.Join(tmpDir, "sdk.html")
	if _, err := os.Stat(sdkPath); err != nil {
		t.Fatalf("sdk.html not created: %v", err)
	}

	contribPath := filepath.Join(tmpDir, "contrib.html")
	if _, err := os.Stat(contribPath); err != nil {
		t.Fatalf("contrib.html not created: %v", err)
	}
}

func TestGenerateVanityMetaTags(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
  /mylib:
    repo: https://github.com/example/mylib
    vcs: git
    display: https://github.com/example/mylib https://github.com/example/mylib/tree/main{/dir} https://github.com/example/mylib/blob/main{/dir}/{file}#L{line}
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	mylibPath := filepath.Join(tmpDir, "mylib.html")
	content, err := os.ReadFile(mylibPath)
	if err != nil {
		t.Fatalf("failed to read mylib.html: %v", err)
	}

	contentStr := string(content)

	// Check go-import meta tag
	if !strings.Contains(contentStr, `<meta name="go-import" content="go.example.com/mylib git https://github.com/example/mylib">`) {
		t.Error("go-import meta tag is incorrect")
	}

	// Check go-source meta tag with custom display
	if !strings.Contains(contentStr, `<meta name="go-source" content="go.example.com/mylib https://github.com/example/mylib https://github.com/example/mylib/tree/main{/dir} https://github.com/example/mylib/blob/main{/dir}/{file}#L{line}">`) {
		t.Error("go-source meta tag with custom display is incorrect")
	}

	// Check refresh meta tag
	if !strings.Contains(contentStr, `<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/go.example.com/mylib/">`) {
		t.Error("refresh meta tag is incorrect")
	}
}

func TestGenerateDefaultDisplay(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
  /lib:
    repo: https://github.com/example/lib
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	libPath := filepath.Join(tmpDir, "lib.html")
	content, err := os.ReadFile(libPath)
	if err != nil {
		t.Fatalf("failed to read lib.html: %v", err)
	}

	contentStr := string(content)

	// Check that default display is generated correctly
	expectedDisplay := "https://github.com/example/lib https://github.com/example/lib/tree/main{/dir} https://github.com/example/lib/blob/main{/dir}/{file}#L{line}"
	if !strings.Contains(contentStr, expectedDisplay) {
		t.Errorf("default display format incorrect. Expected substring: %s\nGot: %s", expectedDisplay, contentStr)
	}
}

func TestGenerateDefaultVCS(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
  /lib:
    repo: https://github.com/example/lib
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	libPath := filepath.Join(tmpDir, "lib.html")
	content, err := os.ReadFile(libPath)
	if err != nil {
		t.Fatalf("failed to read lib.html: %v", err)
	}

	contentStr := string(content)

	// Check that default VCS is 'git'
	if !strings.Contains(contentStr, `<meta name="go-import" content="go.example.com/lib git https://github.com/example/lib">`) {
		t.Error("default VCS should be 'git'")
	}
}

func TestGenerateTrailingSlashes(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
  /lib/:
    repo: https://github.com/example/lib
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	// Should create lib.html (with trailing slash trimmed)
	libPath := filepath.Join(tmpDir, "lib.html")
	if _, err := os.Stat(libPath); err != nil {
		t.Fatalf("lib.html not created (trailing slash should be trimmed): %v", err)
	}

	content, err := os.ReadFile(libPath)
	if err != nil {
		t.Fatalf("failed to read lib.html: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, `<meta name="go-import" content="go.example.com/lib git https://github.com/example/lib">`) {
		t.Error("path with trailing slash should be normalized")
	}
}

func TestGenerateMultiplePaths(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
  /alpha:
    repo: https://github.com/example/alpha
  /beta:
    repo: https://github.com/example/beta
  /zeta:
    repo: https://github.com/example/zeta
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	indexPath := filepath.Join(tmpDir, "index.html")
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read index.html: %v", err)
	}

	contentStr := string(content)

	// Verify all paths are in index
	for _, path := range []string{"alpha", "beta", "zeta"} {
		if !strings.Contains(contentStr, "go.example.com/"+path) {
			t.Errorf("index.html should contain /"+path, path)
		}
	}

	// Verify handlers are sorted alphabetically
	alphaIdx := strings.Index(contentStr, "go.example.com/alpha")
	betaIdx := strings.Index(contentStr, "go.example.com/beta")
	zetaIdx := strings.Index(contentStr, "go.example.com/zeta")

	if alphaIdx > betaIdx || betaIdx > zetaIdx {
		t.Error("handlers should be sorted alphabetically")
	}
}

func TestGenerateInvalidConfig(t *testing.T) {
	input := []byte(`
invalid yaml: [
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err == nil {
		t.Error("generate should fail with invalid YAML")
	}
}

func TestGenerateEmptyConfig(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err != nil {
		t.Fatalf("generate should handle empty paths: %v", err)
	}

	indexPath := filepath.Join(tmpDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		t.Fatalf("index.html should be created even with empty paths: %v", err)
	}
}

func TestGenerateFilePermissions(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
  /lib:
    repo: https://github.com/example/lib
`)

	tmpDir := t.TempDir()
	err := generate(input, tmpDir)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	indexPath := filepath.Join(tmpDir, "index.html")
	stat, err := os.Stat(indexPath)
	if err != nil {
		t.Fatalf("failed to stat index.html: %v", err)
	}

	// Check file permissions are restrictive (0600)
	perm := stat.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("index.html permissions should be 0600, got %o", perm)
	}
}

func TestGenerateDirectoryPermissions(t *testing.T) {
	input := []byte(`
host: go.example.com
paths:
  /lib:
    repo: https://github.com/example/lib
`)

	// Create a specific test directory
	baseDir := t.TempDir()
	outputDir := filepath.Join(baseDir, "output")

	err := generate(input, outputDir)
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	stat, err := os.Stat(outputDir)
	if err != nil {
		t.Fatalf("failed to stat output dir: %v", err)
	}

	// Check directory permissions are restrictive (0700)
	perm := stat.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("output directory permissions should be 0700, got %o", perm)
	}
}
