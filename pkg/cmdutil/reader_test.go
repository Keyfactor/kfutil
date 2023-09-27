package cmdutil

import (
	"net/url"
	"os"
	"reflect"
	"testing"
)

func TestNewReaderBuilder(t *testing.T) {
	builder := NewReaderBuilder()

	t.Run("TestNewReaderBuilder", func(t *testing.T) {
		// Test that the builder is initialized with an empty list of readers.
		builder.ClearReaders()

		// Empty slice of readers (nil)
		var want []Reader

		if !reflect.DeepEqual(builder.readers, want) {
			t.Errorf("got %v, want %v", builder.readers, want)
		}
	})

	t.Run("TestStdin", func(t *testing.T) {
		builder.ClearReaders()

		// Select Stdin reader
		builder.Stdin()

		testString := "TestStdin"

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
			_, err := writeEnd.WriteString(testString)
			if err != nil {
				t.Error(err)
				return
			}
		}()

		got, err := builder.Read()
		if err != nil {
			t.Error(err)
		}

		if !reflect.DeepEqual(got, []byte(testString)) {
			t.Errorf("got %v, want %v", got, testString)
		}
	})

	t.Run("TestURL", func(t *testing.T) {
		builder.ClearReaders()

		testUrlString := "https://www.google.com"
		parsedUrl, err := url.Parse(testUrlString)
		if err != nil {
			t.Error(err)
		}

		// Select URL reader
		builder.URL(parsedUrl)

		_, err = builder.Read()
		if err != nil {
			t.Error(err)
		}

		// If the URL is valid, the response should be a 200. If we get here, we know the URL is valid.
	})

	t.Run("TestFile", func(t *testing.T) {
		builder.ClearReaders()

		testFile := "./testFile.txt"

		// Create a test file
		file, err := os.Create(testFile)
		if err != nil {
			t.Error(err)
		}

		// Write to the test file
		testString := "Hello, World!"
		_, err = file.WriteString(testString)
		if err != nil {
			t.Error(err)
		}

		// Select Path reader
		builder.Path(testFile)

		// Read from the test file
		got, err := builder.Read()
		if err != nil {
			t.Error(err)
		}

		// Compare the read string to the test string
		if !reflect.DeepEqual(got, []byte(testString)) {
			t.Errorf("got %v, want %v", got, testString)
		}

		// Delete the test file
		err = os.Remove(testFile)
		if err != nil {
			t.Error(err)
		}
	})

	// Framework to test where reader fails

	tests := []struct {
		name          string
		setup         func()
		errorExpected bool
		want          []byte
	}{
		{
			// Test Stdin reader.
			name: "TestFileNotFound",
			setup: func() {
				builder.Path("testFile.txt")
			},
			errorExpected: true,
			want:          nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Clear the readers from the builder
			builder.ClearReaders()

			// Set up the test
			test.setup()

			// Read from the builder
			got, err := builder.Read()
			if (err != nil) != test.errorExpected {
				t.Errorf("error = %v, errorExpected %v", err, test.errorExpected)
			}

			// Compare the result
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
