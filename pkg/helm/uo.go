package helm

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v3"
	"os"
	"sort"
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
	currentValue string
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

func (b *UniversalOrchestratorHelmValueBuilder) MainMenu() error {
	mainMenuOptions := []menuOption{
		{
			optionName:   "Change Name",
			optionDesc:   "Change the name of the chart",
			currentValue: b.values.Name,
			handlerFunc: func() error {
				prompt := survey.Input{
					Renderer: survey.Renderer{},
					Message:  "Enter the name of the chart",
					Default:  b.values.Name,
				}
				err := survey.AskOne(&prompt, &b.values.Name, survey.WithValidator(survey.Required))
				if err != nil {
					return err
				}

				// Return to the main menu
				return b.MainMenu()
			},
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
			optionName:   "Change Log Level",
			optionDesc:   "Change the log level of the Universal Orchestrator container",
			currentValue: b.values.LogLevel,
			handlerFunc: func() error {
				prompt := survey.Select{
					Message: "Select the log level of the Universal Orchestrator container",
					Options: []string{"Debug", "Info", "Warn", "Error"},
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
				if option.currentValue != "" {
					desc = fmt.Sprintf("%s (currently %q)", desc, option.currentValue)
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

func alphabetize(list []string) []string {
	// Make a copy of the original list
	sortedList := make([]string, len(list))
	copy(sortedList, list)

	// Sort the copied list alphabetically
	sort.Strings(sortedList)

	return sortedList
}
