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

	dstFile, err := os.Create(destPath)
	if err != nil {
		return -1, err
	}

	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}

func mkDir(destDir string) error {
	_, err := os.Stat(destDir)
	if err == nil {
		return nil
	}

	if os.IsNotExist(err) {
		err := os.Mkdir(destDir, os.ModePerm)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func copyDir(srcDir string, destDir string) error {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}

	if !srcInfo.IsDir() {
		return errors.New("src dir type illegal")
	}

	destInfo, err := os.Stat(destDir)
	if err != nil {
		return err
	}

	if !destInfo.IsDir() {
		return errors.New("des dir type illegal")
	}

	srcDir, err = filepath.Abs(srcDir)
	if err != nil {
		return err
	}

	destDir, err = filepath.Abs(destDir)
	if err != nil {
		return err
	}

	err = filepath.Walk(srcDir, func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo == nil {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		if fileInfo.IsDir() {
			err := mkDir(destPath)
			if err != nil {
				return err
			}

			err = copyDir(path, destPath)
			if err != nil {
				return err
			}
		} else {
			_, err := copyFile(path, destPath)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func CopyDir(srcDir string, destDir string) error {
	_, err := os.Stat(destDir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(destDir, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	return copyDir(srcDir, destDir)
}

func replaceContent(path string, domain, appName string) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	strContent := string(content)

	strContent = strings.ReplaceAll(strContent, "{{domain}}", domain)
	strContent = strings.ReplaceAll(strContent, "{{app_name}}", appName)
	strContent = strings.ReplaceAll(strContent, "{{app_name2}}", strings.ReplaceAll(appName, "-", "_"))
	strContent = strings.ReplaceAll(strContent, "{{APP_NAME}}", strings.ToUpper(strings.ReplaceAll(appName, "-", "_")))

	err = file.Close()
	if err != nil {
		return err
	}

	fileName := filepath.Base(path)
	if fileName == "template_config.ini" {
		err := os.Remove(path)
		if err != nil {
			return err
		}

		path = filepath.Join(filepath.Dir(path), fmt.Sprintf("%s_config.ini", strings.ReplaceAll(appName, "-", "_")))
	}

	file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(strContent)
	if err != nil {
		return err
	}

	return nil
}

const (
	exportScriptTemplate = `#!/bin/bash
export GOPROXY=https://rainbow:WyH8nqyiH8huRhnsQjHa@proxy.rpkg.cc,direct;
export GOSUMDB=off;
export CGO_ENABLED=0;
cd %s;
go mod init %s/%s;
go mod tidy;
`
)

func main() {
	domain := "dev.funplus.com"
	appName := "glc-backend"

	err := CopyDir("template", appName)
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(appName, func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo == nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		} else {
			if !strings.HasSuffix(path, ".go") && strings.HasSuffix(path, ".int") {
				return nil
			}

			err := replaceContent(path, domain, appName)
			if err != nil {
				return err
			}
		}

		return nil
	})

	exportScript := fmt.Sprintf(exportScriptTemplate, appName, domain, appName)

	cmd := exec.Command("/bin/sh", "-c", exportScript)

	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
}
