package helm

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v3"
	"kfutil/pkg/cmdutil"
	"kfutil/pkg/cmdutil/extensions"
	"kfutil/pkg/cmdutil/flags"
	"log"
	"os"
	"reflect"
	"strconv"
)

const (
	installerImage           = "m8rmclarenkf/uo_extension_installer:1.0.5"
	installerImagePullPolicy = "IfNotPresent"
)

type InteractiveUOValueBuilder struct {
	errs            []error
	commandHostname string
	overrideFile    string
	token           string
	extensions      extensions.Extensions
	newValues       UniversalOrchestratorHelmValues
	newValuesString string
	exitAfterPrompt bool
	interactiveMode bool
}

type menuOption struct {
	optionName   string
	optionDesc   string
	currentValue any
	handlerFunc  func() error
}

func NewUniversalOrchestratorHelmValueBuilder() *InteractiveUOValueBuilder {
	return &InteractiveUOValueBuilder{}
}

func (b *InteractiveUOValueBuilder) CommandHostname(hostname string) *InteractiveUOValueBuilder {
	b.commandHostname = hostname

	return b
}

func (b *InteractiveUOValueBuilder) InteractiveMode(interactiveMode bool) *InteractiveUOValueBuilder {
	b.interactiveMode = interactiveMode

	return b
}

func (b *InteractiveUOValueBuilder) Extensions(extensionsString []string) *InteractiveUOValueBuilder {
	if b.extensions == nil {
		b.extensions = make(extensions.Extensions)
	}

	// Serialize extensionsString into a map of extensions to install and the version
	for _, extensionString := range extensionsString {
		extension, err := extensions.ParseExtensionString(extensionString)
		if err != nil {
			cmdutil.PrintError(err)
			b.errs = append(b.errs, err)
		}

		b.extensions[extension.Name] = extension.Version
	}

	// Create an extension installer tool
	installerTool := extensions.NewExtensionInstallerBuilder().Token(b.token)
	installerTool.SetExtensionsToInstall(b.extensions)

	// Lint the extensions
	err := installerTool.ValidateExtensionsToInstall()
	if err != nil {
		b.errs = append(b.errs, err)
	}

	// installerTool now contains the correct list of extensions to install
	b.extensions = installerTool.GetExtensionsToInstall()

	return b
}

func (b *InteractiveUOValueBuilder) OverrideFile(filename string) *InteractiveUOValueBuilder {
	b.overrideFile = filename

	return b
}

func (b *InteractiveUOValueBuilder) Token(token string) *InteractiveUOValueBuilder {
	b.token = token

	return b
}

func (b *InteractiveUOValueBuilder) Values(file flags.FilenameOptions) *InteractiveUOValueBuilder {
	// Read the file into a UniversalOrchestratorHelmValues struct
	bytes, err := file.Read()
	if err != nil {
		log.Printf("[ERROR] Error reading file: %s", err)
		b.errs = append(b.errs, fmt.Errorf("error reading file: %s", err))
	}

	// Serialize the bytes into a UniversalOrchestratorHelmValues struct
	err = yaml.Unmarshal(bytes, &b.newValues)
	if err != nil {
		b.errs = append(b.errs, fmt.Errorf("error unmarshalling values: %s", err))
	}

	return b
}

func (b *InteractiveUOValueBuilder) PreFlight() error {
	// Print any errors and exit if there are any
	if len(b.errs) > 0 {
		for _, err := range b.errs {
			cmdutil.PrintError(err)
		}
		return fmt.Errorf("exiting due to errors")
	}

	if b.newValues.CommandAgentURL == "" && b.commandHostname != "" {
		b.newValues.CommandAgentURL = fmt.Sprintf("https://%s/KeyfactorAgents", b.commandHostname)
	}

	return nil
}

func (b *InteractiveUOValueBuilder) staticBuild() error {
	if b.extensions == nil {
		return errors.New("extensions cannot be nil")
	}

	// Clear the init containers
	b.newValues.InitContainers = make([]InitContainer, 0)

	// Set the init containers for the extension installer
	b.setExtensionInstallerInitContainers(b.extensions)

	// Marshal the newValues struct into a yaml string
	buf, err := yaml.Marshal(b.newValues)
	if err != nil {
		return fmt.Errorf("failed to marshal newValues struct into yaml: %w", err)
	}

	// Set the newValuesString field
	b.newValuesString = string(buf)
	return nil
}

func (b *InteractiveUOValueBuilder) interactiveBuild() error {
	return b.MainMenu()
}

func (b *InteractiveUOValueBuilder) Build() (string, error) {
	var err error

	// If b.extensions is not nil, we are in non-interactive mode
	if b.interactiveMode {
		// We are in interactive mode
		err = b.interactiveBuild()
		if err != nil {
			return "", err
		}
	} else {
		err = b.staticBuild()
		if err != nil {
			return "", err
		}
	}

	return b.newValuesString, nil
}

// GetIsPositiveNumberValidator validates if an input is a number.
func GetIsPositiveNumberValidator() survey.Validator {
	return func(val interface{}) error {
		var theNumber int
		// The reflected value of the result
		value := reflect.ValueOf(val)

		switch value.Kind() {
		case reflect.Int:
			theNumber = int(value.Int())
			return nil
		case reflect.String:
			atoi, err := strconv.Atoi(value.String())
			if err != nil {
				return fmt.Errorf("value must be a number")
			}

			theNumber = atoi
		}

		if theNumber <= 0 {
			return errors.New("value mst be greater than 0")
		}

		return nil
	}
}

func (b *InteractiveUOValueBuilder) MainMenu() error {
	mainMenuOptions := []menuOption{
		{
			optionName:   "Configure UO Name",
			optionDesc:   "Configure the name of the Universal Orchestrator",
			currentValue: "",
			handlerFunc:  b.nameHandler,
		},
		{
			optionName:   "Change Command Agent URL",
			optionDesc:   "Change the base URL to the Command Agents API",
			currentValue: b.newValues.CommandAgentURL,
			handlerFunc: func() error {
				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter the base URL to the Command Agents API",
					Default:  b.newValues.CommandAgentURL,
				}
				err := survey.AskOne(&prompt, &b.newValues.CommandAgentURL, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get Command Agent URL: %w", err)
				}

				// Return to the main menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.MainMenu()
			},
		},
		{
			optionName:   "Change Replica Count",
			optionDesc:   "Change the number of Orchestrator replicas to create.",
			currentValue: b.newValues.ReplicaCount,
			handlerFunc: func() error {
				replicasString := strconv.Itoa(b.newValues.ReplicaCount)

				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter a non-zero number of Orchestrator replicas to create",
					Default:  replicasString,
				}
				err := survey.AskOne(&prompt, &replicasString, survey.WithValidator(GetIsPositiveNumberValidator()))
				if err != nil {
					return fmt.Errorf("failed to get replica count: %w", err)
				}

				b.newValues.ReplicaCount, err = strconv.Atoi(replicasString)
				if err != nil {
					return fmt.Errorf("failed to convert replica count to int: %w", err)
				}

				// Return to the main menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.MainMenu()
			},
		},
		{
			optionName:   "Change Log Level",
			optionDesc:   "Change the log level of the Universal Orchestrator container",
			currentValue: b.newValues.LogLevel,
			handlerFunc: func() error {
				prompt := survey.Select{
					Message: "Select the log level of the Universal Orchestrator container",
					Options: []string{"Trace", "Debug", "Info", "Warn", "Error"},
					Default: b.newValues.LogLevel,
				}
				err := survey.AskOne(&prompt, &b.newValues.LogLevel, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get log level: %w", err)
				}

				// Return to the main menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.MainMenu()
			},
		},
		{
			optionName:  "Change Image",
			optionDesc:  "Change the Universal Orchestrator container image that the chart will use",
			handlerFunc: b.imageHandler,
		},
		{
			optionName:  "Change Auth Settings",
			optionDesc:  "Change authentication settings",
			handlerFunc: b.authMenuHandler,
		},
		{
			optionName:   "Container Root CA Certificate Configuration",
			optionDesc:   "Configure the Universal Orchestrator container to trust a custom CA certificate chain",
			currentValue: "",
			handlerFunc:  b.caChainConfiguration,
		},
		{
			optionName:   "Configure Orchestrator Extensions",
			optionDesc:   "Set the orchestrator extensions to install with the chart",
			currentValue: fmt.Sprintf("%d extensions", len(b.newValues.InitContainers)),
			handlerFunc:  b.selectExtensionsHandler,
		},
		{
			optionName:   "Save and Exit",
			optionDesc:   "Exit the program and write the new newValues to override.yaml",
			currentValue: "",
			handlerFunc:  b.SaveAndExit,
		},
	}

	return b.handleOptions(mainMenuOptions, "Main Menu")
}

func (b *InteractiveUOValueBuilder) SaveAndExit() error {
	// Marshal the newValues struct into a yaml string
	buf, err := yaml.Marshal(b.newValues)
	if err != nil {
		return fmt.Errorf("failed to marshal newValues struct into yaml: %w", err)
	}

	if b.overrideFile == "" {
		// Write the yaml string locally to an override file
		err = os.WriteFile(b.overrideFile, buf, 0644)
	}

	// Set the newValuesString field
	b.newValuesString = string(buf)
	return nil
}

func (b *InteractiveUOValueBuilder) nameHandler() error {
	nameOptions := []menuOption{
		{
			optionName:   "Change Base Orchestrator Name",
			optionDesc:   "Change the base orchestrator name",
			currentValue: b.newValues.BaseOrchestratorName,
			handlerFunc: func() error {
				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter the name of the chart",
					Default:  b.newValues.BaseOrchestratorName,
				}
				err := survey.AskOne(&prompt, &b.newValues.BaseOrchestratorName, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get base orchestrator name: %w", err)
				}

				// Return to the name menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.nameHandler()
			},
		},
		{
			optionName:   "Change Complete Orchestrator Name",
			optionDesc:   "Change the complete orchestrator name and override any computed name",
			currentValue: b.newValues.CompleteName,
			handlerFunc: func() error {
				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter the name of the chart",
					Default:  b.newValues.CompleteName,
				}
				err := survey.AskOne(&prompt, &b.newValues.CompleteName, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get complete orchestrator name: %w", err)
				}

				// Return to the name menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.nameHandler()
			},
		},
		{
			optionName:   "Main Menu",
			optionDesc:   "Return to the main menu",
			currentValue: "",
			handlerFunc: func() error {
				return b.MainMenu()
			},
		},
	}

	return b.handleOptions(nameOptions, "Image Menu")
}

func (b *InteractiveUOValueBuilder) imageHandler() error {
	imageOptions := []menuOption{
		{
			optionName:   "Change Image Repository",
			optionDesc:   "Change the repository of the Universal Orchestrator container image",
			currentValue: b.newValues.Image.Repository,
			handlerFunc: func() error {
				prompt := survey.Input{
					Message: "Enter the repository of the Universal Orchestrator container image",
					Default: b.newValues.Image.Repository,
				}
				err := survey.AskOne(&prompt, &b.newValues.Image.Repository, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get image repository: %w", err)
				}

				// Return to the image menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.imageHandler()
			},
		},
		{
			optionName:   "Change Image Tag",
			optionDesc:   "Change the tag of the Universal Orchestrator container image",
			currentValue: b.newValues.Image.Tag,
			handlerFunc: func() error {
				prompt := survey.Input{
					Message: "Enter the tag of the Universal Orchestrator container image",
					Default: b.newValues.Image.Tag,
				}
				err := survey.AskOne(&prompt, &b.newValues.Image.Tag, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get image tag: %w", err)
				}

				// Return to the image menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.imageHandler()
			},
		},
		{
			optionName:   "Change Image Pull Policy",
			optionDesc:   "Change the pull policy of the Universal Orchestrator container image",
			currentValue: b.newValues.Image.PullPolicy,
			handlerFunc: func() error {
				prompt := survey.Input{
					Message: "Enter the pull policy to use when pulling the Universal Orchestrator container image",
					Default: b.newValues.Image.PullPolicy,
				}
				err := survey.AskOne(&prompt, &b.newValues.Image.PullPolicy, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get image pull policy: %w", err)
				}

				// Return to the image menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.imageHandler()
			},
		},
		{
			optionName:   "Main Menu",
			optionDesc:   "Return to the main menu",
			currentValue: "",
			handlerFunc: func() error {
				// Return to the main menu
				return b.MainMenu()
			},
		},
	}

	return b.handleOptions(imageOptions, "Image Menu")
}

func (b *InteractiveUOValueBuilder) authMenuHandler() error {
	imageOptions := []menuOption{
		{
			optionName:   "Change Authentication Secret Name",
			optionDesc:   "Change the configured name of the K8s secret containing credentials for Command",
			currentValue: b.newValues.Auth.SecretName,
			handlerFunc: func() error {
				prompt := survey.Input{
					Message: "Enter the name of the K8s secret containing credentials for Command",
					Default: b.newValues.Auth.SecretName,
				}
				err := survey.AskOne(&prompt, &b.newValues.Auth.SecretName, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get authentication secret name: %w", err)
				}

				// Return to the auth menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.authMenuHandler()
			},
		},
		{
			optionName:   "Use OAuth/IDP for Authentication to Command",
			optionDesc:   "Use or don't use OAuth/IDP for Authentication to Command",
			currentValue: b.newValues.Auth.UseOauthAuthentication,
			handlerFunc: func() error {
				prompt := survey.Confirm{
					Message: "Use OAuth/IDP for Authentication to Command?",
					Default: b.newValues.Auth.UseOauthAuthentication,
				}
				err := survey.AskOne(&prompt, &b.newValues.Auth.UseOauthAuthentication, survey.WithValidator(survey.Required))
				if err != nil {
					return fmt.Errorf("failed to get authentication secret name: %w", err)
				}

				// Return to the auth menu
				if b.exitAfterPrompt {
					return nil
				}

				return b.authMenuHandler()
			},
		},
		{
			optionName:   "Main Menu",
			optionDesc:   "Return to the main menu",
			currentValue: "",
			handlerFunc: func() error {
				// Return to the main menu
				return b.MainMenu()
			},
		},
	}

	return b.handleOptions(imageOptions, "Auth Menu")
}

func (b *InteractiveUOValueBuilder) handleOptions(options []menuOption, help string) error {
	// Build list of options
	optionStrings := make([]string, 0)
	for _, option := range options {
		optionStrings = append(optionStrings, option.optionName)
	}

	descriptionFunc := func(value string, index int) string {
		for _, option := range options {
			if option.optionName == value {
				desc := option.optionDesc
				currentValueString := anyToString(option.currentValue)
				if currentValueString != "" {
					desc = fmt.Sprintf("%s (currently %q)", desc, currentValueString)
				}
				return desc
			}
		}
		return ""
	}

	// Prompt the user to select an option
	selected := ""
	prompt := &survey.Select{
		Message:     "Select an option:",
		Options:     optionStrings,
		Help:        help,
		Description: descriptionFunc,
	}
	err := survey.AskOne(prompt, &selected)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return err
	}

	// Call the handler function for the selected option
	for _, option := range options {
		if option.optionName == selected {
			return option.handlerFunc()
		}
	}

	// We should never get here

	// If we do get here, the option was not found
	errorMessage := fmt.Sprintf("error: Option %s not found\n", selected)
	fmt.Printf(errorMessage)

	// Return to the main menu
	if b.exitAfterPrompt {
		return nil
	}

	return b.MainMenu()
}

func anyToString(v any) string {
	if v == nil {
		return ""
	}

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', 6, 64)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	default:
		// For types not handled, you could return a string like this,
		// or handle more types as needed.
		return fmt.Sprintf("Unsupported type: %s", rv.Type())
	}
}
