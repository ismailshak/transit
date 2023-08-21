package helpers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ismailshak/transit/logger"
)

func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error(fmt.Sprint(err))
		Exit(1)
	}

	return homeDir + "/.config/transit"
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Debug(fmt.Sprint(err))
		}

		return false
	}

	return true
}

func DirExists(dirPath string) bool {
	info, err := os.Stat(dirPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Debug(fmt.Sprint(err))
		}

		return false
	}

	if !info.IsDir() {
		logger.Debug(fmt.Sprintf("Expected a directory but found a file: %s", dirPath))
		return false
	}

	return true
}

func WriteFile(filePath string, content []byte) error {
	return os.WriteFile(filePath, content, 0777)
}

func CreateDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0777)
}

func CreatePathIfNotFound(configPath string) error {
	if FileExists(configPath) {
		return nil
	}

	dirPath := filepath.Dir(configPath)
	if !DirExists(dirPath) {
		err := CreateDir(dirPath)
		if err != nil {
			return err
		}
	}

	return WriteFile(configPath, []byte{})
}
