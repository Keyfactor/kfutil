package flags

import (
	"github.com/spf13/cobra"
	"kfutil/pkg/cmdutil"
	"os"
	"reflect"
	"testing"
)

// TestFilenameFlags tests the FilenameFlags interface with its intended use
func TestFilenameFlags(t *testing.T) {
	// Create a new FilenameFlags interface
	name := "testfile"
	shorthand := "f"
	usage := "test usage"
	var filenames []string

	t.Run("TestStdin", func(t *testing.T) {

		// First, write some data to stdin

		testBytes := []byte("TestStdin")

		// Direct stdin to Stdin of the test case
		readEnd, writeEnd, _ := os.Pipe()

		// Backup the original stdin and then replace with our pipe
		originalStdin := os.Stdin
		defer func() { os.Stdin = originalStdin }()
		os.Stdin = readEnd

		go func() {
			defer func(writeEnd *os.File) {
				err := writeEnd.Close()
				if err != nil {
					t.Error(err)
					return
				}
			}(writeEnd)
			_, err := writeEnd.Write(testBytes)
			if err != nil {
				t.Error(err)
				return
			}
		}()

		// Next, set up a new command and add the flags
		f := NewFilenameFlags(name, shorthand, usage, filenames)

		cmd := NewCommand(func(cmd *cobra.Command, args []string) error {
			// Function that is run when the command is executed

			// Convert the flags to options
			options := f.ToOptions()
			read, err := options.Read()
			if err != nil {
				return err
			}

			// Compare the read bytes to the test bytes
			if !reflect.DeepEqual(read, testBytes) {
				t.Errorf("got %v, want %v", read, testBytes)
			}

			return nil
		})

		// Add the flags to the command
		f.AddFlags(cmd.Flags())

		// Add filename argument to the command with the value "-" to indicate stdin
		cmd.SetArgs([]string{"-f", "-"})

		// Execute the command
		err := cmd.Execute()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("TestFile", func(t *testing.T) {
		testFile := "./testFile.txt"

		// Create a test file
		file, err := os.Create(testFile)
		if err != nil {
			t.Error(err)
		}

		// Write to the test file
		testBytes := []byte("Hello, World!")
		_, err = file.Write(testBytes)
		if err != nil {
			t.Error(err)
		}

		// Next, set up a new command and add the flags
		f := NewFilenameFlags(name, shorthand, usage, filenames)

		cmd := NewCommand(func(cmd *cobra.Command, args []string) error {
			// Function that is run when the command is executed

			// Convert the flags to options
			options := f.ToOptions()
			read, err := options.Read()
			if err != nil {
				return err
			}

			// Compare the read bytes to the test bytes
			if !reflect.DeepEqual(read, testBytes) {
				t.Errorf("got %v, want %v", read, testBytes)
			}

			return nil
		})

		// Add the flags to the command
		f.AddFlags(cmd.Flags())

		// Add filename argument to the command with the value "testFile.txt"
		cmd.SetArgs([]string{"-f", testFile})

		// Execute the command
		err = cmd.Execute()
		if err != nil {
			t.Error(err)
		}

		// Delete the test file
		err = os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("TestURL", func(t *testing.T) {
		testUrl := "https://raw.githubusercontent.com/Keyfactor/kfutil/main/README.md"

		// Download the file to compare later
		testBytes, err := cmdutil.NewSimpleRestClient().Get(testUrl)
		if err != nil {
			t.Error(err)
		}

		// Next, set up a new command and add the flags
		f := NewFilenameFlags(name, shorthand, usage, filenames)

		cmd := NewCommand(func(cmd *cobra.Command, args []string) error {
			// Function that is run when the command is executed

			// Convert the flags to options
			options := f.ToOptions()
			read, err := options.Read()
			if err != nil {
				return err
			}

			// Compare the read bytes to the test bytes
			if !reflect.DeepEqual(read, testBytes) {
				t.Errorf("got %v, want %v", read, testBytes)
			}

			return nil
		})

		// Add the flags to the command
		f.AddFlags(cmd.Flags())

		// Add filename argument to the command with the value "https://raw.githubusercontent.com/Keyfactor/kfutil/main/README.md"
		cmd.SetArgs([]string{"-f", testUrl})

		// Execute the command
		err = cmd.Execute()
		if err != nil {
			t.Error(err)
		}
	})
}

func NewCommand(runE func(cmd *cobra.Command, args []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "test",
		Long:  "test",
		RunE:  runE,
	}
	return cmd
}

func TestFilenameOptions_IsEmpty(t *testing.T) {
	t.Run("TestEmpty", func(t *testing.T) {
		options := FilenameOptions{}
		if !options.IsEmpty() {
			t.Errorf("got %v, want %v", options.IsEmpty(), true)
		}
	})

	t.Run("TestNotEmpty", func(t *testing.T) {
		options := FilenameOptions{
			Filenames: []string{"test"},
		}
		if options.IsEmpty() {
			t.Errorf("got %v, want %v", options.IsEmpty(), false)
		}
	})
}

func TestFilenameOptions_Merge(t *testing.T) {
	t.Run("TestNil", func(t *testing.T) {
		options := FilenameOptions{
			Filenames: []string{"test"},
		}
		options.Merge(nil)
		if options.IsEmpty() {
			t.Errorf("got %v, want %v", options.IsEmpty(), false)
		}
	})

	t.Run("TestEmpty", func(t *testing.T) {
		options := FilenameOptions{
			Filenames: []string{"test"},
		}
		other := FilenameOptions{}
		options.Merge(&other)
		if options.IsEmpty() {
			t.Errorf("got %v, want %v", options.IsEmpty(), false)
		}
	})

	t.Run("TestNotEmpty", func(t *testing.T) {
		options := FilenameOptions{}
		other := FilenameOptions{
			Filenames: []string{"test"},
		}
		options.Merge(&other)
		if options.IsEmpty() {
			t.Errorf("got %v, want %v", options.IsEmpty(), false)
		}
	})
}

func TestFilenameOptions_Validate(t *testing.T) {
	t.Run("TestEmpty", func(t *testing.T) {
		options := FilenameOptions{}
		err := options.Validate()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("TestNotEmpty", func(t *testing.T) {
		options := FilenameOptions{
			Filenames: []string{"test"},
		}
		err := options.Validate()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("TestEmptyString", func(t *testing.T) {
		options := FilenameOptions{
			Filenames: []string{""},
		}
		err := options.Validate()
		if err == nil {
			t.Errorf("Error expected")
		}
	})
}

func TestFilenameOptions_Read(t *testing.T) {
	t.Run("TestEmpty", func(t *testing.T) {
		options := FilenameOptions{}
		_, err := options.Read()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("TestInvalidUrl", func(t *testing.T) {
		options := FilenameOptions{
			Filenames: []string{"http://exa^mple.com"},
		}
		_, err := options.Read()
		if err == nil {
			t.Errorf("Error expected")
		}
	})
}
