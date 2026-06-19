// Copyright 2025 Keyfactor
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"kfutil/pkg/upgrade"
	"kfutil/pkg/version"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade kfutil to the latest release",
	Long: `Fetches a kfutil release from GitHub, verifies its SHA-256 checksum,
and atomically replaces the running binary.

By default the latest published release is installed. Pass --version with any
valid GitHub tag (e.g. v1.9.0, v1.10.0-beta.1) to install a specific release,
including pre-releases and older versions.

Examples:
  kfutil upgrade                      # install latest
  kfutil upgrade --version v1.8.0    # install a specific tag
  kfutil upgrade --dry-run           # preview without changing anything`,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetVersion, _ := cmd.Flags().GetString("version")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")
		return upgrade.Run(version.VERSION, targetVersion, dryRun, force)
	},
}

func init() {
	upgradeCmd.Flags().String("version", "", "GitHub tag to install (default: latest release)")
	upgradeCmd.Flags().Bool("dry-run", false, "Show what would be downloaded without replacing the binary")
	upgradeCmd.Flags().Bool("force", false, "Upgrade even if already at the target version")
	RootCmd.AddCommand(upgradeCmd)
}
