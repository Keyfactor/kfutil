/*
Copyright 2024 The Keyfactor Command Authors.

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
	"log"

	"github.com/Keyfactor/keyfactor-auth-client-go/auth_providers"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kfutil/pkg/cmdutil"
	"kfutil/pkg/cmdutil/flags"
	"kfutil/pkg/helm"
)

// DefaultValuesLocation TODO when Helm is ready, set this to the default values.yaml location in Git
const DefaultValuesLocation = ""

// Ensure that HelmUoFlags implements Flags
var _ flags.Flags = &HelmUoFlags{}
var _ flags.Options = &HelmUoOptions{}

type HelmUoFlags struct {
	FilenameFlags *flags.FilenameFlags
	GithubToken   *string
	OutPath       *string
	Extensions    *[]string
}

func NewHelmUoFlags() *HelmUoFlags {
	filenameFlagName := "values"
	filenameFlagShorthand := "f"
	filenameUsage := "Filename, directory, or URL to a default values.yaml file to use for the chart"
	var filenames []string

	// General configuration
	githubToken := ""
	outPath := ""

	// Non-interactive configuration
	var extensionsFlag []string

	return &HelmUoFlags{
		FilenameFlags: flags.NewFilenameFlags(filenameFlagName, filenameFlagShorthand, filenameUsage, filenames),
		GithubToken:   &githubToken,
		OutPath:       &outPath,
		Extensions:    &extensionsFlag,
	}
}

func (f *HelmUoFlags) AddFlags(flags *pflag.FlagSet) {
	// Implement Flags interface

	// Add Filename flags
	f.FilenameFlags.AddFlags(flags)

	// Add custom flags
	flags.StringVarP(
		f.GithubToken,
		"token",
		"t",
		*f.GithubToken,
		"Token used for related authentication - required for private repositories",
	)
	flags.StringVarP(
		f.OutPath,
		"out",
		"o",
		*f.OutPath,
		"Path to output the modified values.yaml file. This file can then be used with helm install -f <file> to override the default values.",
	)
	flags.StringSliceVarP(
		f.Extensions,
		"extension",
		"e",
		*f.Extensions,
		"List of extensions to install. Should be in the format <extension name>@<version>. If no version is specified, the latest version will be downloaded.",
	)
}

func NewCmdHelmUo() *cobra.Command {
	helmUoFlags := NewHelmUoFlags()

	cmd := &cobra.Command{
		Use:   HelmUoUse,
		Short: HelmUoShortDescription,
		Long:  HelmUoLongDescription,
		RunE: func(cmd *cobra.Command, args []string) error {
			options, err := helmUoFlags.ToOptions(cmd, args)
			if err != nil {
				cmdutil.PrintError(err)
				return err
			}

			// Build the tool
			builder := helm.NewUniversalOrchestratorHelmValueBuilder().
				Extensions(options.Extensions).
				CommandHostname(options.CommandHostname).
				OverrideFile(options.OutPath).
				Token(options.GithubToken).
				Values(options.FilenameOptions).
				InteractiveMode(options.InteractiveMode).
				Writer(cmd.OutOrStdout())

			// Pre flight
			err = builder.PreFlight()
			if err != nil {
				return err
			}

			// Run the tool
			newValues, err := builder.Build()
			if err != nil {
				cmdutil.PrintError(err)
				return err
			}

			// Write the new values to stdout
			_, err = cmd.OutOrStdout().Write([]byte(newValues))
			if err != nil {
				return err
			}

			return nil
		},
	}

	helmUoFlags.AddFlags(cmd.Flags())

	return cmd
}

type HelmUoOptions struct {
	GithubToken     string
	OutPath         string
	CommandHostname string
	FilenameOptions flags.FilenameOptions
	Extensions      []string
	InteractiveMode bool
}

func (f *HelmUoFlags) ToOptions(cmd *cobra.Command, args []string) (*HelmUoOptions, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("unexpected args: %v", args)
	}

	options := &HelmUoOptions{}

	// Global flags
	flags.GetDebugFlag(cmd)

	// Determine if feature is enabled
	flagEnableExp, _ = cmd.Flags().GetBool("exp")
	_, err := isExperimentalFeatureEnabled(flagEnableExp, true)
	if err != nil {
		return nil, fmt.Errorf("feature gate check failed: %s", err)
	}

	// Get the command config entry from global flags
	commandConfig, _ := auth_providers.ReadConfigFromJSON(flagConfigFile)

	// Get the hostname from the command config
	entry, ok := commandConfig.Servers[flagProfile]
	if ok {
		if entry.Host != "" {
			options.CommandHostname = commandConfig.Servers[flagProfile].Host
		}
	}

	// Get the filename options
	if f.FilenameFlags != nil {
		filenameOptions := f.FilenameFlags.ToOptions()

		if filenameOptions.IsEmpty() {
			// If no filenames were provided, use the default values.yaml location
			log.Printf("[DEBUG] No filenames provided, using default values.yaml location: %q", DefaultValuesLocation)
			filenameOptions.Merge(&flags.FilenameOptions{Filenames: []string{DefaultValuesLocation}})
		}

		if err := filenameOptions.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate filename options: %s", err)
		}

		options.FilenameOptions = filenameOptions
	}

	// Get the custom flags
	if f.GithubToken != nil {
		options.GithubToken = *f.GithubToken
	}
	if f.OutPath != nil {
		options.OutPath = *f.OutPath
	}
	if f.Extensions != nil {
		options.Extensions = *f.Extensions
	}

	return options, options.Validate()
}

func (f *HelmUoOptions) Validate() error {
	// If Extensions is empty, set InteractiveMode
	if len(f.Extensions) == 0 {
		f.InteractiveMode = true
	}
	return nil
}

const (
	HelmUoShortDescription = "Configure the Keyfactor Universal Orchestrator Helm Chart"
	HelmUoLongDescription  = `Configure the Keyfactor Universal Orchestrator Helm Chart by prompting the user for configuration values and outputting a YAML file that can be used with the Helm CLI to install the chart.

Also supported is the ability specify extensions and skip the interactive prompts.
`
	HelmUoUse = `uo [-t <token>] [-o <path>] [-f <file, url, or '-'>] [-e <extension name>@<version>]...`
)
