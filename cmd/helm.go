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
