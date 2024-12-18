// Copyright 2024 Keyfactor
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

import "fmt"

const (
	ColorRed                  = "\033[31m"
	ColorWhite                = "\033[37m"
	DefaultAPIPath            = "KeyfactorAPI"
	DefaultConfigFileName     = "command_config.json"
	DefaultStoreTypesFileName = "store_types.json"
	DefaultGitRepo            = "kfutil"
	DefaultGitRef             = "main"
	FailedAuthMsg             = "Login failed!"
	SuccessfulAuthMsg         = "Login successful!"
	XKeyfactorRequestedWith   = "APIClient"
	XKeyfactorApiVersion      = "1"
	FlagGitRef                = "git-ref"
	FlagGitRepo               = "repo"
	FlagFromFile              = "from-file"
	DebugFuncEnter            = "entered: %s"
	DebugFuncExit             = "exiting: %s"
	DebugFuncCall             = "calling: %s"
	MinHttpTimeout            = 3
)

var ProviderTypeChoices = []string{
	"azid",
}
var ValidAuthProviders = [2]string{"azure-id", "azid"}

// Error messages
var (
	StoreTypeReadError = fmt.Errorf("error reading store type from configuration file")
	InvalidInputError  = fmt.Errorf("invalid input")
)
