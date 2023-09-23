package flags

import (
	"github.com/spf13/cobra"
	"log"
)

func GetFlagBool(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	if err != nil {
		log.Fatalf("error accessing flag %s for command %s: %v", flag, cmd.Name(), err)
	}
	return b
}

func GetFlagString(cmd *cobra.Command, flag string) string {
	s, err := cmd.Flags().GetString(flag)
	if err != nil {
		log.Fatalf("error accessing flag %s for command %s: %v", flag, cmd.Name(), err)
	}
	return s
}
