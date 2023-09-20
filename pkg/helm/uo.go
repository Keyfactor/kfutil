package helm

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"sort"
	"strconv"
)

const (
	defaultOverrideFile      = "override.yaml"
	installerImage           = "m8rmclarenkf/uo_extension_installer:1.0.5"
	installerImagePullPolicy = "IfNotPresent"
)

type UniversalOrchestratorHelmValueBuilder struct {
	commandHostname string
	valuesFile      string
	overrideFile    string
	token           string
	values          UniversalOrchestratorHelmValues
}

type menuOption struct {
	optionName   string
	optionDesc   string
	currentValue any
	handlerFunc  func() error
}

func NewUniversalOrchestratorHelmValueBuilder(filePath string) *UniversalOrchestratorHelmValueBuilder {
	return &UniversalOrchestratorHelmValueBuilder{
		valuesFile:   filePath,
		overrideFile: defaultOverrideFile,
	}
}

func (b *UniversalOrchestratorHelmValueBuilder) SetGithubToken(token string) {
	b.token = token
}

func (b *UniversalOrchestratorHelmValueBuilder) SetOverrideFile(filePath string) {
	b.overrideFile = filePath
}

func (b *UniversalOrchestratorHelmValueBuilder) SetHostname(hostname string) {
	b.commandHostname = hostname
}

func (b *UniversalOrchestratorHelmValueBuilder) load() error {
	// Read in the values file and marshal it into the values struct
	buf, err := os.ReadFile(b.valuesFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, &b.values)
	if err != nil {
		return err
	}

	// Set the Command Agent URL to the hostname of the currently logged-in user if it is not already set
	if b.values.CommandAgentURL == "" {
		b.values.CommandAgentURL = fmt.Sprintf("https://%s/KeyfactorAgents", b.commandHostname)
	}

	return nil
}

func (b *UniversalOrchestratorHelmValueBuilder) Build() {
	err := b.load()
	if err != nil {
		fmt.Printf("Failed to load values file: %v\n", err)
		return
	}

	err = b.MainMenu()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}

// GetIsPositiveNumberValidator validates if an input is a number.
func GetIsPositiveNumberValidator() survey.Validator {
	return func(val interface{}) error {
		var theNumber int
		// the reflect value of the result
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

func (b *UniversalOrchestratorHelmValueBuilder) MainMenu() error {
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
			currentValue: b.values.CommandAgentURL,
			handlerFunc: func() error {
				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter the base URL to the Command Agents API",
					Default:  b.values.CommandAgentURL,
				}
				err := survey.AskOne(&prompt, &b.values.CommandAgentURL, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
				return b.MainMenu()
			},
		},
		{
			optionName:   "Change Replica Count",
			optionDesc:   "Change the number of Orchestrator replicas to create.",
			currentValue: b.values.ReplicaCount,
			handlerFunc: func() error {
				replicasString := strconv.Itoa(b.values.ReplicaCount)

				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter a non-zero number of Orchestrator replicas to create",
					Default:  replicasString,
				}
				err := survey.AskOne(&prompt, &replicasString, survey.WithValidator(GetIsPositiveNumberValidator()))
				if err != nil {
					return err
				}

				b.values.ReplicaCount, err = strconv.Atoi(replicasString)
				if err != nil {
					return err
				}

				// Return to the main menu
				return b.MainMenu()
			},
		},
		{
			optionName:   "Change Log Level",
			optionDesc:   "Change the log level of the Universal Orchestrator container",
			currentValue: b.values.LogLevel,
			handlerFunc: func() error {
				prompt := survey.Select{
					Message: "Select the log level of the Universal Orchestrator container",
					Options: []string{"Trace", "Debug", "Info", "Warn", "Error"},
					Default: b.values.LogLevel,
				}
				err := survey.AskOne(&prompt, &b.values.LogLevel, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
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
			currentValue: fmt.Sprintf("%d extensions", len(b.values.InitContainers)),
			handlerFunc:  b.selectExtensionsHandler,
		},
		{
			optionName:   "Save and Exit",
			optionDesc:   "Exit the program and write the new values to override.yaml",
			currentValue: "",
			handlerFunc: func() error {
				// Marshal the values struct into a yaml string
				buf, err := yaml.Marshal(b.values)
				if err != nil {
					return err
				}

				// Write the yaml string locally to an override file
				err = os.WriteFile(b.overrideFile, buf, 0644)

				// TODO print Helm command to install the chart
				return nil
			},
		},
	}

	return b.handleOptions(mainMenuOptions, "Main Menu")
}

func (b *UniversalOrchestratorHelmValueBuilder) nameHandler() error {
	nameOptions := []menuOption{
		{
			optionName:   "Change Base Orchestrator Name",
			optionDesc:   "Change the base orchestrator name",
			currentValue: b.values.BaseOrchestratorName,
			handlerFunc: func() error {
				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter the name of the chart",
					Default:  b.values.BaseOrchestratorName,
				}
				err := survey.AskOne(&prompt, &b.values.BaseOrchestratorName, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
				return b.nameHandler()
			},
		},
		{
			optionName:   "Change Complete Orchestrator Name",
			optionDesc:   "Change the complete orchestrator name and override any computed name",
			currentValue: b.values.CompleteName,
			handlerFunc: func() error {
				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter the name of the chart",
					Default:  b.values.CompleteName,
				}
				err := survey.AskOne(&prompt, &b.values.CompleteName, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
				return b.nameHandler()
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

	return b.handleOptions(nameOptions, "Image Menu")
}

func (b *UniversalOrchestratorHelmValueBuilder) imageHandler() error {
	imageOptions := []menuOption{
		{
			optionName:   "Change Image Repository",
			optionDesc:   "Change the repository of the Universal Orchestrator container image",
			currentValue: b.values.Image.Repository,
			handlerFunc: func() error {
				prompt := survey.Input{
					Message: "Enter the repository of the Universal Orchestrator container image",
					Default: b.values.Image.Repository,
				}
				err := survey.AskOne(&prompt, &b.values.Image.Repository, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
				return b.imageHandler()
			},
		},
		{
			optionName:   "Change Image Tag",
			optionDesc:   "Change the tag of the Universal Orchestrator container image",
			currentValue: b.values.Image.Tag,
			handlerFunc: func() error {
				prompt := survey.Input{
					Message: "Enter the tag of the Universal Orchestrator container image",
					Default: b.values.Image.Tag,
				}
				err := survey.AskOne(&prompt, &b.values.Image.Tag, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
				return b.imageHandler()
			},
		},
		{
			optionName:   "Change Image Pull Policy",
			optionDesc:   "Change the pull policy of the Universal Orchestrator container image",
			currentValue: b.values.Image.PullPolicy,
			handlerFunc: func() error {
				prompt := survey.Input{
					Message: "Enter the pull policy to use when pulling the Universal Orchestrator container image",
					Default: b.values.Image.PullPolicy,
				}
				err := survey.AskOne(&prompt, &b.values.Image.PullPolicy, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
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

func (b *UniversalOrchestratorHelmValueBuilder) authMenuHandler() error {
	imageOptions := []menuOption{
		{
			optionName:   "Change Authentication Secret Name",
			optionDesc:   "Change the configured name of the K8s secret containing credentials for Command",
			currentValue: b.values.Auth.SecretName,
			handlerFunc: func() error {
				prompt := survey.Input{
					Message: "Enter the name of the K8s secret containing credentials for Command",
					Default: b.values.Auth.SecretName,
				}
				err := survey.AskOne(&prompt, &b.values.Auth.SecretName, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
				return b.imageHandler()
			},
		},
		{
			optionName:   "Use OAuth/IDP for Authentication to Command",
			optionDesc:   "Use or don't use OAuth/IDP for Authentication to Command",
			currentValue: b.values.Auth.UseOauthAuthentication,
			handlerFunc: func() error {
				prompt := survey.Confirm{
					Message: "Use OAuth/IDP for Authentication to Command?",
					Default: b.values.Auth.UseOauthAuthentication,
				}
				err := survey.AskOne(&prompt, &b.values.Auth.UseOauthAuthentication, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
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

func (b *UniversalOrchestratorHelmValueBuilder) handleOptions(options []menuOption, help string) error {
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

	// If we get here, the option was not found
	errorMessage := fmt.Sprintf("error: Option %s not found\n", selected)
	fmt.Printf(errorMessage)
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

func alphabetize(list []string) []string {
	// Make a copy of the original list
	sortedList := make([]string, len(list))
	copy(sortedList, list)

	// Sort the copied list alphabetically
	sort.Strings(sortedList)

	return sortedList
}
