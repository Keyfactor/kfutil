/*
Copyright 2023 The Keyfactor Command Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kfutil/pkg/cmdutil/extensions"
	"kfutil/pkg/cmdutil/flags"
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
	// GithubOrg is the Github organization to download extensions from. Default is keyfactor.
	GithubOrg *string
	// OutDir is the path to the extensions directory to download extensions into. Default is ./extensions
	OutDir *string
	// AutoConfirm configures the command to not prompt for confirmation before downloading extensions
	AutoConfirm *bool
	// Upgrade looks in the extensions directory for existing extensions and upgrades them if they are out of date
	Upgrade *bool
	// Prune removes extensions from the extensions directory that are not in the extension configuration file or specified on the command line
	Prune *bool
}

func NewOrchsExtFlags() *OrchsExtFlags {
	filenameFlagName := "config"
	filenameFlagShorthand := "c"
	filenameUsage := "Filename, directory, or URL to an extension configuration file to use for the extension"
	var filenames []string

	githubToken := ""
	githubOrg := ""
	outPath := ""
	var extensionsFlag []string
	var autoConfirm bool
	var upgrade bool
	var prune bool

	return &OrchsExtFlags{
		ExtensionConfigFilename: flags.NewFilenameFlags(filenameFlagName, filenameFlagShorthand, filenameUsage, filenames),
		Extensions:              &extensionsFlag,
		GithubToken:             &githubToken,
		GithubOrg:               &githubOrg,
		OutDir:                  &outPath,
		AutoConfirm:             &autoConfirm,
		Upgrade:                 &upgrade,
		Prune:                   &prune,
	}
}

func (f *OrchsExtFlags) AddFlags(flags *pflag.FlagSet) {
	// Implement Flags interface

	// Add Filename flags
	f.ExtensionConfigFilename.AddFlags(flags)

	// Add custom flags
	flags.StringVarP(f.GithubToken, "token", "t", *f.GithubToken, "Token used for related authentication - required for private repositories")
	flags.StringVarP(f.GithubOrg, "org", "", *f.GithubOrg, "Github organization to download extensions from. Default is keyfactor.")
	flags.StringVarP(f.OutDir, "out", "o", *f.OutDir, "Path to the extensions directory to download extensions into. Default is ./extensions")
	flags.StringSliceVarP(f.Extensions, "extension", "e", *f.Extensions, "List of extensions to download. Should be in the format <extension name>@<version>. If no version is specified, the latest official version will be downloaded.")
	flags.BoolVarP(f.AutoConfirm, "confirm", "y", *f.AutoConfirm, "Automatically confirm the download of extensions")
	flags.BoolVarP(f.Upgrade, "update", "u", *f.Upgrade, "Update existing extensions if they are out of date.")
	flags.BoolVarP(f.Prune, "prune", "P", *f.Prune, "Remove extensions from the extensions directory that are not in the extension configuration file or specified on the command line")
}

func NewCmdOrchsExt() *cobra.Command {
	orchsExtFlags := NewOrchsExtFlags()

	cmd := &cobra.Command{
		Use:     OrchsExtUsage,
		Aliases: OrchsExtAliases,
		Short:   OrchsExtShortDescription,
		Long:    OrchsExtLongDescription,
		Example: OrchsExtExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := orchsExtFlags.ToOptions(cmd, args)
			if err != nil {
				return fmt.Errorf("failed to get runtime options from flags: %s", err)
			}

			installer := extensions.NewExtensionInstallerBuilder().
				ExtensionDir(options.OutPath).
				InteractiveMode(options.InteractiveMode).
				Token(options.GithubToken).
				Org(options.GithubOrg).
				// Extensions is a slice of strings in the format <extension name>@<version>
				Extensions(options.Extensions).
				// ExtensionConfigFilename is the filename, directory, or URL to an extension configuration file to
				// use for the extension
				ExtensionConfig(options.ExtensionConfigOptions).
				// AutoConfirm configures the command to not prompt for confirmation before downloading extensions
				AutoConfirm(options.AutoConfirm).
				Writer(cmd.OutOrStdout())

			if options.Upgrade {
				installer.Upgrade()
			}

			if options.Prune {
				installer.Prune()
			}

			if err = installer.PreFlight(); err != nil {
				return fmt.Errorf("extension installer preflight failed: %s", err)
			}

			err = installer.Run()
			if err != nil {
				_, err = cmd.OutOrStderr().Write([]byte(err.Error()))
				if err != nil {
					return err
				}
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
	GithubOrg              string
	OutPath                string
	ExtensionConfigOptions flags.FilenameOptions
	Extensions             []string
	AutoConfirm            bool
	Upgrade                bool
	Prune                  bool

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

	if f.GithubOrg != nil {
		options.GithubOrg = *f.GithubOrg
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

	if f.Upgrade != nil {
		options.Upgrade = *f.Upgrade
	}

	if f.Prune != nil {
		options.Prune = *f.Prune
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

const (
	OrchsExtShortDescription = `Download and configure extensions for Keyfactor Command Universal Orchestrator`
	OrchsExtLongDescription  = `
Keyfactor Command Universal Orchestrator utility for downloading and configuring extensions.

This command will download extensions for Keyfactor Command Universal Orchestrator. Extensions can be downloaded from a configuration file or by specifying the extension name and version.
`
	OrchsExtUsage   = `ext [-t <token>] [--org <Github org>] [-o <out path>] [-c <config file> | -e <extension name>@<version>] [-y] [-u] [-P]`
	OrchsExtExample = `ext -t <token> -e <extension>@<version>,<extension>@<version> -o ./app/extensions --confirm, --update, --prune`
)

var (
	OrchsExtAliases = []string{"extensions"}
)
