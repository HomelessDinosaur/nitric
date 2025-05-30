package python

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/nitrictech/nitric/cli/pkg/schema"
	"github.com/nitrictech/nitric/cli/pkg/sdk"
	"github.com/spf13/afero"
)

//go:embed python_proto/**/**/*.py
var protoFiles embed.FS

// GeneratePythonSDK generates Python SDK
func GeneratePythonSDK(fs afero.Fs, appSpec schema.Application, outPath string) error {
	tmpl := template.Must(template.New("storage").Parse(pythonStorageTemplate))
	data := sdk.AppSpecToTemplateData(appSpec)

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	sdkPath := filepath.Join(outPath, "sdk")

	// Create nitric/proto directory
	protoPath := filepath.Join(outPath, "proto")
	err = fs.MkdirAll(protoPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create proto directory: %w", err)
	}

	// Copy proto files
	err = copyProtoFiles(fs, protoPath)
	if err != nil {
		return fmt.Errorf("failed to copy proto files: %w", err)
	}

	storagePath := filepath.Join(sdkPath, "storage")

	err = fs.MkdirAll(storagePath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	init_files := []string{
		filepath.Join(outPath, "__init__.py"),
		filepath.Join(sdkPath, "__init__.py"),
	}

	for _, filePath := range init_files {
		if err := afero.WriteFile(fs, filePath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to write init file: %w", err)
		}
	}

	if err := afero.WriteFile(fs, filepath.Join(storagePath, "__init__.py"), buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write bucket file: %w", err)
	}

	fmt.Printf("Python SDK generated at %s\n", outPath)

	return nil
}

// copyProtoFiles recursivelycopies proto files from embedded filesystem to output directory
func copyProtoFiles(fs afero.Fs, outPath string) error {
	return walkEmbeddedFS(protoFiles, "python_proto", func(path string, isDir bool) error {
		// Skip the root directory
		if path == "python_proto" {
			return nil
		}

		// Get relative path without python_proto prefix
		targetPath := filepath.Join(outPath, path[len("python_proto/"):])

		if isDir {
			return fs.MkdirAll(targetPath, 0755)
		}

		// Read file from embedded filesystem
		data, err := protoFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read proto file %s: %w", path, err)
		}

		// Write file to output directory
		if err := afero.WriteFile(fs, targetPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write proto file %s: %w", targetPath, err)
		}

		return nil
	})
}

// walkEmbeddedFS walks through the embedded filesystem and calls the copyFile for each file
func walkEmbeddedFS(fs embed.FS, root string, copyFile func(path string, isDir bool) error) error {
	entries, err := fs.ReadDir(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			if err := copyFile(path, true); err != nil {
				return err
			}

			if err := walkEmbeddedFS(fs, path, copyFile); err != nil {
				return err
			}
		} else {
			if err := copyFile(path, false); err != nil {
				return err
			}
		}
	}

	return nil
}
