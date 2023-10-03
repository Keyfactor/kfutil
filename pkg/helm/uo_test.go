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

package helm

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
	"kfutil/pkg/cmdtest"
	"kfutil/pkg/cmdutil/extensions"
	"kfutil/pkg/cmdutil/flags"
	"os"
	"reflect"
	"strings"
	"testing"
)

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}

func GetGithubToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func (b *InteractiveUOValueBuilder) ClearValues() {
	b.newValues = UniversalOrchestratorHelmValues{}
	b.newValues.LogLevel = "Info"
}

func ExpectSelectOption(console *cmdtest.Console) {
	console.ExpectString("Select an option:  [Use arrows to move, type to filter, ? for more help]")
}

func TestNewUniversalOrchestratorHelmValueBuilder(t *testing.T) {
	t.Run("NewToolBuilder", func(t *testing.T) {
		builder := NewUniversalOrchestratorHelmValueBuilder()

		if !reflect.DeepEqual(builder, &InteractiveUOValueBuilder{}) {
			t.Error("NewToolBuilder() did not return a ToolBuilder")
		}
	})

	t.Run("CommandHostname", func(t *testing.T) {
		builder := NewUniversalOrchestratorHelmValueBuilder().CommandHostname("test")

		if !reflect.DeepEqual(builder.commandHostname, "test") {
			t.Error("CommandHostname() did not set commandHostname")
		}
	})

	t.Run("OverrideFile", func(t *testing.T) {
		builder := NewUniversalOrchestratorHelmValueBuilder().OverrideFile("test")

		if !reflect.DeepEqual(builder.overrideFile, "test") {
			t.Error("OverrideFile() did not set overrideFile")
		}
	})

	t.Run("Token", func(t *testing.T) {
		builder := NewUniversalOrchestratorHelmValueBuilder().Token("test")

		if !reflect.DeepEqual(builder.token, "test") {
			t.Error("Token() did not set token")
		}
	})

	t.Run("Extensions", func(t *testing.T) {
		t.Run("ValidExtensions", func(t *testing.T) {
			builder := NewUniversalOrchestratorHelmValueBuilder().Extensions([]string{"test-extension@1.0.0"})

			testExtensions := make(extensions.Extensions)
			testExtensions["test-extension"] = "1.0.0"

			if !reflect.DeepEqual(builder.extensions, testExtensions) {
				t.Errorf("Expected extensionsToInstall to be %v, got %v", testExtensions, builder.extensions)
			}
		})

		t.Run("InvalidExtensions", func(t *testing.T) {
			builder := NewUniversalOrchestratorHelmValueBuilder().Extensions([]string{"test-extension@1.0.0@1.0.0"})

			if len(builder.errs) == 0 {
				t.Error("Expected error, got none")
			}
		})

		t.Run("NilExtensions", func(t *testing.T) {
			builder := NewUniversalOrchestratorHelmValueBuilder().Extensions(nil)

			if len(builder.extensions) != 0 {
				t.Errorf("Expected extensionsToInstall to be nil, got %v", builder.extensions)
			}
		})
	})

	t.Run("Values", func(t *testing.T) {
		t.Run("ReadError", func(t *testing.T) {
			fileFlags := flags.FilenameOptions{
				Filenames: []string{"test"},
			}

			builder := NewUniversalOrchestratorHelmValueBuilder().Values(fileFlags)

			if len(builder.errs) == 0 {
				t.Error("Values() did not set errs")
			}
		})

		t.Run("UnmarshalError", func(t *testing.T) {
			fileFlags := flags.FilenameOptions{
				Filenames: []string{"https://raw.githubusercontent.com/Keyfactor/kfutil/main/README.md"},
			}

			builder := NewUniversalOrchestratorHelmValueBuilder().Values(fileFlags)

			if !strings.Contains(builder.errs[0].Error(), "error unmarshalling values") {
				t.Error("Values() did not set errs despite unmarshal error")
			}
		})

		t.Run("Success", func(t *testing.T) {
			testFile := "./testFile.yaml"

			// Create blank file to read from
			_, err := os.Create(testFile)
			if err != nil {
				t.Error(err)
			}

			fileFlags := flags.FilenameOptions{
				Filenames: []string{testFile},
			}

			// A blank file is valid YAML, so this should not set errs
			builder := NewUniversalOrchestratorHelmValueBuilder().Values(fileFlags)

			// Delete the test file
			err = os.Remove(testFile)
			if err != nil {
				t.Error(err)
			}

			if len(builder.errs) > 0 {
				t.Error("Values() set errs despite success")
			}
		})
	})

	t.Run("NonInteractiveMode", func(t *testing.T) {
		extension, err := extensions.NewGithubReleaseFetcher("", GetGithubToken()).GetFirstExtension()
		if err != nil {
			return
		}

		builder := NewUniversalOrchestratorHelmValueBuilder().
			InteractiveMode(false).
			Extensions([]string{string(extension)})

		err = builder.PreFlight()
		if err != nil {
			t.Error(err)
		}

		_, err = builder.Build()
		if err != nil {
			t.Error(err)
		}

		if len(builder.newValues.InitContainers) != 1 {
			t.Errorf("expected 1 init container, got %d", len(builder.newValues.InitContainers))
		}
	})
}

func TestInteractiveUOValueBuilder_staticBuild(t *testing.T) {
	t.Run("NilExtensions", func(t *testing.T) {
		builder := NewTestBuilder()

		err := builder.staticBuild()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("EmptyExtensions", func(t *testing.T) {
		builder := NewTestBuilder()

		err := builder.staticBuild()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("InvalidExtensions", func(t *testing.T) {
		// TODO
	})
}

func TestInteractiveUOValueBuilder(t *testing.T) {
	t.Helper()
	interactiveBuilder := NewTestBuilder()

	tests := []cmdtest.PromptTest{
		{
			Name: "ConfigureUOName",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Configure UO Name")
				ExpectSelectOption(console)
				console.SendLine("Change Base Orchestrator Name")
				console.ExpectString("Enter the name of the chart ")
				console.SendLine("test")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if interactiveBuilder.newValues.BaseOrchestratorName != "test" {
					return fmt.Errorf("expected base orchestrator name to be test, got %s", interactiveBuilder.newValues.BaseOrchestratorName)
				}
				return nil
			},
		},
		{
			Name: "ChangeCompleteOrchestratorName",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Configure UO Name")
				ExpectSelectOption(console)
				console.SendLine("Change Complete Orchestrator Name")
				console.ExpectString("Enter the name of the chart ")
				console.SendLine("test")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if interactiveBuilder.newValues.CompleteName != "test" {
					return fmt.Errorf("expected complete orchestrator name to be test, got %s", interactiveBuilder.newValues.CompleteName)
				}
				return nil
			},
		},
		{
			Name: "ChangeReplicaCount",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Change Replica Count")
				console.ExpectString("Enter a non-zero number of Orchestrator replicas to create (0) ")
				console.SendLine("2")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if interactiveBuilder.newValues.ReplicaCount != 2 {
					return fmt.Errorf("expected replica count to be 2, got %d", interactiveBuilder.newValues.ReplicaCount)
				}
				return nil
			},
		},
		{
			Name: "ChangeLogLevel",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Change Log Level")
				console.ExpectString("Select the log level of the Universal Orchestrator container  [Use arrows to move, type to filter]")
				console.SendLine("Debug")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if interactiveBuilder.newValues.LogLevel != "Debug" {
					return fmt.Errorf("expected log level to be Debug, got %s", interactiveBuilder.newValues.LogLevel)
				}
				return nil
			},
		},
		{
			Name: "ChangeImageRepository",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Change Image")
				ExpectSelectOption(console)
				console.SendLine("Change Image Repository")
				console.ExpectString("Enter the repository of the Universal Orchestrator container image ")
				console.SendLine("test")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if interactiveBuilder.newValues.Image.Repository != "test" {
					return fmt.Errorf("expected image repository to be test, got %s", interactiveBuilder.newValues.Image.Repository)
				}
				return nil
			},
		},
		{
			Name: "ChangeImageTag",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Change Image")
				ExpectSelectOption(console)
				console.SendLine("Change Image Tag")
				console.ExpectString("Enter the tag of the Universal Orchestrator container image ")
				console.SendLine("test")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if interactiveBuilder.newValues.Image.Tag != "test" {
					return fmt.Errorf("expected image tag to be test, got %s", interactiveBuilder.newValues.Image.Tag)
				}
				return nil
			},
		},
		{
			Name: "ChangeImagePullPolicy",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Change Image")
				ExpectSelectOption(console)
				console.SendLine("Change Image Pull Policy")
				console.ExpectString("Enter the pull policy to use when pulling the Universal Orchestrator container image ")
				console.SendLine("test")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if interactiveBuilder.newValues.Image.PullPolicy != "test" {
					return fmt.Errorf("expected image pull policy to be test, got %s", interactiveBuilder.newValues.Image.PullPolicy)
				}
				return nil
			},
		},
		{
			Name: "ChangeAuthSecretName",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Change Auth Settings")
				ExpectSelectOption(console)
				console.SendLine("Change Authentication Secret Name")
				console.ExpectString("Enter the name of the K8s secret containing credentials for Command ")
				console.SendLine("test")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if interactiveBuilder.newValues.Auth.SecretName != "test" {
					return fmt.Errorf("expected auth secret name to be test, got %s", interactiveBuilder.newValues.Auth.SecretName)
				}
				return nil
			},
		},
		{
			Name: "UseOauthIDP",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Change Auth Settings")
				ExpectSelectOption(console)
				console.SendLine("Use OAuth/IDP for Authentication to Command")
				console.ExpectString("Use OAuth/IDP for Authentication to Command? (y/N) ")
				console.SendLine("y")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if !interactiveBuilder.newValues.Auth.UseOauthAuthentication {
					return fmt.Errorf("expected auth use oauth authentication to be true, got false")
				}
				return nil
			},
		},
		{
			Name: "ContainerRootCACertificateConfiguration",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Container Root CA Certificate Configuration")
				console.ExpectString("Enter the name of the configmap containing the CA chain (ca-roots) ")
				console.SendLine("root-ca-configmap")
				console.ExpectString("Enter the file name of the certificate inside the configmap. In most cases this should be ca-certificates.crt (ca-certificates.crt) ")
				console.SendLine("cert.pem")
				console.ExpectString("Enter the mount path where the certificate inside the configmap will be mounted (/etc/ssl/certs) ")
				console.SendLine("/etc/user/certs")
				console.ExpectString("Save changes? (Y/n) ")
				console.SendLine("Y")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if len(interactiveBuilder.newValues.Volumes) != 1 {
					return fmt.Errorf("expected 1 volume, got %d", len(interactiveBuilder.newValues.Volumes))
				}

				if interactiveBuilder.newValues.Volumes[0].ConfigMap.Name != "root-ca-configmap" {
					return fmt.Errorf("expected volume name to be root-ca-configmap, got %s", interactiveBuilder.newValues.Volumes[0].Name)
				}

				if interactiveBuilder.newValues.Volumes[0].ConfigMap.Items[0].Key != "cert.pem" {
					return fmt.Errorf("expected volume key to be cert.pem, got %s", interactiveBuilder.newValues.Volumes[0].ConfigMap.Items[0].Key)
				}

				if interactiveBuilder.newValues.Volumes[0].ConfigMap.Items[0].Path != "cert.pem" {
					return fmt.Errorf("expected volume path to be cert.pem, got %s", interactiveBuilder.newValues.Volumes[0].ConfigMap.Items[0].Path)
				}

				if interactiveBuilder.newValues.VolumeMounts[0].SubPath != "cert.pem" {
					return fmt.Errorf("expected volume mount name to be cert.pem, got %s", interactiveBuilder.newValues.VolumeMounts[0].SubPath)
				}

				return nil
			},
		},
		{
			Name: "ConfigureOrchestratorExtensions",
			Procedure: func(console *cmdtest.Console) {
				ExpectSelectOption(console)
				console.SendLine("Configure Orchestrator Extensions")
				console.ExpectString("Select the extensions to install - the most recent versions are displayed  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]")
				console.Send(string(terminal.KeyArrowDown))
				console.SendLine(" ")
				console.ExpectEOF()
			},
			CheckProcedure: func() error {
				if len(interactiveBuilder.newValues.InitContainers) != 1 {
					return fmt.Errorf("expected 1 init container, got %d", len(interactiveBuilder.newValues.InitContainers))
				}
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			interactiveBuilder.ClearValues()

			cmdtest.RunTest(t, test.Procedure, func() error {
				return interactiveBuilder.MainMenu()
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

func NewTestBuilder() *InteractiveUOValueBuilder {
	toolBuilder := NewUniversalOrchestratorHelmValueBuilder().
		CommandHostname("test").
		OverrideFile("test").
		Token(GetGithubToken())

	toolBuilder.exitAfterPrompt = true

	return toolBuilder
}
