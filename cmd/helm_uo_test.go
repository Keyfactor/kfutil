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

package cmd

import (
	"fmt"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"kfutil/pkg/cmdtest"
	"kfutil/pkg/cmdutil/extensions"
	"kfutil/pkg/helm"
	"os"
	"testing"
)

var filename = fmt.Sprintf("https://raw.githubusercontent.com/Keyfactor/containerized-uo-deployment-dev/main/universal-orchestrator/values.yaml?token=%s", os.Getenv("TOKEN"))

func TestHelmUo_SaveAndExit(t *testing.T) {
	t.Skip()
	tests := []cmdtest.CommandTest{
		{
			PromptTest: cmdtest.PromptTest{
				Name: "TestHelmUo_SaveAndExit",
				Procedure: func(console *cmdtest.Console) {
					console.SendLine("Save and Exit")
				},
			},
			CommandArguments: []string{"helm", "uo", "-t", GetGithubToken(), "-f", filename},
			CheckProcedure: func(output []byte) error {
				// Compile output to UniversalOrchestratorHelmValues
				values := helm.UniversalOrchestratorHelmValues{}
				err := yaml.Unmarshal(output, &values)
				if err != nil {
					return err
				}

				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var output []byte
			var err error

			cmdtest.RunTest(t, test.Procedure, func() error {
				output, err = cmdtest.TestExecuteCommand(t, RootCmd, test.CommandArguments...)
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

func TestHelmUo(t *testing.T) {
	uoCmd := NewCmdHelmUo()
	var debug, noPrompt, exp bool
	var profile, config string
	uoCmd.Flags().BoolVarP(&debug, "debug", "b", false, "debug")
	uoCmd.Flags().BoolVarP(&noPrompt, "no-prompt", "y", false, "no-prompt")
	uoCmd.Flags().StringVarP(&profile, "profile", "p", "", "profile")
	uoCmd.Flags().StringVarP(&config, "config", "c", "", "config")
	uoCmd.Flags().BoolVarP(&exp, "exp", "", false, "exp")

	// Get an orchestrator name
	extension, err := extensions.NewGithubReleaseFetcher("", GetGithubToken()).GetFirstExtension()
	if err != nil {
		t.Error(err)
	}

	// Create a blank YAML file on disk
	_, err = os.Create("test.yaml")
	if err != nil {
		t.Error(err)
	}

	args := []string{"--exp", "-t", GetGithubToken(), "-e", fmt.Sprintf("%s", extension), "-f", "test.yaml"}

	valuesYamlString, err := cmdtest.TestExecuteCommand(t, uoCmd, args...)
	if err != nil {
		t.Error(err)
	}

	// Serialize the values.yaml string to a struct
	values := helm.UniversalOrchestratorHelmValues{}
	err = yaml.Unmarshal(valuesYamlString, &values)
	if err != nil {
		t.Error(err)
	}

	// Check that the values.yaml struct has the extension
	if len(values.InitContainers) == 0 {
		t.Errorf("Expected at least one init container, got %v", len(values.InitContainers))
	}

	// Remove the test.yaml file
	err = os.Remove("test.yaml")
	if err != nil {
		t.Error(err)
	}
}

func GetGithubToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func TestHelmUoFlags_AddFlags(t *testing.T) {
	flagSet := pflag.FlagSet{}
	huf := NewHelmUoFlags()
	huf.AddFlags(&flagSet)

	// Check FilenameFlags
	getStringSlice, err := flagSet.GetStringSlice("values")
	if err != nil {
		t.Error(err)
	}

	if len(getStringSlice) != 0 {
		t.Errorf("Expected empty string slice, got %v", len(getStringSlice))
	}

	// Check GithubToken
	getString, err := flagSet.GetString("token")
	if err != nil {
		t.Error(err)
	}

	if getString != "" {
		t.Errorf("Expected empty string, got %s", getString)
	}

	// Check OutPath
	getString, err = flagSet.GetString("out")
	if err != nil {
		t.Error(err)
	}
}
