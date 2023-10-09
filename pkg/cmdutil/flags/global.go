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

package flags

import (
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"strconv"
)

func GetDebugFlag(cmd *cobra.Command) bool {
	debugModeEnabled := false

	// Get the debug flag from the global flags
	debugFlag := GetFlagBool(cmd, "debug")

	// Get the debug flag from the environment variable
	envDebug := os.Getenv("KFUTIL_DEBUG")
	envValue, _ := strconv.ParseBool(envDebug)

	switch {
	case (envValue && !debugFlag) || (envValue && debugFlag):
		log.SetOutput(os.Stdout)
		debugModeEnabled = true
	case debugFlag:
		log.SetOutput(os.Stdout)
		debugModeEnabled = true
	default:
		log.SetOutput(io.Discard)
		debugModeEnabled = false
	}

	log.Println("Debug mode enabled: ", debugModeEnabled)
	return debugModeEnabled
}

func GetNoPromptFlag(cmd *cobra.Command) bool {
	return GetFlagBool(cmd, "no-prompt")
}

func GetProfileFlag(cmd *cobra.Command) string {
	profile := GetFlagString(cmd, "profile")
	if profile == "" {
		profile = "default"
	}
	return profile
}

func GetConfigFlag(cmd *cobra.Command) string {
	return GetFlagString(cmd, "config")
}
