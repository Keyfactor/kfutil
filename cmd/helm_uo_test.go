package cmd

import (
	"gopkg.in/yaml.v3"
	"kfutil/pkg/helm"
	"os"
	"strings"
	"testing"
)

func TestHelmUo_SaveAndExit(t *testing.T) {
	helmUo := []string{"helm", "uo"}

	cases := []struct {
		Title      string
		SubCommand []string
		Stdin      []string
	}{{
		Title:      "Test default helm uo command",
		Stdin:      []string{"Save and Exit\n"},
		SubCommand: helmUo,
	}}

	for _, tc := range cases {
		t.Run(tc.Title, func(t *testing.T) {

			// Direct stdin to Stdin of the test case
			readEnd, writeEnd, _ := os.Pipe()

			// Backup the original stdin and then replace with our pipe
			originalStdin := os.Stdin
			defer func() { os.Stdin = originalStdin }()
			os.Stdin = readEnd

			// Write to the pipe in a separate go-routine so it doesn't block.
			go func() {
				defer func(writeEnd *os.File) {
					err := writeEnd.Close()
					if err != nil {
						t.Error(err)
						return
					}
				}(writeEnd)
				_, err := writeEnd.WriteString(strings.Join(tc.Stdin, ""))
				if err != nil {
					t.Error(err)
					return
				}
			}()

			bytes, err := testExecuteCommand(t, RootCmd, tc.SubCommand...)
			if err != nil {
				t.Fatal(err)
			}

			// Serialize the bytes into a UniversalOrchestratorHelmValues struct
			var out helm.UniversalOrchestratorHelmValues
			err = yaml.Unmarshal(bytes, &out)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
