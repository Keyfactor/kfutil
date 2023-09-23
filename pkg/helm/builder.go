package helm

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"kfutil/pkg/cmdutil"
	"kfutil/pkg/cmdutil/flags"
	"log"
)

type ToolBuilder struct {
	errs            []error
	commandHostname string
	overrideFile    string
	token           string
	values          UniversalOrchestratorHelmValues
}

func NewToolBuilder() *ToolBuilder {
	return &ToolBuilder{}
}

func (b *ToolBuilder) CommandHostname(hostname string) *ToolBuilder {
	b.commandHostname = hostname

	return b
}

func (b *ToolBuilder) OverrideFile(filename string) *ToolBuilder {
	b.overrideFile = filename

	return b
}

func (b *ToolBuilder) Token(token string) *ToolBuilder {
	b.token = token

	return b
}

func (b *ToolBuilder) Values(file flags.FilenameOptions) *ToolBuilder {
	// Read the file into a UniversalOrchestratorHelmValues struct
	bytes, err := file.Read()
	if err != nil {
		log.Printf("[ERROR] Error reading file: %s", err)
		b.errs = append(b.errs, fmt.Errorf("error reading file: %s", err))
	}

	// Serialize the bytes into a UniversalOrchestratorHelmValues struct
	err = yaml.Unmarshal(bytes, &b.values)
	if err != nil {
		b.errs = append(b.errs, fmt.Errorf("error unmarshalling values: %s", err))
	}

	return b
}

func (b *ToolBuilder) PreFlight() *ToolBuilder {
	// Print any errors and exit if there are any
	if len(b.errs) > 0 {
		for _, err := range b.errs {
			cmdutil.PrintError(err)
		}
		log.Fatal("[ERROR] Exiting due to errors")
	}

	return b
}

func (b *ToolBuilder) BuildUniversalOrchestratorHelmValueTool() func() error {
	return func() error {
		err := NewUniversalOrchestratorHelmValueBuilder(b).Build()
		if err != nil {
			return fmt.Errorf("interactive value builder tool exited: %s", err)
		}
		return nil
	}
}
