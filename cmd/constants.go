package cmd

import "fmt"

const (
	ColorRed                = "\033[31m"
	ColorWhite              = "\033[37m"
	DefaultAPIPath          = "KeyfactorAPI"
	DefaultConfigFileName   = "command_config.json"
	FailedAuthMsg           = "Login failed!"
	SuccessfulAuthMsg       = "Login successful!"
	XKeyfactorRequestedWith = "APIClient"
	XKeyfactorApiVersion    = "1"
	FlagGitRef              = "git-ref"
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
