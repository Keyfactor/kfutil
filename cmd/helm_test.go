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
	"kfutil/pkg/cmdtest"
	"strings"
	"testing"
)

func TestHelm(t *testing.T) {
	t.Run("Test helm command", func(t *testing.T) {
		// The helm command doesn't have any flags or a RunE function, so the output should be the same as the help menu.
		cmd := NewCmdHelm()

		t.Logf("Testing %q", cmd.Use)

		helmNoFlag, err := cmdtest.TestExecuteCommand(t, cmd, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		helmHelp, err := cmdtest.TestExecuteCommand(t, cmd, "-h")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		diff := strings.Compare(string(helmNoFlag), string(helmHelp))
		if diff != 0 {
			t.Errorf("Expected helmNoFlag to equal helmHelp, but got: %v", diff)
		}
	})
}
