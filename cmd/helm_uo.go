package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kfutil/pkg/cmdutil"
	"kfutil/pkg/cmdutil/flags"
	"kfutil/pkg/helm"
	"log"
)

// DefaultValuesLocation TODO when Helm is ready, set this to the default values.yaml location in Git
const DefaultValuesLocation = ""

// Ensure that HelmUoFlags implements Flags
var _ flags.Flags = &HelmUoFlags{}

type HelmUoFlags struct {
	FilenameFlags *flags.FilenameFlags
	GithubToken   *string
	OutPath       *string
}

func NewHelmUoFlags() *HelmUoFlags {
	filenameFlagName := "values"
	filenameUsage := "Filename, directory, or URL to a default values.yaml file to use for the chart"
	var filenames []string

	githubToken := ""
	outPath := ""

	return &HelmUoFlags{
		FilenameFlags: flags.NewFilenameFlags(filenameFlagName, filenameUsage, filenames),
		GithubToken:   &githubToken,
		OutPath:       &outPath,
	}
}

func (f *HelmUoFlags) AddFlags(flags *pflag.FlagSet) {
	// Implement Flags interface

	// Add Filename flags
	f.FilenameFlags.AddFlags(flags)

	// Add custom flags
	flags.StringVarP(f.GithubToken, "token", "t", *f.GithubToken, "Token used for related authentication - required for private repositories")
	flags.StringVarP(f.OutPath, "out", "o", *f.OutPath, "Path to output the modified values.yaml file. This file can then be used with helm install -f <file> to override the default values.")
}

func NewCmdHelmUo() *cobra.Command {
	helmUoFlags := NewHelmUoFlags()

	cmd := &cobra.Command{
		Use:   "uo",
		Short: "Keyfactor Helm Chart Utilities for the Containerized Universal Orchestrator",
		Long:  `Keyfactor Helm Chart Utilities used to configure charts and assist in the deployment of the Keyfactor Command Universal Orchestrator.`,
		Run: func(cmd *cobra.Command, args []string) {
			options, err := helmUoFlags.ToOptions(cmd, args)
			if err != nil {
				cmdutil.PrintError(err)
				log.Fatalf("[ERROR] Exiting: %s", err)
			}

			err = options.Run()
			if err != nil {
				cmdutil.PrintError(err)
				log.Fatalf("[ERROR] Exiting: %s", err)
			}
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
}

func (f *HelmUoFlags) ToOptions(cmd *cobra.Command, args []string) (*HelmUoOptions, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("unexpected args: %v", args)
	}

	options := &HelmUoOptions{}

	// Global flags
	flags.GetDebugFlag(cmd)

	// Get the command config entry from global flags
	noPrompt := flags.GetNoPromptFlag(cmd)
	profile := flags.GetProfileFlag(cmd)
	configFile := flags.GetConfigFlag(cmd)
	commandConfig, _ := authConfigFile(configFile, profile, noPrompt, false)

	// Get the hostname from the command config
	entry, ok := commandConfig.Servers[profile]
	if ok {
		if entry.Hostname != "" {
			options.CommandHostname = commandConfig.Servers[profile].Hostname
		}
	}

	// Get the filename options
	if f.FilenameFlags != nil {
		filenameOptions := f.FilenameFlags.ToOptions()

		if filenameOptions.IsEmpty() {
			// If no filenames were provided, use the default values.yaml location
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

	return options, nil
}

func (o *HelmUoOptions) Run() error {
	tool := helm.NewToolBuilder().
		// Set up the builder
		CommandHostname(o.CommandHostname).
		OverrideFile(o.OutPath).
		Token(o.GithubToken).
		Values(o.FilenameOptions).
		// Pre flight
		PreFlight().
		// Run the interactive tool
		BuildUniversalOrchestratorHelmValueTool()

	// Run the interactive tool
	return tool()
}
