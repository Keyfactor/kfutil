package cmdutil

import (
	"fmt"
)

func PrintError(err error) {
	// Print errors in red
	if err != nil {
		fmt.Printf("\033[31m%s\u001B[0m\n", err)
	}
}
