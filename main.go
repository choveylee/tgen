// Command tgen generates a backend service scaffold from the local template directory.
package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	domain := "dev.choveylee.top"
	appName := "knowledge-base-backend"

	if err := generate("template", appName, domain, appName); err != nil {
		fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}

// generate validates and initializes a private copy before publishing the project.
func generate(sourceDir, destinationDir, domain, appName string) error {
	if appName == "" || appName == "." || appName == ".." || strings.ContainsAny(appName, `/\`) {
		return errors.New("tgen: app name must be a single project name")
	}

	sourcePath, destinationPath, mode, err := copyPaths(sourceDir, destinationDir)
	if err != nil {
		return err
	}

	stagingPath, err := os.MkdirTemp(filepath.Dir(destinationPath), ".tgen-")
	if err != nil {
		return fmt.Errorf("tgen: create staging directory: %w", err)
	}

	defer os.RemoveAll(stagingPath)

	if err := copyTree(sourcePath, stagingPath); err != nil {
		return fmt.Errorf("tgen: copy template: %w", err)
	}

	if err := filepath.WalkDir(stagingPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && shouldApplyPlaceholders(path) {
			return replaceContent(path, domain, appName)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("tgen: render template: %w", err)
	}

	script := fmt.Sprintf(exportScriptTemplate, shellQuote(stagingPath), shellQuote(domain), shellQuote(appName))
	command := exec.Command("/bin/sh", "-c", script)
	command.Env = append(os.Environ(), "GOWORK=off", "CGO_ENABLED=0")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tgen: initialize generated module: %w\n%s", err, output)
	}

	if err := os.Chmod(stagingPath, mode); err != nil {
		return fmt.Errorf("tgen: set project directory permissions: %w", err)
	}

	// POSIX rename can replace an empty directory in one operation. It fails
	// if another process adds files to the destination before publication.
	if err := syscall.Rename(stagingPath, destinationPath); err != nil {
		return fmt.Errorf("tgen: publish generated project: %w", err)
	}

	return nil
}

// copyPaths resolves symlink aliases before checking that the destination is
// outside the template. A destination symlink is never a valid output target.
func copyPaths(sourceDir, destinationDir string) (string, string, fs.FileMode, error) {
	sourcePath, err := resolvePath(sourceDir)
	if err != nil {
		return "", "", 0, fmt.Errorf("tgen: resolve template directory: %w", err)
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return "", "", 0, fmt.Errorf("tgen: inspect template directory: %w", err)
	}

	if !sourceInfo.IsDir() {
		return "", "", 0, errors.New("tgen: template path must be a directory")
	}

	parentPath, err := resolvePath(filepath.Dir(destinationDir))
	if err != nil {
		return "", "", 0, fmt.Errorf("tgen: resolve project parent directory: %w", err)
	}

	destinationPath := filepath.Join(parentPath, filepath.Base(destinationDir))
	relativePath, err := filepath.Rel(sourcePath, destinationPath)
	if err != nil {
		return "", "", 0, err
	}

	if relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", "", 0, errors.New("tgen: destination must be outside the template directory")
	}

	info, err := os.Lstat(destinationPath)
	if errors.Is(err, fs.ErrNotExist) {
		return sourcePath, destinationPath, 0o755, nil
	}

	if err != nil {
		return "", "", 0, fmt.Errorf("tgen: inspect project directory: %w", err)
	}

	if !info.IsDir() {
		return "", "", 0, errors.New("tgen: project path must be a directory")
	}

	entries, err := os.ReadDir(destinationPath)
	if err != nil {
		return "", "", 0, fmt.Errorf("tgen: read project directory: %w", err)
	}

	if len(entries) != 0 {
		return "", "", 0, fmt.Errorf("tgen: destination directory %q must be empty", destinationDir)
	}

	return sourcePath, destinationPath, info.Mode().Perm(), nil
}

// resolvePath returns an absolute path with symlinks resolved.
func resolvePath(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(absolutePath)
}

// CopyDir copies the template into an absent or empty destination directory.
func CopyDir(sourceDir, destinationDir string) error {
	sourcePath, destinationPath, mode, err := copyPaths(sourceDir, destinationDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(destinationPath, mode); err != nil {
		return err
	}

	return copyTree(sourcePath, destinationPath)
}

// copyTree accepts directories and regular files and preserves permission bits.
func copyTree(sourceDir, destinationDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path == sourceDir {
			return nil
		}

		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		destinationPath := filepath.Join(destinationDir, relativePath)
		if info.IsDir() {
			return os.Mkdir(destinationPath, info.Mode().Perm())
		}

		if !info.Mode().IsRegular() {
			return fmt.Errorf("tgen: template entry %q must be a regular file or directory", relativePath)
		}

		return copyFile(path, destinationPath, info.Mode().Perm())
	})
}

// copyFile creates a new file and rejects attempts to replace an existing path.
func copyFile(sourcePath, destinationPath string, mode fs.FileMode) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}

	defer source.Close()

	destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}

	_, err = io.Copy(destination, source)

	return errors.Join(err, destination.Close())
}

// replaceContent renders placeholders and reserves the final config filename.
func replaceContent(path, domain, appName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	appCodeName := strings.ReplaceAll(appName, "-", "_")
	replacer := strings.NewReplacer(
		"{{domain}}", domain,
		"{{app_name}}", appName,
		"{{app_name2}}", appCodeName,
		"{{APP_NAME}}", strings.ToUpper(appCodeName),
	)
	content := []byte(replacer.Replace(string(data)))
	if filepath.Base(path) != "template_config.ini" {
		return os.WriteFile(path, content, 0o644)
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	configPath := filepath.Join(filepath.Dir(path), appCodeName+"_config.ini")
	config, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return err
	}

	_, err = config.Write(content)
	if err := errors.Join(err, config.Close()); err != nil {
		return err
	}

	return os.Remove(path)
}

// shouldApplyPlaceholders limits rendering to known text files.
func shouldApplyPlaceholders(path string) bool {
	if filepath.Base(path) == "Dockerfile" {
		return true
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".sql", ".sh", ".ini":
		return true
	default:
		return false
	}
}

// shellQuote makes one argument safe to interpolate into the shell command.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

const exportScriptTemplate = `#!/bin/sh
set -eu
cd %s
go mod edit -module=%s/%s
go mod tidy
`
