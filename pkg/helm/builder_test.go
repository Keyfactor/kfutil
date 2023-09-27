package helm

import (
	"kfutil/pkg/cmdutil/flags"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestNewToolBuilder(t *testing.T) {

	t.Run("NewToolBuilder", func(t *testing.T) {
		builder := NewToolBuilder()

		if !reflect.DeepEqual(builder, &ToolBuilder{}) {
			t.Error("NewToolBuilder() did not return a ToolBuilder")
		}
	})

	t.Run("CommandHostname", func(t *testing.T) {
		builder := NewToolBuilder().CommandHostname("test")

		if !reflect.DeepEqual(builder.commandHostname, "test") {
			t.Error("CommandHostname() did not set commandHostname")
		}
	})

	t.Run("OverrideFile", func(t *testing.T) {
		builder := NewToolBuilder().OverrideFile("test")

		if !reflect.DeepEqual(builder.overrideFile, "test") {
			t.Error("OverrideFile() did not set overrideFile")
		}
	})

	t.Run("Token", func(t *testing.T) {
		builder := NewToolBuilder().Token("test")

		if !reflect.DeepEqual(builder.token, "test") {
			t.Error("Token() did not set token")
		}
	})

	t.Run("Values", func(t *testing.T) {
		t.Run("ReadError", func(t *testing.T) {
			fileFlags := flags.FilenameOptions{
				Filenames: []string{"test"},
			}

			builder := NewToolBuilder().Values(fileFlags)

			if len(builder.errs) == 0 {
				t.Error("Values() did not set errs")
			}
		})

		t.Run("UnmarshalError", func(t *testing.T) {
			fileFlags := flags.FilenameOptions{
				Filenames: []string{"https://raw.githubusercontent.com/Keyfactor/kfutil/main/README.md"},
			}

			builder := NewToolBuilder().Values(fileFlags)

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
			builder := NewToolBuilder().Values(fileFlags)

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
}

func NewTestBuilder() *InteractiveUOValueBuilder {
	builder := NewToolBuilder().
		CommandHostname("test").
		OverrideFile("test").
		Token(GetGithubToken())

	toolBuilder := NewUniversalOrchestratorHelmValueBuilder(builder)
	toolBuilder.exitAfterPrompt = true

	return toolBuilder
}
