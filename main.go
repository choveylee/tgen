// Command tgen generates a backend service scaffold from the local template directory.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func copyFile(srcPath, destPath string) (int64, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return -1, err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return -1, err
	}

	dstFile, err := os.Create(destPath)
	if err != nil {
		return -1, err
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}

func copyTree(srcDir, destDir string) error {
	srcDir, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}

	destDir, err = filepath.Abs(destDir)
	if err != nil {
		return err
	}

	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}
	if !srcInfo.IsDir() {
		return errors.New("source path must refer to a directory")
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		destPath := filepath.Join(destDir, rel)
		if info.IsDir() {
			return os.MkdirAll(destPath, 0o755)
		}

		_, err = copyFile(path, destPath)
		return err
	})
}

// CopyDir recursively copies the contents of srcDir into destDir, creating destDir when necessary.
func CopyDir(srcDir string, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	return copyTree(srcDir, destDir)
}

func replaceContent(path string, domain, appName string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	strContent := string(data)
	strContent = strings.ReplaceAll(strContent, "{{domain}}", domain)
	strContent = strings.ReplaceAll(strContent, "{{app_name}}", appName)
	strContent = strings.ReplaceAll(strContent, "{{app_name2}}", strings.ReplaceAll(appName, "-", "_"))
	strContent = strings.ReplaceAll(strContent, "{{APP_NAME}}", strings.ToUpper(strings.ReplaceAll(appName, "-", "_")))

	outPath := path
	if filepath.Base(path) == "template_config.ini" {
		if err := os.Remove(path); err != nil {
			return err
		}
		outPath = filepath.Join(filepath.Dir(path), fmt.Sprintf("%s_config.ini", strings.ReplaceAll(appName, "-", "_")))
	}

	return os.WriteFile(outPath, []byte(strContent), 0o666)
}

// shouldApplyPlaceholders reports whether placeholder substitution should be applied to path.
// It intentionally limits substitution to known text file types so that binary assets are not modified.
func shouldApplyPlaceholders(path string) bool {
	base := filepath.Base(path)
	if base == "Dockerfile" {
		return true
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".sql", ".sh", ".ini":
		return true
	default:
		return false
	}
}

const (
	exportScriptTemplate = `#!/bin/bash
export GOPROXY=https://goproxy.cn,direct;
export GOSUMDB=off;
export CGO_ENABLED=0;
cd %s;
go mod init %s/%s;
go mod tidy;
go get gorm.io/plugin/dbresolver@v1.6.2
go get go@1.25.0
`
)

func main() {
	domain := "dev.choveylee.top"
	appName := "test-backend"

	err := CopyDir("template", appName)
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(appName, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileInfo == nil || fileInfo.IsDir() {
			return nil
		}
		if !shouldApplyPlaceholders(path) {
			return nil
		}

		return replaceContent(path, domain, appName)
	})
	if err != nil {
		panic(err)
	}

	exportScript := fmt.Sprintf(exportScriptTemplate, appName, domain, appName)

	cmd := exec.Command("/bin/sh", "-c", exportScript)

	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
}
