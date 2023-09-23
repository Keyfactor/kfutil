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
		break
	case debugFlag:
		log.SetOutput(os.Stdout)
		debugModeEnabled = true
		break
	default:
		log.SetOutput(io.Discard)
		debugModeEnabled = false
		break
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
