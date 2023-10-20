package flags

import (
	"github.com/spf13/cobra"
	"os"
	"testing"
)

func TestGetDebugFlag(t *testing.T) {
	// Create a command with a debug flag
	cmd := &cobra.Command{}
	debug := false
	cmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")

	tests := []struct {
		name  string
		want  bool
		setup func()
	}{
		{
			name: "TestNoDebugFlag",
			want: false,
			setup: func() {
			},
		},
		{
			name: "TestDebugFlag",
			want: true,
			setup: func() {
				err := cmd.Flags().Set("debug", "true")
				if err != nil {
					t.Errorf("error setting debug flag: %v", err)
				}
			},
		},
		{
			name: "TestEnvDebugFlag",
			want: true,
			setup: func() {
				err := os.Setenv("KFUTIL_DEBUG", "true")
				if err != nil {
					t.Errorf("error setting env debug flag: %v", err)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setup()
			got := GetDebugFlag(cmd)
			if got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestGetNoPromptFlag(t *testing.T) {
	// Create a command with a no-prompt flag
	cmd := &cobra.Command{}
	debug := false
	cmd.Flags().BoolVarP(&debug, "no-prompt", "n", false, "Disable prompts")

	t.Run("TestNoPromptFlag", func(t *testing.T) {
		err := cmd.Flags().Set("no-prompt", "true")
		if err != nil {
			t.Errorf("error setting no-prompt flag: %v", err)
		}
		got := GetNoPromptFlag(cmd)
		if got != true {
			t.Errorf("got %v, want %v", got, true)
		}
	})
}

func TestGetProfileFlag(t *testing.T) {
	// Create a command with a profile flag
	cmd := &cobra.Command{}
	profile := ""
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "Set the profile")

	// Test the default profile
	t.Run("TestNoProfileFlag", func(t *testing.T) {
		got := GetProfileFlag(cmd)
		if got != "default" {
			t.Errorf("got %v, want %v", got, "default")
		}
	})

	// Test a custom profile
	t.Run("TestProfileFlag", func(t *testing.T) {
		err := cmd.Flags().Set("profile", "test")
		if err != nil {
			t.Errorf("error setting profile flag: %v", err)
		}
		got := GetProfileFlag(cmd)
		if got != "test" {
			t.Errorf("got %v, want %v", got, "test")
		}
	})
}

func TestGetConfigFlag(t *testing.T) {
	// Create a command with a config flag
	cmd := &cobra.Command{}
	config := ""
	cmd.Flags().StringVarP(&config, "config", "c", "", "Set the config file")

	t.Run("TestConfigFlag", func(t *testing.T) {
		err := cmd.Flags().Set("config", "test")
		if err != nil {
			t.Errorf("error setting config flag: %v", err)
		}
		got := GetConfigFlag(cmd)
		if got != "test" {
			t.Errorf("got %v, want %v", got, "test")
		}
	})
}
