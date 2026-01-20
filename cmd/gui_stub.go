//go:build !gui

// Copyright 2026 Keyfactor
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
	"fmt"

	"github.com/spf13/cobra"
)

// guiCmd represents the gui command (stub for CLI-only builds)
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface (not available in this build)",
	Long: `The GUI is not available in this build of kfutil.

To use the graphical user interface, you need to install the GUI-enabled version:
- Download 'kfutil-gui' from the releases page
- Or build from source with: go build -tags gui

The GUI provides a visual interface for:
- Configuring authentication to Keyfactor Command
- Viewing and managing installed store types
- Browsing and deploying store types from the internal catalog
- Importing and exporting store type configurations
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		return fmt.Errorf("GUI is not available in this build. Install 'kfutil-gui' or build with '-tags gui'")
	},
}

func init() {
	RootCmd.AddCommand(guiCmd)
}
