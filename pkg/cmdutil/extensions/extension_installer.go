/*
Copyright 2023 The Keyfactor Command Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package extensions

import (
	"archive/zip"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v3"
	"io"
	"kfutil/pkg/cmdutil"
	"kfutil/pkg/cmdutil/flags"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const extensionNameVersionDelimiter = "@"
const defaultExtensionOutDir = "./extensions"
const directoryPermissions = 0755
const filePermissions = 0644

type ExtensionInstaller struct {
	githubToken          string
	githubOrg            string
	extensionsToInstall  Extensions
	currentlyInstalled   Extensions
	interactiveMode      bool
	extensionDirName     string
	requiresConfirmation bool
	errs                 []error
	upgrade              bool
	prune                bool
	writer               io.Writer
}

func NewExtensionInstallerBuilder() *ExtensionInstaller {
	return &ExtensionInstaller{
		extensionsToInstall: make(Extensions),
		currentlyInstalled:  make(Extensions),
	}
}

func (b *ExtensionInstaller) InteractiveMode(interactiveMode bool) *ExtensionInstaller {
	b.interactiveMode = interactiveMode

	return b
}

func (b *ExtensionInstaller) Writer(writer io.Writer) *ExtensionInstaller {
	b.writer = writer

	return b
}

func (b *ExtensionInstaller) Extensions(extensionsString []string) *ExtensionInstaller {
	// Serialize extensionsString into a map of extensions to install and the version
	for _, extensionString := range extensionsString {
		extension, err := ParseExtensionString(extensionString)
		if err != nil {
			cmdutil.PrintError(err)
			b.errs = append(b.errs, err)
		}

		b.extensionsToInstall[extension.Name] = extension.Version
	}

	return b
}

func ParseExtensionString(extensionString string) (Extension, error) {
	extension := Extension{}

	// Split the extensionString on the delimiter, if it exists
	extensionParts := strings.Split(extensionString, extensionNameVersionDelimiter)
	if len(extensionParts) == 1 {
		extension.Name = ExtensionName(extensionParts[0])
		extension.Version = "latest"
	} else if len(extensionParts) == 2 {
		extension.Name = ExtensionName(extensionParts[0])
		extension.Version = Version(extensionParts[1])
	} else {
		return extension, fmt.Errorf("invalid extension string: %s", extensionString)
	}

	return extension, nil
}

func (b *ExtensionInstaller) ExtensionConfig(extensionConfigOptions flags.FilenameOptions) *ExtensionInstaller {
	if extensionConfigOptions.IsEmpty() {
		return b
	}

	if err := extensionConfigOptions.Validate(); err != nil {
		b.errs = append(b.errs, fmt.Errorf("failed to validate filename options: %s", err))
		return b
	}

	// Read the file into a UniversalOrchestratorHelmValues struct
	extensionConfigFileBytes, err := extensionConfigOptions.Read()
	if err != nil {
		cmdutil.PrintError(err)
		b.errs = append(b.errs, err)
	}

	// extensionConfigFileBytes contains a map of extension names to versions
	// Serialize the bytes into a map of extensions to install and the version
	extensionsToInstallMap := make(map[string]string)
	err = yaml.Unmarshal(extensionConfigFileBytes, &extensionsToInstallMap)
	if err != nil {
		b.errs = append(b.errs, fmt.Errorf("error unmarshalling extension config file: %s", err))
	}

	// Convert the map of extensions to install and the version into a map of extensions to install and a slice of versions
	for extensionName, extensionVersion := range extensionsToInstallMap {
		b.extensionsToInstall[ExtensionName(extensionName)] = Version(extensionVersion)
	}

	return b
}

func (b *ExtensionInstaller) ExtensionDir(dir string) *ExtensionInstaller {
	if dir == "" {
		dir = defaultExtensionOutDir
	}

	// Don't make any changes to the extension directory unless user confirms actions
	// This happens in the Run() method

	b.extensionDirName = dir

	return b
}

func (b *ExtensionInstaller) Prune() *ExtensionInstaller {
	b.prune = true

	return b
}

func (b *ExtensionInstaller) cacheExtensionsDir() error {
	// Return if the extension directory doesn't exist or is empty
	if _, err := os.Stat(b.extensionDirName); os.IsNotExist(err) {
		return nil
	}

	// Read entries in the extensions directory
	extensionEntries, err := os.ReadDir(b.extensionDirName)
	if err != nil {
		return err
	}

	// Create a slice of the entries in the extensions directory
	var currentlyInstalledExtensions []string
	for _, entry := range extensionEntries {
		if entry.IsDir() {
			currentlyInstalledExtensions = append(currentlyInstalledExtensions, entry.Name())
		}
	}

	if b.currentlyInstalled == nil {
		b.currentlyInstalled = make(Extensions)
	}

	// If the name in each entry of the slice conforms to the format <extension name>_<version>, add it to the extensionsToInstall map
	for _, extension := range currentlyInstalledExtensions {
		extensionParts := strings.Split(extension, "_")
		if len(extensionParts) == 2 {
			extensionName := extensionParts[0]
			extensionVersion := extensionParts[1]

			b.currentlyInstalled[ExtensionName(extensionName)] = Version(extensionVersion)
		}
	}

	return nil
}

func (b *ExtensionInstaller) Confirm() bool {
	// Render the list of extensions to install
	confirmPromptString := "Installing extensions could overwrite existing extensions and configurations. The following changes will be made:\n"
	for extensionName, extensionVersion := range b.extensionsToInstall {
		var versionString string
		if currentVersion, ok := b.currentlyInstalled[extensionName]; ok && currentVersion != extensionVersion {
			versionString = fmt.Sprintf("%s -> %s", currentVersion, extensionVersion)
		} else if ok && currentVersion == extensionVersion {
			confirmPromptString += fmt.Sprintf("%s @ %s is already installed.\n", extensionName, extensionVersion)
			continue
		} else {
			versionString = string(extensionVersion)
		}

		confirmPromptString += fmt.Sprintf("Install %s: %s\n", extensionName, versionString)
	}

	if b.prune {
		for extensionName, extensionVersion := range b.currentlyInstalled {
			if _, ok := b.extensionsToInstall[extensionName]; !ok {
				confirmPromptString += fmt.Sprintf("Remove %s: %s\n", extensionName, extensionVersion)
			}
		}
	}

	b.AddRuntimeLog(confirmPromptString)

	// Confirm that the user wants to install the extensions
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: "Install extensions?",
		Help:    confirmPromptString,
		Default: false,
	}
	err := survey.AskOne(confirmPrompt, &confirm)
	if err != nil {
		b.errs = append(b.errs, err)
	}

	return confirm
}

func (b *ExtensionInstaller) AutoConfirm(autoConfirm bool) *ExtensionInstaller {
	b.requiresConfirmation = !autoConfirm

	return b
}

func (b *ExtensionInstaller) Token(token string) *ExtensionInstaller {
	b.githubToken = token

	return b
}

func (b *ExtensionInstaller) Org(org string) *ExtensionInstaller {
	b.githubOrg = org

	return b
}

func (b *ExtensionInstaller) Upgrade() *ExtensionInstaller {
	b.upgrade = true

	return b
}

func (b *ExtensionInstaller) ValidateExtensionsToInstall() error {
	for extensionName, extensionVersion := range b.extensionsToInstall {
		// Ensure that at least one version was provided
		if extensionVersion == "" {
			return fmt.Errorf("no version provided for extension %s", extensionName)
		}

		// If the version is "latest", get the latest version from Github
		if extensionVersion == "latest" {
			// Get the list of versions from Github
			versions, err := NewGithubReleaseFetcher(b.githubOrg, b.githubToken).GetExtensionVersions(extensionName)
			if err != nil {
				return fmt.Errorf("failed to get list of versions for extension %s. Does it exist?: %s", extensionName, err)
			}

			// Set the version to the latest version
			b.extensionsToInstall[extensionName] = versions[0]

			// We know that the version is valid, so we can skip the rest of the checks
			continue
		}

		// Validate that the extension exists
		if exists, err := NewGithubReleaseFetcher(b.githubOrg, b.githubToken).ExtensionExists(extensionName, extensionVersion); err != nil {
			return fmt.Errorf("failed to determine if extension exists %s:%s: %s", extensionName, extensionVersion, err)
		} else if !exists {
			return fmt.Errorf("extension %s:%s does not exist", extensionName, extensionVersion)
		}
	}

	return nil
}

func (b *ExtensionInstaller) PreFlight() error {
	// Print any errors and exit if there are any
	for _, err := range b.errs {
		cmdutil.PrintError(err)
	}

	if len(b.errs) > 0 {
		// Return the first error; the rest have already been printed
		return b.errs[0]
	}

	if err := b.ValidateExtensionsToInstall(); err != nil {
		return fmt.Errorf("failed to validate extensions: %s", err)
	}

	// Set default extension dir name if it is empty
	if b.extensionDirName == "" {
		b.extensionDirName = defaultExtensionOutDir
	}

	return nil
}

func (b *ExtensionInstaller) applyUpdatesToExtensionsToInstall() error {
	for extensionName := range b.currentlyInstalled {
		// Get the latest version of the extension
		latestExtensionVersions, err := NewGithubReleaseFetcher(b.githubOrg, b.githubToken).GetExtensionVersions(extensionName)
		if err != nil {
			return fmt.Errorf("failed to get latest version for extension %s: %s", extensionName, err)
		}

		if len(latestExtensionVersions) == 0 {
			return fmt.Errorf("failed to get latest version for %q: no versions found", extensionName)
		}

		// Set or overwrite the version in the extensionsToInstall map
		b.extensionsToInstall[extensionName] = latestExtensionVersions[0]
	}

	return nil
}

func (b *ExtensionInstaller) cleanExtensionDirectory() error {
	// Return if the extension directory doesn't exist or is empty
	if _, err := os.Stat(b.extensionDirName); os.IsNotExist(err) {
		return nil
	}

	for extensionName, extensionVersion := range b.extensionsToInstall {
		// If the extension is currently installed and the version is different than the version to install, delete the
		// directory containing the currently installed extension
		if currentlyInstalledVersion, ok := b.currentlyInstalled[extensionName]; ok && currentlyInstalledVersion != extensionVersion {
			extensionDir := filepath.Join(b.extensionDirName, fmt.Sprintf("%s_%s", extensionName, currentlyInstalledVersion))
			b.AddRuntimeLog("Removing %s_%s", extensionName, currentlyInstalledVersion)

			err := os.RemoveAll(extensionDir)
			if err != nil {
				return fmt.Errorf("failed to remove extension directory %s: %s", extensionDir, err)
			}
		}
	}

	// If the prune flag was set, remove any extensions that are currently installed but not in the extensionsToInstall map
	if b.prune {
		// Read entries in the extensions directory
		extensionEntries, err := os.ReadDir(b.extensionDirName)
		if err != nil {
			return fmt.Errorf("failed to read entries in extension directory: %s", err)
		}

		for _, entry := range extensionEntries {
			deletionRequired := false

			if !entry.IsDir() {
				deletionRequired = true
			}

			entryComponents := strings.Split(entry.Name(), "_")
			if len(entryComponents) != 2 {
				deletionRequired = true
			}

			if len(entryComponents) == 2 {
				extensionName := ExtensionName(entryComponents[0])
				extensionVersion := Version(entryComponents[1])

				// Determine if the extension is in the extensionsToInstall map
				if version, ok := b.extensionsToInstall[extensionName]; ok {
					if extensionVersion != version {
						deletionRequired = true
					}
				} else {
					deletionRequired = true
				}
			}

			if deletionRequired {
				err = os.RemoveAll(filepath.Join(b.extensionDirName, entry.Name()))
				if err != nil {
					return fmt.Errorf("failed to remove extension directory %s: %s", entry.Name(), err)
				}

				b.AddRuntimeLog("Removed %q", entry.Name())
			}
		}
	}

	return nil
}

func (b *ExtensionInstaller) AddRuntimeLog(format string, a ...any) {
	if b.writer != nil {
		_, err := fmt.Fprintf(b.writer, format+"\n", a...)
		if err != nil {
			log.Println(err)
		}
	}
}

func (b *ExtensionInstaller) Run() error {
	// Create the extensions directory if it doesn't exist
	if _, err := os.Stat(b.extensionDirName); os.IsNotExist(err) {
		err = os.MkdirAll(b.extensionDirName, directoryPermissions)
		if err != nil {
			return fmt.Errorf("failed to create extensions directory: %s", err)
		}
	}

	// Cache the extensions in the extension directory to the currentlyInstalled map.
	// Useful for determining which extensions to uninstall if the user is upgrading
	err := b.cacheExtensionsDir()
	if err != nil {
		return fmt.Errorf("failed to cache extensions in extension directory: %s", err)
	}

	// If interactive mode is enabled, prompt the user to confirm the extensions to install
	if b.interactiveMode && !b.upgrade {
		err = b.PromptForExtensions()
		if err != nil {
			return fmt.Errorf("failed to prompt for extensions: %s", err)
		}
	}

	// If the update flag is set, remove any extensions that are currently installed but not in the extensionsToInstall map
	if b.upgrade {
		err = b.applyUpdatesToExtensionsToInstall()
		if err != nil {
			return fmt.Errorf("failed to apply upgrades to currently installed extensions: %s", err)
		}
	}

	// Confirm that the user wants to install the extensions or otherwise make changes to the extension directory
	if b.requiresConfirmation && !b.Confirm() {
		log.Println("Action cancelled by user")
		b.AddRuntimeLog("Action cancelled by user")
		return nil
	}

	err = b.cleanExtensionDirectory()
	if err != nil {
		return fmt.Errorf("failed to clean extension directory: %s", err)
	}

	for name, version := range b.extensionsToInstall {
		if _, ok := b.currentlyInstalled[name]; ok && version == b.currentlyInstalled[name] {
			log.Printf("extension %s:%s is already installed - do nothing.", name, version)
			b.AddRuntimeLog("Extension %s:%s is already installed", name, version)
			continue
		}

		zipFilePath := filepath.Join(b.extensionDirName, fmt.Sprintf("%s_%s.zip", name, version))

		b.AddRuntimeLog("Downloading %s", zipFilePath)

		extensionZipBytes, err := NewGithubReleaseFetcher(b.githubOrg, b.githubToken).DownloadExtension(name, version)
		if err != nil {
			return fmt.Errorf("failed to download extension %s:%s (release must contain contain \"%s_%s.zip\"): %s", name, version, name, version, err)
		}

		// Write the extension zip file to disk
		err = os.WriteFile(zipFilePath, *extensionZipBytes, filePermissions)
		if err != nil {
			return fmt.Errorf("failed to write extension zip file to disk: %s", err)
		}

		destDir := filepath.Join(b.extensionDirName, fmt.Sprintf("%s_%s", name, version))

		// Unzip the extension
		err = b.unzip(zipFilePath, destDir)
		if err != nil {
			return fmt.Errorf("failed to unzip extension: %s", err)
		}

		// Delete the zip file
		err = os.Remove(zipFilePath)
		if err != nil {
			return fmt.Errorf("failed to delete extension zip file: %s", err)
		}
	}

	return nil
}

func (b *ExtensionInstaller) unzip(zipFilePath, destinationDirectory string) error {
	b.AddRuntimeLog("Unzipping %s to %s", zipFilePath, destinationDirectory)

	archive, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file at %q: %s", zipFilePath, err)
	}
	defer func(r *zip.ReadCloser) {
		err = r.Close()
		if err != nil {
			cmdutil.PrintError(err)
		}
	}(archive)

	for _, f := range archive.File {
		filePath := filepath.Join(destinationDirectory, f.Name)
		log.Println("unzipping file", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(destinationDirectory)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", filePath)
		}
		if f.FileInfo().IsDir() {
			log.Println("creating directory...")
			err = os.MkdirAll(filePath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		if strings.Contains(filePath, "\\") && runtime.GOOS != "windows" {
			return fmt.Errorf("illegal file path: %s", filePath)
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %s", filepath.Dir(filePath), err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to open file %s: %s", filePath, err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %s", filePath, err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return fmt.Errorf("failed to copy file %s: %s", filePath, err)
		}

		err = dstFile.Close()
		if err != nil {
			return fmt.Errorf("failed to close file %s: %s", filePath, err)
		}
		err = fileInArchive.Close()
		if err != nil {
			return fmt.Errorf("failed to close file %s: %s", filePath, err)
		}
	}

	return nil
}

func (b *ExtensionInstaller) PromptForExtensions() error {
	extensionList, err := NewGithubReleaseFetcher(b.githubOrg, b.githubToken).GetExtensionList()
	if err != nil {
		return fmt.Errorf("failed to get list of extensions: %s", err)
	}

	extensionNameList := make([]string, 0)
	for extensionName := range extensionList {
		extensionNameList = append(extensionNameList, string(extensionName))
	}

	descriptionFunc := func(value string, index int) string {
		for extensionName, version := range extensionList {
			if string(extensionName) == value {
				description := fmt.Sprintf("%s", version)

				if installed, ok := b.currentlyInstalled[extensionName]; ok {
					description = fmt.Sprintf("%s (currently %s)", description, installed)
				}

				return description
			}
		}
		return ""
	}

	var extensionsToInstall []string
	prompt := &survey.MultiSelect{
		Message:     "Select the extensions to install - the most recent versions are displayed",
		Options:     alphabetize(extensionNameList),
		Description: descriptionFunc,
	}
	err = survey.AskOne(prompt, &extensionsToInstall)
	if err != nil {
		return fmt.Errorf("failed to prompt for extensions: %s", err)
	}

	for _, extension := range extensionsToInstall {
		// extensionsToInstall is a subset of extensionList since user was prompted to select from extensionList
		// So we can safely assume that extension is in extensionList

		// Add the extension to the list of extensions to install
		b.extensionsToInstall[ExtensionName(extension)] = extensionList[ExtensionName(extension)]
	}

	return nil
}

func (b *ExtensionInstaller) SetExtensionsToInstall(extensionsToInstall Extensions) {
	b.extensionsToInstall = extensionsToInstall
}

func (b *ExtensionInstaller) GetExtensionsToInstall() Extensions {
	return b.extensionsToInstall
}

func alphabetize(list []string) []string {
	// Make a copy of the original list
	sortedList := make([]string, len(list))
	copy(sortedList, list)

	// Sort the copied list alphabetically
	sort.Strings(sortedList)

	return sortedList
}
