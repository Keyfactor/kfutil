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

package helm

import (
	"fmt"
	"kfutil/pkg/cmdutil/extensions"
	"strings"
)

func (b *InteractiveUOValueBuilder) cacheCurrentlyInstalledExtensions() extensions.Extensions {
	installedExtensions := make(extensions.Extensions)
	extensionInitContainer := InitContainer{}

	for _, container := range b.newValues.InitContainers {
		if container.Name == installerInitContainerName {
			extensionInitContainer = container
			break
		}
	}

	if extensionInitContainer.Name == "" {
		return installedExtensions
	}

	// Parse the extension list from the init container args
	for _, env := range extensionInitContainer.Args {
		if strings.HasPrefix(env, "--extension=") {
			extension, _ := extensions.ParseExtensionString(strings.ReplaceAll(env, "--extension=", ""))
			installedExtensions[extension.Name] = extension.Version
		}
	}

	return installedExtensions
}

// setExtensionInstallerInitContainers sets the extension installer container args verbatim from the extensionsToInstall map.
// If an extension is already installed, it will be upgraded to the latest version.
func (b *InteractiveUOValueBuilder) setExtensionInstallerInitContainers(extensionsToInstall extensions.Extensions) {
	var extensionInstallerInitContainer *InitContainer

	// Reference the extension installer init container
	for i, container := range b.newValues.InitContainers {
		if container.Name == installerInitContainerName {
			extensionInstallerInitContainer = &b.newValues.InitContainers[i]
			break
		}
	}

	// Create the extension installer init container if it does not exist
	if extensionInstallerInitContainer == nil {
		initContainer := InitContainer{
			Name:            installerInitContainerName,
			Image:           installerImage,
			ImagePullPolicy: installerImagePullPolicy,
			Env:             []Environment{},
			VolumeMounts: []VolumeMount{
				{
					Name:      "command-pv-claim",
					MountPath: "/app/extensions",
					SubPath:   "",
					ReadOnly:  false,
				},
			},
			Command: []string{"kfutil", "orchs", "extensions", "--out=/app/extensions", "--confirm", "-P"},
			Args:    []string{},
		}
		b.newValues.InitContainers = append(b.newValues.InitContainers, initContainer)

		// Reference the extension installer init container
		extensionInstallerInitContainer = &b.newValues.InitContainers[len(b.newValues.InitContainers)-1]
	}

	// Reset the extension installer init container args
	extensionInstallerInitContainer.Args = []string{}

	for extension, latestVersion := range extensionsToInstall {
		extensionAlreadyInstalled := false

		// Upgrade the extension if it is already installed
		for i, arg := range extensionInstallerInitContainer.Args {
			if strings.HasPrefix(arg, "--extension=") {
				installedExtension, _ := extensions.ParseExtensionString(arg)
				if installedExtension.Name == extension && installedExtension.Version != latestVersion {
					extensionAlreadyInstalled = true
					extensionInstallerInitContainer.Args[i] = fmt.Sprintf("--extension=%s@%s", extension, latestVersion)
					b.AddRuntimeLog(fmt.Sprintf("Upgrading extension %s to version %s", extension, latestVersion))
					break
				}
			}
		}

		if extensionAlreadyInstalled {
			continue
		}

		// Install the extension if it is not already installed
		extensionInstallerInitContainer.Args = append(extensionInstallerInitContainer.Args, fmt.Sprintf("--extension=%s@%s", extension, latestVersion))
	}
}

func (b *InteractiveUOValueBuilder) selectExtensionsHandler() error {
	// Create an extension installer tool
	installerTool := extensions.NewExtensionInstallerBuilder().Token(b.token)

	// Get a list of currently installed extensions
	installedExtensions := b.cacheCurrentlyInstalledExtensions()

	// Set up the installer tool
	installerTool.SetCurrentlyInstalled(installedExtensions)

	// Prompt the user to select which extensions to install
	err := installerTool.PromptForExtensions()
	if err != nil {
		return fmt.Errorf("failed to prompt for extensions: %s", err)
	}

	// Get the list of extensions to install
	extensionsToInstall := installerTool.GetExtensionsToInstall()

	// Copy installedExtensions to extensionsToInstall.
	// setExtensionInstallerInitContainers will upgrade any extensions that are already installed,
	// and will remove any extensions that are no longer selected.
	for extension, version := range installedExtensions {
		if _, ok := extensionsToInstall[extension]; !ok {
			extensionsToInstall[extension] = version
		}
	}

	// Set the init containers for the extension installer
	b.setExtensionInstallerInitContainers(extensionsToInstall)

	// Return to the auth menu
	if b.exitAfterPrompt {
		return nil
	}

	return b.MainMenu()
}
