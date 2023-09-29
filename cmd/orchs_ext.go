package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kfutil/pkg/cmdutil"
	"kfutil/pkg/cmdutil/extensions"
	"kfutil/pkg/cmdutil/flags"
	"log"
)

const defaultExtensionOutDir = "./extensions"

// Ensure that OrchsExtFlags implements Flags
var _ flags.Flags = &OrchsExtFlags{}
var _ flags.Options = &OrchsExtOptions{}

type OrchsExtFlags struct {
	// ExtensionConfigFilename is the filename, directory, or URL to an extension configuration file to use for the extension
	// An extension configuration file is a YAML file that contains a map of extension names to versions to download
	// To download the latest version of an extension, use the value "latest"
	ExtensionConfigFilename *flags.FilenameFlags
	// Extensions is a list of extensions to download. Should be in the format <extension name>@<version>.
	// If no version is specified, the latest version will be downloaded.
	// Extensions is mutually exclusive with ExtensionConfigFilename
	Extensions *[]string
	// GithubToken is the token used for related authentication - required for private repositories
	GithubToken *string
	// OutDir is the path to the extensions directory to download extensions into. Default is ./extensions
	OutDir *string
	// AutoConfirm configures the command to not prompt for confirmation before downloading extensions
	AutoConfirm *bool
}

func NewOrchsExtFlags() *OrchsExtFlags {
	filenameFlagName := "config"
	filenameFlagShorthand := "c"
	filenameUsage := "Filename, directory, or URL to an extension configuration file to use for the extension"
	var filenames []string

	githubToken := ""
	outPath := ""
	var extensions []string
	var autoConfirm bool

	return &OrchsExtFlags{
		ExtensionConfigFilename: flags.NewFilenameFlags(filenameFlagName, filenameFlagShorthand, filenameUsage, filenames),
		Extensions:              &extensions,
		GithubToken:             &githubToken,
		OutDir:                  &outPath,
		AutoConfirm:             &autoConfirm,
	}
}

func (f *OrchsExtFlags) AddFlags(flags *pflag.FlagSet) {
	// Implement Flags interface

	// Add Filename flags
	f.ExtensionConfigFilename.AddFlags(flags)

	// Add custom flags
	flags.StringVarP(f.GithubToken, "token", "t", *f.GithubToken, "Token used for related authentication - required for private repositories")
	flags.StringVarP(f.OutDir, "out", "o", *f.OutDir, "Path to the extensions directory to download extensions into. Default is ./extensions")
	flags.StringSliceVarP(f.Extensions, "extensions", "e", *f.Extensions, "List of extensions to download. Should be in the format <extension name>:<version>. If no version is specified, the latest official version will be downloaded.")
	flags.BoolVarP(f.AutoConfirm, "confirm", "y", *f.AutoConfirm, "Automatically confirm the download of extensions")
}

func NewCmdOrchsExt() *cobra.Command {
	orchsExtFlags := NewOrchsExtFlags()

	cmd := &cobra.Command{
		Use:   "ext",
		Short: "Keyfactor Command Universal Orchestrator utility for downloading and configuring extensions",
		Long:  CmdOrchsExtLongDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := orchsExtFlags.ToOptions(cmd, args)
			if err != nil {
				return fmt.Errorf("failed to get runtime options from flags: %s", err)
			}

			installer := extensions.NewExtensionInstallerBuilder().
				ExtensionDir(options.OutPath).
				InteractiveMode(options.InteractiveMode).
				Token(options.GithubToken).
				// Extensions is a slice of strings in the format <extension name>@<version>
				Extensions(options.Extensions).
				// ExtensionConfigFilename is the filename, directory, or URL to an extension configuration file to use for the extension
				ExtensionConfig(options.ExtensionConfigOptions).
				// AutoConfirm configures the command to not prompt for confirmation before downloading extensions
				AutoConfirm(options.AutoConfirm)

			if err = installer.PreFlight(); err != nil {
				return fmt.Errorf("extension installer preflight failed: %s", err)
			}

			err = installer.Run()
			if err != nil {
				cmdutil.PrintError(err)
				log.Fatalf("[ERROR] Exiting: %s", err)
			}

			return nil
		},
	}

	// Add flags
	orchsExtFlags.AddFlags(cmd.Flags())

	return cmd
}

type OrchsExtOptions struct {
	// Runtime options
	GithubToken            string
	OutPath                string
	ExtensionConfigOptions flags.FilenameOptions
	Extensions             []string
	AutoConfirm            bool

	// Interpreted options
	InteractiveMode bool
}

func (f *OrchsExtFlags) ToOptions(cmd *cobra.Command, args []string) (*OrchsExtOptions, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}

	options := &OrchsExtOptions{}

	// Global flags
	flags.GetDebugFlag(cmd)

	// Get the values from the flags
	if f.ExtensionConfigFilename != nil {
		filenameOptions := f.ExtensionConfigFilename.ToOptions()

		if err := filenameOptions.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate filename options: %s", err)
		}

		options.ExtensionConfigOptions = filenameOptions
	}

	if f.GithubToken != nil {
		options.GithubToken = *f.GithubToken
	}

	if f.OutDir != nil {
		options.OutPath = *f.OutDir
	}

	if f.Extensions != nil {
		options.Extensions = *f.Extensions
	}

	if f.AutoConfirm != nil {
		options.AutoConfirm = *f.AutoConfirm
	}

	// Set the default out path if it is empty
	if options.OutPath == "" {
		options.OutPath = defaultExtensionOutDir
	}

	return options, options.Validate()
}

// Validate checks that at least one filename was provided
func (f *OrchsExtOptions) Validate() error {
	// Check that either ExtensionConfigFilename or Extensions was provided (xor)
	if !f.ExtensionConfigOptions.IsEmpty() && len(f.Extensions) != 0 {
		return fmt.Errorf("only one of --config or --extensions can be provided")
	}

	// If neither ExtensionConfigFilename or Extensions was provided, set InteractiveMode
	if f.ExtensionConfigOptions.IsEmpty() && len(f.Extensions) == 0 {
		f.InteractiveMode = true
	}

	return nil
}

const CmdOrchsExtLongDescription = `
Keyfactor Command Universal Orchestrator utility for downloading and configuring extensions.

This command will download extensions for Keyfactor Command Universal Orchestrator. Extensions can be downloaded from a configuration file or by specifying the extension name and version.
`
