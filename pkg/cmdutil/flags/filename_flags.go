package flags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kfutil/pkg/cmdutil"
	"net/url"
	"strings"
)

// Ensure that FilenameFlags implements Flags
var _ Flags = &FilenameFlags{}

var SupportedExtensions = []string{".json", ".yaml", ".yml"}

type FilenameFlags struct {
	Name  string
	Usage string

	Filenames *[]string
}

func NewFilenameFlags(name string, usage string, filenames []string) *FilenameFlags {
	return &FilenameFlags{
		Name:      name,
		Usage:     usage,
		Filenames: &filenames,
	}
}

func (f *FilenameFlags) AddFlags(flags *pflag.FlagSet) {
	if f.Filenames != nil {
		flags.StringSliceVarP(f.Filenames, f.Name, "f", *f.Filenames, f.Usage)
		annotations := make([]string, 0, len(SupportedExtensions))
		for _, ext := range SupportedExtensions {
			annotations = append(annotations, strings.TrimLeft(ext, "."))
		}
		err := flags.SetAnnotation("filename", cobra.BashCompFilenameExt, annotations)
		if err != nil {
			return
		}
	}
}

type FilenameOptions struct {
	Filenames []string
}

func (f *FilenameFlags) ToOptions() FilenameOptions {
	options := FilenameOptions{}

	if f == nil {
		return options
	}

	if f.Filenames != nil {
		options.Filenames = *f.Filenames
	}

	return options
}

// Validate checks that at least one filename was provided
func (f *FilenameOptions) Validate() error {
	if f.IsEmpty() {
		return nil
	}

	for _, filename := range f.Filenames {
		if filename == "" {
			return fmt.Errorf("filename cannot be empty")
		}
	}

	return nil
}

func (f *FilenameOptions) IsEmpty() bool {
	return len(f.Filenames) == 0
}

func (f *FilenameOptions) Merge(other *FilenameOptions) {
	if other == nil {
		return
	}

	if other.Filenames != nil {
		f.Filenames = append(f.Filenames, other.Filenames...)
	}
}

func (f *FilenameOptions) Read() ([]byte, error) {
	if f.IsEmpty() {
		return nil, nil
	}

	b := cmdutil.NewReaderBuilder()

	paths := f.Filenames
	for _, s := range paths {
		switch {
		case s == "-":
			b.Stdin()
		case strings.Index(s, "http://") == 0 || strings.Index(s, "https://") == 0:
			fileUrl, err := url.Parse(s)
			if err != nil {
				return nil, fmt.Errorf("the URL passed to filename %q is not valid: %v", s, err)
			}
			b.URL(fileUrl)
		default:
			b.Path(paths...)
		}
	}

	return b.Read()
}
