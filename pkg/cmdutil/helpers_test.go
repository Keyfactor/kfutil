package cmdutil

import (
	"fmt"
	"testing"
)

func TestPrintError(t *testing.T) {
	err := fmt.Errorf("test error")
	PrintError(err)
}
