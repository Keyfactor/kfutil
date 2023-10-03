package extensions

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
	"kfutil/pkg/cmdtest"
	"kfutil/pkg/cmdutil/flags"
	"os"
	"reflect"
	"testing"
)

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}

func directoryTestHelper(dirName string) (func() error, error) {
	// Delete the test directory if it exists
	if _, err := os.Stat(dirName); !os.IsNotExist(err) {
		err = os.RemoveAll(dirName)
		if err != nil {
			return nil, err
		}
	}

	// Create the test directory
	err := os.Mkdir(dirName, 0755)
	if err != nil {
		return nil, err
	}

	// Return a function to delete the test directory
	return func() error {
		// Delete the test directory
		err = os.RemoveAll(dirName)
		if err != nil {
			return err
		}

		return nil
	}, nil
}

func TestParseExtensionString(t *testing.T) {
	t.Run("ValidExtensionString", func(t *testing.T) {
		extensionString := "test-extension@1.0.0"
		extension, err := ParseExtensionString(extensionString)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if extension.Name != "test-extension" {
			t.Errorf("Expected extension name to be %s, got %s", "test-extension", extension.Name)
		}

		if extension.Version != "1.0.0" {
			t.Errorf("Expected extension version to be %s, got %s", "1.0.0", extension.Version)
		}
	})

	t.Run("ValidExtensionStringNoVersion", func(t *testing.T) {
		extensionString := "test-extension"
		extension, err := ParseExtensionString(extensionString)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if extension.Name != "test-extension" {
			t.Errorf("Expected extension name to be %s, got %s", "test-extension", extension.Name)
		}

		if extension.Version != "latest" {
			t.Errorf("Expected extension version to be %s, got %s", "latest", extension.Version)
		}
	})

	t.Run("InvalidExtensionString", func(t *testing.T) {
		extensionString := "test-extension@1.0.0@1.0.0"
		_, err := ParseExtensionString(extensionString)
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})
}

func TestExtensionInstaller_InteractiveMode(t *testing.T) {
	builder := NewExtensionInstallerBuilder().InteractiveMode(true)

	if !builder.interactiveMode {
		t.Error("Expected interactiveMode to be true")
	}
}

func TestExtensionInstaller_Extensions(t *testing.T) {
	t.Run("ValidExtensions", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().Extensions([]string{"test-extension@1.0.0"})

		extensions := make(Extensions)
		extensions["test-extension"] = "1.0.0"

		if !reflect.DeepEqual(builder.extensionsToInstall, extensions) {
			t.Errorf("Expected extensionsToInstall to be %v, got %v", extensions, builder.extensionsToInstall)
		}
	})

	t.Run("InvalidExtensions", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().Extensions([]string{"test-extension@1.0.0@1.0.0"})

		if len(builder.errs) == 0 {
			t.Error("Expected error, got none")
		}
	})

	t.Run("NilExtensions", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().Extensions(nil)

		if len(builder.extensionsToInstall) != 0 {
			t.Errorf("Expected extensionsToInstall to be nil, got %v", builder.extensionsToInstall)
		}
	})
}

func TestExtensionInstaller_ExtensionConfig(t *testing.T) {
	const testFilename = "test_extensions.yaml"

	// Create the filename options
	options := flags.FilenameOptions{}

	t.Run("EmptyFilenameOptions", func(t *testing.T) {
		// Test with no files in the options
		builder := NewExtensionInstallerBuilder().ExtensionConfig(options)

		// Verify that extensionsToInstall was set correctly
		if len(builder.extensionsToInstall) != 0 {
			t.Errorf("Expected extensionsToInstall to have length 0, got %d", len(builder.extensionsToInstall))
		}
	})

	t.Run("ValidFilenameOptions", func(t *testing.T) {
		// Create a test file with some extensions
		file, err := os.Create(testFilename)
		if err != nil {
			t.Error(err)
		}

		// Write to the test file
		testBytes := []byte("test-extension: 1.0.0\ntest-extension1: latest\n")
		_, err = file.Write(testBytes)
		if err != nil {
			t.Error(err)
		}

		// Add the test file to the options
		options.Filenames = []string{testFilename}

		// Create the builder and set the extension config
		builder := NewExtensionInstallerBuilder().ExtensionConfig(options)

		// Verify that extensionsToInstall was set correctly
		if len(builder.extensionsToInstall) != 2 {
			t.Errorf("Expected extensionsToInstall to have length 2, got %d", len(builder.extensionsToInstall))
		}

		// Verify that the extensions were set correctly
		if version, ok := builder.extensionsToInstall["test-extension"]; ok {
			if version != "1.0.0" {
				t.Errorf("Expected version to be %s, got %s", "1.0.0", version)
			}
		} else {
			t.Errorf("Expected test-extension to be in extensionsToInstall")
		}

		if version, ok := builder.extensionsToInstall["test-extension1"]; ok {
			if version != "latest" {
				t.Errorf("Expected version to be %s, got %s", "latest", version)
			}
		} else {
			t.Errorf("Expected test-extension1 to be in extensionsToInstall")
		}

		// Delete the test file
		err = os.Remove(testFilename)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestExtensionInstaller_ExtensionDir(t *testing.T) {
	builder := NewExtensionInstallerBuilder().ExtensionDir("testExtDir")

	if builder.extensionDirName != "testExtDir" {
		t.Errorf("Expected extensionDir to be %s, got %s", "test", builder.extensionDirName)
	}
}

func TestExtensionInstaller_cacheExtensionsDir(t *testing.T) {
	// Prepare the test directory
	cleanupDir, err := directoryTestHelper("testExtDir")

	// Create two directories in the test directory with test extension names
	extensions := []Extension{
		{
			Name:    "test-extension",
			Version: "1.0.0",
		},
		{
			Name:    "test-extension1",
			Version: "2.0.0",
		},
	}
	for _, name := range extensions {
		err = os.Mkdir("testExtDir/"+fmt.Sprintf("%s_%s", name.Name, name.Version), 0755)
		if err != nil {
			t.Error(err)
		}
	}

	// Create the builder and set the extension dir
	builder := NewExtensionInstallerBuilder()
	builder.extensionDirName = "testExtDir"

	// Call cacheExtensionsDir
	err = builder.cacheExtensionsDir()
	if err != nil {
		t.Error(err)
	}

	// Validate that the extensions were cached correctly
	if len(builder.currentlyInstalled) != 2 {
		t.Errorf("Expected cachedExtensions to have length 2, got %d", len(builder.extensionsToInstall))
	}
	for _, extension := range extensions {
		if _, ok := builder.currentlyInstalled[extension.Name]; !ok {
			t.Errorf("Expected %s to be in cachedExtensions", extension.Name)
		} else if builder.currentlyInstalled[extension.Name] != extension.Version {
			t.Errorf("Expected %s to be %s, got %s", extension.Name, extension.Version, builder.extensionsToInstall[extension.Name])
		}
	}

	// Delete the test directory
	err = cleanupDir()
	if err != nil {
		t.Error(err)
	}
}

func TestExtensionInstaller_applyUpgradesToExtensionsToInstall(t *testing.T) {
	builder := NewExtensionInstallerBuilder()
	builder.currentlyInstalled = make(Extensions)

	// Get an extension
	extension, err := NewGithubReleaseFetcher("", GetGithubToken()).GetFirstExtension()
	if err != nil {
		t.Error(err)
	}

	// Add the extension to the currentlyInstalled map with a highly unlikely version
	builder.currentlyInstalled[extension] = "0.1.483"

	err = builder.applyUpdatesToExtensionsToInstall()
	if err != nil {
		t.Error(err)
	}

	// If the extension is in the map, it should have been upgraded
	if builder.extensionsToInstall[extension] == "0.1.483" {
		t.Errorf("Expected extension to be upgraded, got %s", builder.extensionsToInstall[extension])
	}
}

func TestExtensionInstaller_cleanExtensionDirectory(t *testing.T) {
	t.Run("UpgradableExtensions", func(t *testing.T) {
		cleanupDir, err := directoryTestHelper("testExtDir")
		if err != nil {
			t.Error(err)
		}

		// Create two directories in the test directory with test extension names
		extensions := []Extension{
			{
				Name:    "test-extension",
				Version: "1.0.0",
			},
			{
				Name:    "test-extension1",
				Version: "2.0.0",
			},
		}
		for _, name := range extensions {
			err = os.Mkdir("testExtDir/"+fmt.Sprintf("%s_%s", name.Name, name.Version), 0755)
			if err != nil {
				t.Error(err)
			}
		}

		// Create the builder and set the extension dir
		builder := NewExtensionInstallerBuilder().ExtensionDir("testExtDir")

		// Specify more recent versions of the extensions
		builder.extensionsToInstall = make(Extensions)
		builder.extensionsToInstall["test-extension"] = "2.0.0"
		builder.extensionsToInstall["test-extension1"] = "3.0.0"

		// Cache the extensions in the builder object
		err = builder.cacheExtensionsDir()
		if err != nil {
			t.Error(err)
		}

		// Call cleanExtensionDirectory
		err = builder.cleanExtensionDirectory()
		if err != nil {
			t.Error(err)
		}

		// Validate that the extensions were cleaned correctly
		// The extensions should have been deleted
		for _, extension := range extensions {
			if _, err = os.Stat("testExtDir/" + fmt.Sprintf("%s_%s", extension.Name, extension.Version)); !os.IsNotExist(err) {
				t.Errorf("Expected %s to be deleted", extension.Name)
			}
		}

		// Delete the test directory
		err = cleanupDir()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("NonSpecifiedExtensions", func(t *testing.T) {
		cleanupDir, err := directoryTestHelper("testExtDir")
		if err != nil {
			t.Error(err)
		}

		// Create two directories in the test directory with test extension names
		extensions := []Extension{
			{
				Name:    "test-extension",
				Version: "3.2.1",
			},
			{
				Name:    "test-extension1",
				Version: "1.2.3",
			},
		}
		for _, name := range extensions {
			err = os.Mkdir("testExtDir/"+fmt.Sprintf("%s_%s", name.Name, name.Version), 0755)
			if err != nil {
				t.Error(err)
			}
		}

		// Create the builder and set the extension dir
		builder := NewExtensionInstallerBuilder().ExtensionDir("testExtDir")

		// Set the prune flag in the builder
		builder.Prune()

		// Only specify one of the extensions to install
		builder.extensionsToInstall = make(Extensions)
		builder.extensionsToInstall["test-extension"] = "3.2.1"

		// Cache the extensions in the builder object
		err = builder.cacheExtensionsDir()
		if err != nil {
			t.Error(err)
		}

		// Call cleanExtensionDirectory
		err = builder.cleanExtensionDirectory()
		if err != nil {
			t.Error(err)
		}

		// Validate that the extensions were cleaned correctly
		// test-extension1_1.2.3 should have been deleted
		if _, err = os.Stat("testExtDir/test-extension1_1.2.3"); !os.IsNotExist(err) {
			t.Errorf("Expected test-extension1_1.2.3 to be deleted")
		}

		// Delete the test directory
		err = cleanupDir()
		if err != nil {
			t.Error(err)
		}
	})
}

func TestExtensionInstaller_Confirm(t *testing.T) {
	builder := NewExtensionInstallerBuilder().AutoConfirm(true)

	if builder.requiresConfirmation {
		t.Error("Expected requiresConfirmation to be false")
	}
}

func TestExtensionInstaller_Token(t *testing.T) {
	builder := NewExtensionInstallerBuilder().Token("testToken")

	if builder.githubToken != "testToken" {
		t.Errorf("Expected githubToken to be %s, got %s", "testToken", builder.githubToken)
	}
}

func TestExtensionInstaller_Org(t *testing.T) {
	builder := NewExtensionInstallerBuilder().Org("testOrg")

	if builder.githubOrg != "testOrg" {
		t.Errorf("Expected githubOrg to be %s, got %s", "testOrg", builder.githubOrg)
	}
}

func TestExtensionInstaller_Upgrade(t *testing.T) {
	builder := NewExtensionInstallerBuilder().Upgrade()

	if !builder.upgrade {
		t.Error("Expected upgrade to be true")
	}
}

func TestExtensionInstaller_Prune(t *testing.T) {
	builder := NewExtensionInstallerBuilder().Prune()

	if !builder.prune {
		t.Error("Expected prune to be true")
	}
}

func TestExtensionInstaller_PreFlight(t *testing.T) {
	t.Run("ErrorsPresent", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().ExtensionDir("")

		builder.errs = append(builder.errs, fmt.Errorf("test error"))

		err := builder.PreFlight()
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("InvalidDelimeter", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().ExtensionDir("").
			Extensions([]string{"test-extension:1.0.0"})

		err := builder.PreFlight()
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("NoErrors", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().ExtensionDir("")

		err := builder.PreFlight()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("NoExtensionVersion", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().ExtensionDir("")

		builder.extensionsToInstall = make(Extensions)
		builder.extensionsToInstall["test-extension"] = ""

		err := builder.PreFlight()
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("ExtensionVersionIsLatest", func(t *testing.T) {
		extension, err := NewGithubReleaseFetcher("", GetGithubToken()).GetFirstExtension()
		if err != nil {
			t.Error(err)
		}

		builder := NewExtensionInstallerBuilder().ExtensionDir("").Token(GetGithubToken())

		builder.extensionsToInstall = make(Extensions)
		builder.extensionsToInstall[extension] = "latest"

		err = builder.PreFlight()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("ExtensionDoesNotExist", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder()

		builder.extensionsToInstall = make(Extensions)
		builder.extensionsToInstall["theres-no-way-this-extension-exists"] = "1.0.0"

		err := builder.PreFlight()
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})
}

func TestExtensionInstaller_PromptForExtensions(t *testing.T) {
	builder := NewExtensionInstallerBuilder().
		InteractiveMode(true).
		Token(GetGithubToken()).
		AutoConfirm(false)

	tests := []cmdtest.PromptTest{
		{
			Name: "SelectExtensions",
			Procedure: func(console *cmdtest.Console) {
				console.ExpectString("Select the extensions to install - the most recent versions are displayed  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]")
				console.Send(string(terminal.KeyArrowDown))
				console.SendLine(" ")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if len(builder.extensionsToInstall) != 1 {
					return fmt.Errorf("expected extensionsToInstall to have length 1, got %d", len(builder.extensionsToInstall))
				}

				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {

			cmdtest.RunTest(t, test.Procedure, func() error {
				return builder.PromptForExtensions()
			})

			if test.CheckProcedure != nil {
				err := test.CheckProcedure()
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestExtensionInstaller_Run(t *testing.T) {
	t.Run("NonInteractiveMode", func(t *testing.T) {
		extension, err := NewGithubReleaseFetcher("", GetGithubToken()).GetFirstExtension()
		if err != nil {
			t.Error(err)
		}

		builder := NewExtensionInstallerBuilder().
			InteractiveMode(false).
			Extensions([]string{string(extension) + "@latest"}).
			ExtensionDir("testExtDir").
			Token(GetGithubToken())

		err = builder.PreFlight()
		if err != nil {
			t.Error(err)
		}

		err = builder.Run()
		if err != nil {
			t.Error(err)
		}

		// Remove the test directory
		err = os.RemoveAll("testExtDir")
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("InteractiveMode", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().
			InteractiveMode(true).
			ExtensionDir("testExtDir").
			Token(GetGithubToken()).
			// No autoconfirm
			AutoConfirm(false)

		procedure := func(console *cmdtest.Console) {
			console.ExpectString("Select the extensions to install - the most recent versions are displayed  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]")
			console.Send(string(terminal.KeyArrowDown))
			console.SendLine(" ")
			console.Send(string(terminal.KeyArrowDown))
			console.SendLine(" ")
			console.Send(string(terminal.KeyArrowDown))
			console.Send(string(terminal.KeyArrowDown))
			console.SendLine(" ")
			console.Send(string(terminal.KeyArrowDown))
			console.SendLine(" ")
			console.Send(string(terminal.KeyArrowDown))
			console.SendLine(" ")
			console.ExpectString("Install extensions? [? for help] (y/N) ")
			console.SendLine("y")
			console.ExpectEOF()
		}

		cmdtest.RunTest(t, procedure, func() error {
			return builder.Run()
		})

		// Remove the test directory
		err := os.RemoveAll("testExtDir")
		if err != nil {
			t.Error(err)
		}
	})
}
