package cmd

import (
	"fmt"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"kfutil/pkg/cmdtest"
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
