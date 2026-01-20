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
	"kfutil/pkg/gui"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// guiCmd represents the gui command
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Long: `Launch the kfutil graphical user interface (GUI) for managing
certificate store types.

The GUI provides a visual interface for:
- Configuring authentication to Keyfactor Command
- Viewing and managing installed store types
- Browsing and deploying store types from the internal catalog
- Importing and exporting store type configurations

Example:
  kfutil gui
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		log.Debug().Msg("Launching GUI...")

		isExperimental := true

		informDebug(debugFlag)
		debugErr := warnExperimentalFeature(expEnabled, isExperimental)
		if debugErr != nil {
			return debugErr
		}

		err := gui.LaunchApp()
		if err != nil {
			log.Error().Err(err).Msg("GUI exited with error")
			return err
		}

		log.Debug().Msg("GUI closed normally")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(guiCmd)
}
