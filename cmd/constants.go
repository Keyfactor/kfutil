// Package cmd Copyright 2024 Keyfactor
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
	ColorRed                              = "\033[31m"
	ColorWhite                            = "\033[37m"
	DefaultAPIPath                        = "KeyfactorAPI"
	DefaultConfigFileName                 = "command_config.json"
	DefaultROTAuditStoresOutfilePath      = "rot_audit_selected_stores.csv"
	DefaultROTAuditAddCertsOutfilePath    = "rot_audit_selected_certs_add.csv"
	DefaultROTAuditRemoveCertsOutfilePath = "rot_audit_selected_certs_remove.csv"
	FailedAuthMsg                         = "Login failed!"
	SuccessfulAuthMsg                     = "Login successful!"
	XKeyfactorRequestedWith               = "APIClient"
	XKeyfactorApiVersion                  = "1"
	FlagGitRef                            = "git-ref"
	FlagFromFile                          = "from-file"
	DebugFuncEnter                        = "entered: %s"
	DebugFuncExit                         = "exiting: %s"
	DebugFuncCall                         = "calling: %s"
	ErrMsgEmptyResponse                   = "empty response received from Keyfactor Command %s"
)

// CLI Menu Defaults
const (
	DefaultMenuPageSizeSmall = 25
	DefaultMenuPageSizeLarge = 100
)

var (
	DefaultSourceTypeOptions = []string{"API", "File"}
)

var ProviderTypeChoices = []string{
	"azid",
}
var ValidAuthProviders = [2]string{"azure-id", "azid"}
var ErrKfcEmptyResponse = fmt.Errorf("empty response recieved from Keyfactor Command")

// Error messages
var (
	StoreTypeReadError      = fmt.Errorf("error reading store type from configuration file")
	InvalidInputError       = fmt.Errorf("invalid input")
	InvalidROTCertsInputErr = fmt.Errorf(
		"at least one of `--add-certs` or `--remove-certs` is required to perform a" +
			" root of trust audit",
	)
)
