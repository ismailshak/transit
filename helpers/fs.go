package helpers

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ismailshak/transit/logger"
)

const (
	// File mode for read+write+execute
	RWX = 0777
	// File mode for read+write
	RW_ = 0776
)

func GetConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error(fmt.Sprint(err))
		Exit(1)
	}

	return filepath.Join(homeDir, ".config", "transit")
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
	return os.WriteFile(filePath, content, RW_)
}

// If directory already exists, nothing will happen
func CreateDir(dirPath string) error {
	return os.MkdirAll(dirPath, RW_)
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

func Unzip(rc *zip.ReadCloser, dest string) error {
	for _, f := range rc.File {
		err := unzipFile(f, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzipFile(f *zip.File, dest string) error {
	// Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(dest, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// Create directories if needed
	if f.FileInfo().IsDir() {
		if err := CreateDir(filePath); err != nil {
			return fmt.Errorf("failed to create subdirectory: %s", err)
		}

		return nil
	}

	// Create a destination file for unzipped content
	destFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %s", err)
	}

	defer destFile.Close()

	// Unzip the content of a file and copy it to the destination file
	zippedFile, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zipped file: %s", err)
	}

	defer zippedFile.Close()

	if _, err := io.Copy(destFile, zippedFile); err != nil {
		return fmt.Errorf("failed to copy zipped file content: %s", err)
	}

	return nil
}
