/*
Copyright 2023 The Keyfactor Command Authors.

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
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Keyfactor/keyfactor-go-client/v2/api"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"kfutil/pkg/cmdutil/flags"
	"kfutil/pkg/keyfactor/v1"
)

// Ensure that StoreTypesGetFlags implements Flags
var _ flags.Flags = &StoreTypesGetFlags{}
var _ flags.Options = &StoreTypesGetOptions{}

type StoreTypesGetFlags struct {
	// storeTypeID is the ID of the certificate store type to get.
	storeTypeID *int
	// storeTypeName is the name of the certificate store type to get.
	storeTypeName *string
	// genericFormat strips all fields specific to the Command instance from the output.
	genericFormat *bool
	// gitRef is the git branch or tag to reference when pulling store-types from the internet.
	gitRef *string
	// outputToIntegrationManifest updates the integration manifest with the store type. It overrides the store type in the manifest if it already exists.
	outputToIntegrationManifest *bool
}

func CreateStoreTypesGetFlags() *StoreTypesGetFlags {
	var storeTypeID int
	var storeTypeName string
	var genericFormat bool
	var gitRef string
	var outputToIntegrationManifest bool

	return &StoreTypesGetFlags{
		storeTypeID:                 &storeTypeID,
		storeTypeName:               &storeTypeName,
		genericFormat:               &genericFormat,
		gitRef:                      &gitRef,
		outputToIntegrationManifest: &outputToIntegrationManifest,
	}
}

func (f *StoreTypesGetFlags) AddFlags(flags *pflag.FlagSet) {
	flags.IntVarP(f.storeTypeID, "id", "i", -1, "ID of the certificate store type to get.")
	flags.StringVarP(f.storeTypeName, "name", "n", "", "Name of the certificate store type to get.")
	flags.BoolVarP(f.genericFormat, "generic", "g", false, "Output the store type in a generic format stripped of all fields specific to the Command instance.")
	flags.StringVarP(f.gitRef, FlagGitRef, "b", "main", "The git branch or tag to reference when pulling store-types from the internet.")
	flags.BoolVarP(f.outputToIntegrationManifest, "output-to-integration-manifest", "", false, "Update the integration manifest with the store type. It overrides the store type in the manifest if it already exists. If the integration manifest does not exist in the current directory, it will be created.")
}

func CreateCmdStoreTypesGet() *cobra.Command {
	storeTypesGetFlags := CreateStoreTypesGetFlags()

	cmd := &cobra.Command{
		Use:   StoreTypesGetUsage,
		Short: StoreTypesGetShort,
		Long:  StoreTypesGetLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Compute the runtime options from flags passed to the command
			options, err := storeTypesGetFlags.ToOptions(cmd, args)
			if err != nil {
				return fmt.Errorf("failed to get runtime options from flags: %s", err)
			}
			if options.storeTypeInterface == nil {
				return fmt.Errorf("store type not specified - this should never happen")
			}

			// Silence usage on error
			cmd.SilenceUsage = true

			// Debug + expEnabled checks
			debugErr := warnExperimentalFeature(expEnabled, false)
			if debugErr != nil {
				return debugErr
			}
			informDebug(debugFlag)

			// Authenticate
			authConfig := createAuthConfigFromParams(kfcHostName, kfcUsername, kfcPassword, kfcDomain, kfcAPIPath)
			kfClient, _ := initClient(configFile, profile, providerType, providerProfile, noPrompt, authConfig, false)

			if kfClient == nil {
				return fmt.Errorf("failed to initialize Keyfactor client")
			}

			storeTypes, err := kfClient.GetCertificateStoreType(options.storeTypeInterface)
			if err != nil {
				log.Error().Err(err).Msg(fmt.Sprintf("unable to get certificate store type %s", options.storeTypeInterface))
				return err
			}
			output, jErr := formatStoreTypeOutput(storeTypes, outputFormat, options.outputType)
			if jErr != nil {
				log.Error().Err(jErr).Msg("unable to format certificate store type output")
				return jErr
			}

			// If outputToIntegrationManifest is true, update the integration manifest with the store type
			if options.outputToIntegrationManifest {
				imv1 := manifestv1.CreateIntegrationManifest()
				err = imv1.LoadFromFilesystem()
				if err != nil {
					return err
				}

				err = imv1.CopyIntoStoreType(output)
				if err != nil {
					return err
				}

				err = imv1.SaveToFilesystem()
				if err != nil {
					return err
				}

				_, err = cmd.OutOrStdout().Write([]byte(fmt.Sprintf("Successfully updated integration manifest with store type %s\n", options.storeTypeInterface)))
			} else {
				_, err = cmd.OutOrStdout().Write([]byte(output))
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	// Add the flags
	storeTypesGetFlags.AddFlags(cmd.Flags())

	return cmd
}

type StoreTypesGetOptions struct {
	storeTypeID                 int
	storeTypeName               string
	genericFormat               bool
	gitRef                      string
	storeTypeInterface          interface{}
	outputType                  string
	outputToIntegrationManifest bool
}

func (f *StoreTypesGetFlags) ToOptions(cmd *cobra.Command, args []string) (*StoreTypesGetOptions, error) {
	if len(args) > 0 {
		return nil, fmt.Errorf("unexpected arguments: %v", args)
	}

	options := &StoreTypesGetOptions{}

	// Global flags
	flags.GetDebugFlag(cmd)

	// Get the values from the flags
	if f.storeTypeID != nil {
		options.storeTypeID = *f.storeTypeID
	}

	if f.storeTypeName != nil {
		options.storeTypeName = *f.storeTypeName
	}

	if f.genericFormat != nil {
		options.genericFormat = *f.genericFormat
	}

	if f.gitRef != nil {
		options.gitRef = *f.gitRef
	}

	if f.outputToIntegrationManifest != nil {
		options.outputToIntegrationManifest = *f.outputToIntegrationManifest
	}

	return options, options.Validate()
}

func (f *StoreTypesGetOptions) Validate() error {
	// storeTypeID and storeTypeName are mutually exclusive
	if f.storeTypeID > 0 && f.storeTypeName != "" {
		return fmt.Errorf("only one of --id or --name can be provided")
	}

	// Check inputs and prompt if necessary
	// The f.storeTypeInterface is used to pass the store type to the API
	if f.storeTypeID < 0 && f.storeTypeName == "" {
		validStoreTypes := getValidStoreTypes("", f.gitRef)
		prompt := &survey.Select{
			Message: "Choose a store type:",
			Options: validStoreTypes,
		}
		var selected string
		err := survey.AskOne(prompt, &selected)
		if err != nil {
			fmt.Println(err)
			return err
		}
		f.storeTypeInterface = selected
	} else if f.storeTypeID >= 0 {
		f.storeTypeInterface = f.storeTypeID
	} else if f.storeTypeName != "" {
		f.storeTypeInterface = f.storeTypeName
	} else {
		log.Error().Err(InvalidInputError).Send()
		return InvalidInputError
	}

	// Set the default git ref if it is empty
	if f.gitRef == "" {
		f.gitRef = "main"
	}

	// Set the output type to full unless genericFormat is true
	f.outputType = "full"
	if f.genericFormat {
		f.outputType = "generic"
	}

	// If outputToIntegrationManifest is true, set the output type to generic
	if f.outputToIntegrationManifest {
		f.outputType = "generic"
	}

	return nil
}

func formatStoreTypeOutput(storeType *api.CertificateStoreType, outputFormat string, outputType string) (string, error) {
	var sOut interface{}
	sOut = storeType
	if outputType == "generic" {
		// Convert to api.GenericCertificateStoreType
		var genericProperties []api.StoreTypePropertyDefinitionGeneric
		for _, prop := range *storeType.Properties {
			genericProp := api.StoreTypePropertyDefinitionGeneric{
				Name:         prop.Name,
				DisplayName:  prop.DisplayName,
				Type:         prop.Type,
				DependsOn:    prop.DependsOn,
				DefaultValue: prop.DefaultValue,
				Required:     prop.Required,
			}
			genericProperties = append(genericProperties, genericProp)
		}

		var genericEntryParameters []api.EntryParameterGeneric
		for _, param := range *storeType.EntryParameters {
			genericParam := api.EntryParameterGeneric{
				Name:         param.Name,
				DisplayName:  param.DisplayName,
				Type:         param.Type,
				RequiredWhen: param.RequiredWhen,
				DependsOn:    param.DependsOn,
				DefaultValue: param.DefaultValue,
				Options:      param.Options,
			}
			genericEntryParameters = append(genericEntryParameters, genericParam)
		}

		genericStoreType := api.CertificateStoreTypeGeneric{
			Name:                storeType.Name,
			ShortName:           storeType.ShortName,
			Capability:          storeType.Capability,
			SupportedOperations: storeType.SupportedOperations,
			Properties:          &genericProperties,
			EntryParameters:     &genericEntryParameters,
			PasswordOptions:     storeType.PasswordOptions,
			//StorePathType:       storeType.StorePathType,
			StorePathValue:    storeType.StorePathValue,
			PrivateKeyAllowed: storeType.PrivateKeyAllowed,
			//JobProperties:      jobProperties,
			ServerRequired:     storeType.ServerRequired,
			PowerShell:         storeType.PowerShell,
			BlueprintAllowed:   storeType.BlueprintAllowed,
			CustomAliasAllowed: storeType.CustomAliasAllowed,
		}
		sOut = genericStoreType
	}

	switch {
	case outputFormat == "yaml" || outputFormat == "yml":
		output, jErr := yaml.Marshal(sOut)
		if jErr != nil {
			return "", jErr
		}
		return fmt.Sprintf("%s", output), nil
	default:
		output, jErr := json.MarshalIndent(sOut, "", "  ")
		if jErr != nil {
			return "", jErr
		}
		return fmt.Sprintf("%s", output), nil
	}
}

const (
	StoreTypesGetUsage = `get [-i <store-type-id> | -n <store-type-name>] [-g] [-b <git-ref>] [-o]`
	StoreTypesGetShort = `Get a specific store type by either name or ID.`
	StoreTypesGetLong  = StoreTypesGetShort
)
