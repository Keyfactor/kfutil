package helm

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2/terminal"
	"kfutil/pkg/cmdtest"
	"os"
	"testing"
)

func TestInteractiveUOValueBuilder_selectExtensionsHandler(t *testing.T) {
	t.Skip()
	interactiveBuilder := NewTestBuilder()

	tests := []cmdtest.PromptTest{
		{
			Name: "Select an extension",
			Procedure: func(console *cmdtest.Console) {
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
				return interactiveBuilder.selectExtensionsHandler()
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

func GetGithubToken() string {
	return os.Getenv("GITHUB_TOKEN")
}

func TestNewGithubReleaseFetcher(t *testing.T) {
	t.Skip()
	fetcher := NewGithubReleaseFetcher(GetGithubToken())

	list, err := fetcher.GetExtensionList()
	if err != nil {
		t.Error(err)
	}

	if len(list) == 0 {
		t.Error("No extensions returned")
	}
}

func TestGithubReleaseFetcher_Get(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		errorExpected bool
		object        any
	}{
		{
			name:          "ValidURL",
			url:           "https://api.github.com/orgs/keyfactor/repos?type=public&page=1&per_page=5",
			errorExpected: false,
			object:        []gitHubRepo{},
		},
		{
			name:          "InvalidURL",
			url:           "https://api.git^hub.com/orgs/keyfactor/repos?type=public&page=1&per_page=5",
			errorExpected: true,
			object:        []gitHubRepo{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := NewGithubReleaseFetcher(GetGithubToken())

			err := f.Get(test.url, &test.object)
			if err != nil && !test.errorExpected {
				t.Errorf("Unexpected error: %s", err)
			}

			if err == nil && test.errorExpected {
				t.Error("Expected error, got none")
			}
		})
	}
}
