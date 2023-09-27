package helm

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2/core"
	"kfutil/pkg/cmdtest"
	"testing"
)

func init() {
	// disable color output for all prompts to simplify testing
	core.DisableColor = true
}

func (b *InteractiveUOValueBuilder) ClearValues() {
	b.newValues = UniversalOrchestratorHelmValues{}
	b.newValues.LogLevel = "Info"
}

func ExpectSelectOption(console *cmdtest.Console) {
	console.ExpectString("Select an option:  [Use arrows to move, type to filter, ? for more help]")
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
