package flags

import (
	"github.com/spf13/cobra"
	"testing"
)

func TestGetFlagBool(t *testing.T) {
	// Create a command with a bool flag
	cmd := &cobra.Command{}
	var boolFlag bool
	cmd.Flags().BoolVarP(&boolFlag, "boolflag", "b", boolFlag, "Boolean flag")

	t.Run("TestNoBoolFlag", func(t *testing.T) {
		b := GetFlagBool(cmd, "boolflag")
		if b != false {
			t.Errorf("got %v, want %v", b, false)
		}
	})

	t.Run("TestBoolFlag", func(t *testing.T) {
		err := cmd.Flags().Set("boolflag", "true")
		if err != nil {
			t.Errorf("error setting boolflag flag: %v", err)
		}
		b := GetFlagBool(cmd, "boolflag")
		if b != true {
			t.Errorf("got %v, want %v", b, true)
		}
	})
}

func TestGetFlagString(t *testing.T) {
	// Create a command with a string flag
	cmd := &cobra.Command{}
	var stringFlag string
	cmd.Flags().StringVarP(&stringFlag, "stringflag", "s", stringFlag, "String flag")

	t.Run("TestNoStringFlag", func(t *testing.T) {
		s := GetFlagString(cmd, "stringflag")
		if s != "" {
			t.Errorf("got %v, want %v", s, "")
		}
	})

	t.Run("TestStringFlag", func(t *testing.T) {
		err := cmd.Flags().Set("stringflag", "test")
		if err != nil {
			t.Errorf("error setting stringflag flag: %v", err)
		}
		s := GetFlagString(cmd, "stringflag")
		if s != "test" {
			t.Errorf("got %v, want %v", s, "test")
		}
	})
}
