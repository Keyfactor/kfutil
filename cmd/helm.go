package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"kfutil/pkg/cmdutil/flags"
)

// Ensure that HelmFlags implements Flags
var _ flags.Flags = &HelmFlags{}

type HelmFlags struct {
}

func NewHelmFlags() *HelmFlags {
	return &HelmFlags{}
}

func (f *HelmFlags) AddFlags(flags *pflag.FlagSet) {
	// Implement Flags interface
}

func NewCmdHelm() *cobra.Command {
	helmFlags := NewHelmFlags()

	cmd := &cobra.Command{
		Use:     "helm",
		Short:   helmShortDescription,
		Long:    helmLongDescription,
		Example: "kubectl helm uo | helm install -f - keyfactor-universal-orchestrator keyfactor/keyfactor-universal-orchestrator",
	}

	helmFlags.AddFlags(cmd.Flags())

	// Add subcommands
	cmd.AddCommand(NewCmdHelmUo())

	return cmd
}

// Example usage:
// kubectl helm uo | helm install -f - keyfactor-universal-orchestrator keyfactor/keyfactor-universal-orchestrator

func init() {
	// Helm Command
	helmCmd := NewCmdHelm()
	RootCmd.AddCommand(helmCmd)
}

const (
	helmShortDescription = `Helm utilities for configuring Keyfactor Helm charts`
	helmLongDescription  = helmShortDescription
)
