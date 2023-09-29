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
	"sort"
	"strings"
)

const extensionNameVersionDelimiter = "@"
const defaultExtensionOutDir = "./extensions"
const directoryPermissions = 0755
const filePermissions = 0644

type ExtensionInstaller struct {
	githubToken          string
	extensionsToInstall  Extensions
	interactiveMode      bool
	extensionDirName     string
	requiresConfirmation bool
	errs                 []error
}

func NewExtensionInstallerBuilder() *ExtensionInstaller {
	return &ExtensionInstaller{}
}

func (b *ExtensionInstaller) InteractiveMode(interactiveMode bool) *ExtensionInstaller {
	b.interactiveMode = interactiveMode

	return b
}

func (b *ExtensionInstaller) Extensions(extensionsString []string) *ExtensionInstaller {
	if b.extensionsToInstall == nil {
		b.extensionsToInstall = make(Extensions)
	}

	// Serialize extensionsString into a map of extensions to install and the version
	for _, extensionString := range extensionsString {
		extension, err := ParseExtensionString(extensionString)
		if err != nil {
			cmdutil.PrintError(err)
			b.errs = append(b.errs, err)
		}

		b.extensionsToInstall[extension.Name] = append(make([]Version, 0), extension.Version)
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

	if b.extensionsToInstall == nil {
		b.extensionsToInstall = make(Extensions)
	}

	// Convert the map of extensions to install and the version into a map of extensions to install and a slice of versions
	for extensionName, extensionVersion := range extensionsToInstallMap {
		b.extensionsToInstall[ExtensionName(extensionName)] = append(make([]Version, 0), Version(extensionVersion))
	}

	return b
}

func (b *ExtensionInstaller) ExtensionDir(dir string) *ExtensionInstaller {
	if dir == "" {
		dir = defaultExtensionOutDir
	}

	// If the directory doesn't exist, create it
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, directoryPermissions)
		if err != nil {
			b.errs = append(b.errs, err)
		}
	}

	b.extensionDirName = dir

	return b
}

func (b *ExtensionInstaller) Confirm() bool {
	// Render the list of extensions to install
	extensionListString := ""
	for extensionName, extensionVersions := range b.extensionsToInstall {
		extensionListString += fmt.Sprintf("%s: %s\n", extensionName, extensionVersions)
	}

	// Confirm that the user wants to install the extensions
	confirm := false
	confirmPrompt := &survey.Confirm{
		Message: "Install extensions?",
		Help:    extensionListString,
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

func (b *ExtensionInstaller) PreFlight() error {
	for extensionName, extensionVersions := range b.extensionsToInstall {
		// Ensure that at least one version was provided
		if len(extensionVersions) == 0 {
			return fmt.Errorf("no version provided for extension %s", extensionName)
		}

		// If the version is "latest", get the latest version from Github
		if len(extensionVersions) == 1 && extensionVersions[0] == "latest" {
			// Get the list of versions from Github
			versions, err := NewGithubReleaseFetcher(b.githubToken).GetExtensionVersions(extensionName)
			if err != nil {
				b.errs = append(b.errs, err)
			}

			// Set the version to the latest version
			b.extensionsToInstall[extensionName] = versions

			// We know that the version is valid, so we can skip the rest of the checks
			continue
		}

		// Validate that the extension exists
		if exists, err := NewGithubReleaseFetcher(b.githubToken).ExtensionExists(extensionName, extensionVersions[0]); err != nil {
			return fmt.Errorf("failed to determine if extension exists %s:%s: %s", extensionName, extensionVersions[0], err)
		} else if !exists {
			return fmt.Errorf("extension %s:%s does not exist", extensionName, extensionVersions[0])
		}
	}

	// Print any errors and exit if there are any
	for _, err := range b.errs {
		cmdutil.PrintError(err)
	}

	if len(b.errs) > 0 {
		// Return the first error; the rest have already been printed
		return b.errs[0]
	}

	// Set default extension dir name if it is empty
	if b.extensionDirName == "" {
		b.extensionDirName = defaultExtensionOutDir
	}

	// Create the extension directory if it doesn't exist
	if _, err := os.Stat(b.extensionDirName); os.IsNotExist(err) {
		return fmt.Errorf("extension directory %s does not exist", b.extensionDirName)
	}

	return nil
}

func (b *ExtensionInstaller) Run() error {
	// If interactive mode is enabled, prompt the user to confirm the extensions to install
	if b.interactiveMode {
		err := b.PromptForExtensions()
		if err != nil {
			return fmt.Errorf("failed to prompt for extensions: %s", err)
		}
	}

	// Confirm that the user wants to install the extensions
	if b.requiresConfirmation && !b.Confirm() {
		log.Println("Action cancelled by user")
		return nil
	}

	for name, version := range b.extensionsToInstall {
		zipFilePath := filepath.Join(b.extensionDirName, fmt.Sprintf("%s_%s.zip", name, version[0]))

		extensionZipBytes, err := NewGithubReleaseFetcher(b.githubToken).DownloadExtension(name, version[0])
		if err != nil {
			return fmt.Errorf("failed to download extension %s:%s: %s", name, version[0], err)
		}

		// Write the extension zip file to disk
		err = os.WriteFile(zipFilePath, *extensionZipBytes, filePermissions)
		if err != nil {
			return fmt.Errorf("failed to write extension zip file to disk: %s", err)
		}

		destDir := filepath.Join(b.extensionDirName, fmt.Sprintf("%s_%s", name, version[0]))

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

func (b *ExtensionInstaller) unzip(zipFilePath string, destinationDirectory string) error {
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
	if b.extensionsToInstall == nil {
		b.extensionsToInstall = make(Extensions)
	}

	extensionList, err := NewGithubReleaseFetcher(b.githubToken).GetExtensionList()
	if err != nil {
		return fmt.Errorf("failed to get list of extensions: %s", err)
	}

	extensionNameList := make([]string, 0)
	for extensionName, _ := range extensionList {
		extensionNameList = append(extensionNameList, string(extensionName))
	}

	descriptionFunc := func(value string, index int) string {
		for extensionName, versions := range extensionList {
			if string(extensionName) == value {
				description := fmt.Sprintf("%s", versions[0])

				if installed, ok := b.extensionsToInstall[extensionName]; ok {
					description = fmt.Sprintf("%s (currently %s)", description, installed[0])
				}

				return description
			}
		}
		return ""
	}

	// TODO make selection process declarative - i.e. extensionList not in the extensionsToInstall slice are not installed
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
		versions := extensionList[ExtensionName(extension)]
		// Also, extensionList is only populated if there are releases for the extension
		latestVersion := versions[0]

		// Add the extension to the list of extensions to install
		b.extensionsToInstall[ExtensionName(extension)] = append(make([]Version, 0), latestVersion)
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
