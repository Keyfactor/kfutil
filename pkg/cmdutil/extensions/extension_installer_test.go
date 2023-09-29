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

func TestNewExtensionInstallerBuilder(t *testing.T) {
	builder := NewExtensionInstallerBuilder()

	if builder == nil {
		t.Error("Expected builder to not be nil")
	}

	if !reflect.DeepEqual(*builder, ExtensionInstaller{}) {
		t.Errorf("Expected builder to be %v, got %v", ExtensionInstaller{}, *builder)
	}
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
		extensions["test-extension"] = []Version{"1.0.0"}

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
		if versions, ok := builder.extensionsToInstall["test-extension"]; ok {
			if len(versions) != 1 {
				t.Errorf("Expected versions to have length 1, got %d", len(versions))
			}
			if versions[0] != "1.0.0" {
				t.Errorf("Expected version to be %s, got %s", "1.0.0", versions[0])
			}
		} else {
			t.Errorf("Expected test-extension to be in extensionsToInstall")
		}

		if versions, ok := builder.extensionsToInstall["test-extension1"]; ok {
			if len(versions) != 1 {
				t.Errorf("Expected versions to have length 1, got %d", len(versions))
			}
			if versions[0] != "latest" {
				t.Errorf("Expected version to be %s, got %s", "latest", versions[0])
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
	// Delete the test directory if it exists
	if _, err := os.Stat("testExtDir"); !os.IsNotExist(err) {
		err = os.Remove("testExtDir")
		if err != nil {
			t.Error(err)
		}
	}

	builder := NewExtensionInstallerBuilder().ExtensionDir("testExtDir")

	if builder.extensionDirName != "testExtDir" {
		t.Errorf("Expected extensionDir to be %s, got %s", "test", builder.extensionDirName)
	}

	// Verify that the directory was created
	if _, err := os.Stat("testExtDir"); os.IsNotExist(err) {
		t.Errorf("Expected directory %s to exist, got %v", "testExtDir", err)
	}

	// Delete the test directory
	err := os.RemoveAll("testExtDir")
	if err != nil {
		t.Error(err)
	}
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

func TestExtensionInstaller_PreFlight(t *testing.T) {
	t.Run("ErrorsPresent", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder().ExtensionDir("")

		builder.errs = append(builder.errs, fmt.Errorf("test error"))

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
		builder.extensionsToInstall["test-extension"] = make([]Version, 0)

		err := builder.PreFlight()
		if err == nil {
			t.Errorf("Expected error, got none")
		}
	})

	t.Run("ExtensionVersionIsLatest", func(t *testing.T) {
		extension, err := NewGithubReleaseFetcher(GetGithubToken()).GetFirstExtension()
		if err != nil {
			t.Error(err)
		}

		builder := NewExtensionInstallerBuilder().ExtensionDir("").Token(GetGithubToken())

		builder.extensionsToInstall = make(Extensions)
		builder.extensionsToInstall[extension] = []Version{"latest"}

		err = builder.PreFlight()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("ExtensionDoesNotExist", func(t *testing.T) {
		builder := NewExtensionInstallerBuilder()

		builder.extensionsToInstall = make(Extensions)
		builder.extensionsToInstall["theres-no-way-this-extension-exists"] = []Version{"1.0.0"}

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
		extension, err := NewGithubReleaseFetcher(GetGithubToken()).GetFirstExtension()
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
