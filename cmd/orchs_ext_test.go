/*
Copyright 2024 The Keyfactor Command Authors.

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

package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2/terminal"
	"kfutil/pkg/cmdtest"
	"kfutil/pkg/cmdutil/extensions"
	"os"
	"strings"
	"testing"
)

func isDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func setupExtensionDirectory(t *testing.T, dirName string) error {
	// Delete directory if it already exists
	if _, err := os.Stat(dirName); !os.IsNotExist(err) {
		err = os.RemoveAll(dirName)
		if err != nil {
			t.Error(err)
		}
	}

	return nil
}

func verifyExtensionDirectory(t *testing.T, dirName string) error {
	// Verify that the extension directory was created
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		t.Error(fmt.Sprintf("Extension directory %s was not created", dirName))
	}

	// Verify that the extension directory is not empty
	if empty, err := isDirEmpty(dirName); err != nil {
		t.Error(err)
	} else if empty {
		t.Error(fmt.Sprintf("Extension directory %s is empty", dirName))
	}

	// Remove extension directory
	err := os.RemoveAll(dirName)
	if err != nil {
		t.Error(err)
	}

	return nil
}

func TestOrchsExt(t *testing.T) {
	t.Run("TestOrchsExt_ExtensionFlag", func(t *testing.T) {
		extCmd := NewCmdOrchsExt()
		var debug bool
		extCmd.Flags().BoolVarP(&debug, "debug", "b", false, "debug")

		// Get an orchestrator name
		extension, err := extensions.NewGithubReleaseFetcher("", GetGithubToken()).GetFirstExtension()
		if err != nil {
			t.Error(err)
		}

		// Set up extension directory
		dirName := "testExtDir"
		err = setupExtensionDirectory(t, dirName)
		if err != nil {
			t.Error(err)
		}

		args := []string{"-t", GetGithubToken(), "-e", fmt.Sprintf("%s@latest", extension), "-o", dirName, "-y"}

		_, err = cmdtest.TestExecuteCommand(t, extCmd, args...)
		if err != nil {
			t.Error(err)
		}

		err = verifyExtensionDirectory(t, dirName)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("TestOrchsExt_ConfigFile", func(t *testing.T) {
		extCmd := NewCmdOrchsExt()
		var debug bool
		extCmd.Flags().BoolVarP(&debug, "debug", "b", false, "debug")

		// Get an orchestrator name
		extension, err := extensions.NewGithubReleaseFetcher("", GetGithubToken()).GetFirstExtension()
		if err != nil {
			t.Error(err)
		}

		// Create config YAML if it doesn't exist
		if _, err = os.Stat("config.yaml"); os.IsNotExist(err) {
			file, err := os.Create("config.yaml")
			if err != nil {
				t.Error(err)
			}
			err = file.Close()
			if err != nil {
				t.Error(err)
			}
		}

		// Open config YAML
		file, err := os.OpenFile("config.yaml", os.O_RDWR, 0644)
		if err != nil {
			t.Error(err)
		}

		// Write config YAML
		_, err = file.Write([]byte(fmt.Sprintf("%s: latest\n", extension)))
		if err != nil {
			t.Error(err)
		}

		// Close config YAML
		err = file.Close()
		if err != nil {
			t.Error(err)
		}

		// Set up extension directory
		dirName := "testExtDir"
		err = setupExtensionDirectory(t, dirName)
		if err != nil {
			t.Error(err)
		}

		args := []string{"-t", GetGithubToken(), "-c", "config.yaml", "-o", dirName, "-y"}

		_, err = cmdtest.TestExecuteCommand(t, extCmd, args...)
		if err != nil {
			t.Error(err)
		}

		// Remove config YAML
		err = os.Remove("config.yaml")
		if err != nil {
			t.Error(err)
		}

		err = verifyExtensionDirectory(t, dirName)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("TestOrchsExt_Upgrades", func(t *testing.T) {
		extCmd := NewCmdOrchsExt()
		var debug bool
		extCmd.Flags().BoolVarP(&debug, "debug", "b", false, "debug")

		// Get an orchestrator name
		extension, err := extensions.NewGithubReleaseFetcher("", GetGithubToken()).GetFirstExtension()
		if err != nil {
			t.Fatal(err)
		}

		// Set up extension directory
		dirName := "testExtDir"
		err = setupExtensionDirectory(t, dirName)
		if err != nil {
			t.Fatal(err)
		}

		// Create a directory for the extension with a version that is not probable to be the latest
		extensionDir := fmt.Sprintf("%s/%s_%s", dirName, extension, "v0.48.289")
		err = os.MkdirAll(extensionDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Setup the command
		args := []string{"-t", GetGithubToken(), "-o", dirName, "-y", "-u"}
		_, err = cmdtest.TestExecuteCommand(t, extCmd, args...)
		if err != nil {
			t.Error(err)
		}

		// Verify that extensionDir does not exist, but a new directory with the latest version does
		if _, err = os.Stat(extensionDir); !os.IsNotExist(err) {
			t.Error(fmt.Sprintf("Extension directory %s was not removed", extensionDir))
		}

		entries, err := os.ReadDir(dirName)
		if err != nil {
			t.Error(err)
		}

		// Verify that the new directory exists
		newVersionPresent := false
		for _, entry := range entries {
			if entry.IsDir() && strings.Contains(entry.Name(), string(extension)) && !strings.Contains(entry.Name(), "v0.48.289") {
				newVersionPresent = true
			}
		}

		if !newVersionPresent {
			t.Error("New version of extension was not installed")
		}

		// Remove extension directory
		err = os.RemoveAll(dirName)
	})

	tests := []cmdtest.CommandTest{
		{
			PromptTest: cmdtest.PromptTest{
				Name: "TestOrchsExt_InteractiveMode",
				Procedure: func(console *cmdtest.Console) {
					//console.ExpectString("Select the extensions to install - the most recent versions are displayed  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]")
					console.Send(string(terminal.KeyArrowDown))
					console.SendLine(" ")
					console.SendLine("y")
				},
			},
			CommandArguments: []string{"-t", GetGithubToken(), "-o", "testExtDir"},
			Config: func() error {
				// Set up extension directory
				dirName := "testExtDir"
				err := setupExtensionDirectory(t, dirName)
				if err != nil {
					t.Error(err)
				}

				return nil
			},
			CheckProcedure: func(output []byte) error {
				err := verifyExtensionDirectory(t, "testExtDir")
				if err != nil {
					return err
				}

				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Skip()
			var output []byte
			var err error

			if test.Config != nil {
				err = test.Config()
				if err != nil {
					t.Error(err)
				}
			}

			extCmd := NewCmdOrchsExt()
			var debug bool
			extCmd.Flags().BoolVarP(&debug, "debug", "b", false, "debug")

			cmdtest.RunTest(t, test.Procedure, func() error {
				output, err = cmdtest.TestExecuteCommand(t, extCmd, test.CommandArguments...)
				if err != nil {
					return err
				}

				return nil
			})

			if test.CheckProcedure != nil {
				err = test.CheckProcedure(output)
				if err != nil {
					t.Error(err)
				}
			}
		})
	}
}
