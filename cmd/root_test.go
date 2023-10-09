// Package cmd Copyright 2023 Keyfactor
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
	"testing"
)

func Test_RootCmd(t *testing.T) {
	// Test root help
	testCmd := RootCmd
	testCmd.SetArgs([]string{"--help"})
	testCmd.Execute()

	// test root halp
	testCmd.SetArgs([]string{"-h"})
	testCmd.Execute()

	// test root halp
	testCmd.SetArgs([]string{"--halp"})
	//testCmd.Execute()
	// check if error was returned
	if err := testCmd.Execute(); err == nil {
		t.Errorf("RootCmd() = %v, shouldNotPass %v", err, true)
	}
}
