package extensions

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestGooseChase(t *testing.T) {
	// Download all latest orchestrator/PAM extensions to a directory, then unzip them
	// If any of the extensions fail to download or unzip, fail the test

	extensionDirName := "test-extensions"
	if _, err := os.Stat(extensionDirName); os.IsNotExist(err) {
		err = os.MkdirAll(extensionDirName, directoryPermissions)
		if err != nil {
			t.Fatalf("failed to create extensions directory: %s", err)
		}
	}

	// Get a list of all latest orchestrator/PAM extensions
	extensions, err := NewGithubReleaseFetcher("", GetGithubToken()).GetExtensionList()
	if err != nil {
		t.Fatal(err)
	}

	var failedList []string

	// For each extension, download
	for extension, version := range extensions {
		zipFilePath := filepath.Join(extensionDirName, fmt.Sprintf("%s_%s.zip", extension, version))

		extensionZipBytes, err := NewGithubReleaseFetcher("", GetGithubToken()).DownloadExtension(extension, version)
		if err != nil {
			t.Fatalf("failed to download extension %s:%s (release must contain contain \"%s_%s.zip\"): %s", extension, version, extension, version, err)
		}

		// Write the extension zip file to disk
		err = os.WriteFile(zipFilePath, *extensionZipBytes, filePermissions)
		if err != nil {
			t.Fatalf("failed to write extension zip file to disk: %s", err)
		}

		destDir := filepath.Join(extensionDirName, fmt.Sprintf("%s_%s", extension, version))

		// Unzip the extension
		err = NewExtensionInstallerBuilder().unzip(zipFilePath, destDir)
		if err == nil {
			// Delete the zip file if the extension was successfully unzipped
			err = os.Remove(zipFilePath)
			if err != nil {
				t.Errorf("failed to delete extension zip file: %s", err)
			}
			continue
		}

		log.Printf("failed to unzip extension: %s", err)
		failedList = append(failedList, fmt.Sprintf("%s:%s", extension, version))
	}

	if len(failedList) > 0 {
		t.Errorf("failed to download the following extensions: %s", failedList)
	}
}

func TestGooseChaseWithLocalZip(t *testing.T) {
	zipFilePath := "test-extensions/iis-orchestrator_2.2.2.zip"
	extensionDirName := "test-extensions/iis-orchestrator_2.2.2"

	// Unzip the extension
	err := NewExtensionInstallerBuilder().unzip(zipFilePath, extensionDirName)
	if err != nil {
		t.Fatalf("failed to unzip extension: %s", err)
	}
}
