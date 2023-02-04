package helpers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ismailshak/transit/helpers"
)

func TestCreateIfNotFound_NotFound(t *testing.T) {
	testFilePath := t.TempDir() + "./config/create-config.yml"

	err := helpers.CreatePathIfNotFound(testFilePath)
	if err != nil {
		t.Logf("Failed to create path: %s", err)
		t.FailNow()
	}

	if !helpers.FileExists(testFilePath) {
		t.Logf("file '%s' was not created", testFilePath)
		t.FailNow()
	}

	testContent := "Hello"
	err = helpers.WriteFile(testFilePath, []byte(testContent))
	if err != nil {
		t.Logf("File created could not be written to: %s", err)
		t.FailNow()
	}

	content, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Logf("File created could not be read: %s", err)
		t.FailNow()
	}

	if string(content) != testContent {
		t.Logf("File created did not have the correct content. Expected '%s' but found '%s'", testContent, string(content))
		t.FailNow()
	}
}

func TestCreateIfNotFound_Found(t *testing.T) {
	testFilePath := t.TempDir() + "./config/exists-config.yml"
	os.MkdirAll(filepath.Dir(testFilePath), 0777)

	testContent := "Already exists"
	err := helpers.WriteFile(testFilePath, []byte(testContent))
	if err != nil {
		t.Logf("Failed to create test file: %s", err)
		t.FailNow()
	}

	err = helpers.CreatePathIfNotFound(testFilePath)
	if err != nil {
		t.Logf("Failed to create path: %s", err)
		t.FailNow()
	}

	content, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Logf("File created could not be read: %s", err)
		t.FailNow()
	}

	if string(content) != testContent {
		t.Logf("File created did not have the correct content. Expected '%s' but found '%s'", testContent, string(content))
		t.FailNow()
	}
}
