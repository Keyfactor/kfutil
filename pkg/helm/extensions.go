package helm

import (
	"fmt"
	"kfutil/pkg/cmdutil/extensions"
)

func (b *InteractiveUOValueBuilder) cacheCurrentlyInstalledExtensions() extensions.Extensions {
	installedExtensions := make(extensions.Extensions)
	for _, container := range b.newValues.InitContainers {
		var extensionName extensions.ExtensionName
		extensionVersion := extensions.Version("")
		for _, env := range container.Env {
			if env.Name == "EXTENSION_NAME" {
				extensionName = extensions.ExtensionName(env.Value)
			}
			if env.Name == "EXTENSION_VERSION" {
				extensionVersion = extensions.Version(env.Value)
			}
		}
		if extensionName != "" {
			installedExtensions[extensionName] = extensionVersion
		}
	}

	return installedExtensions
}

func (b *InteractiveUOValueBuilder) setExtensionInstallerInitContainers(extensionsToInstall extensions.Extensions) {
	for extension, latestVersion := range extensionsToInstall {
		initContainerName := fmt.Sprintf("%s-installer", extension)
		extensionAlreadyInstalled := false

		// If the extension is already covered in the Init Containers, upgrade it
		for i, container := range b.newValues.InitContainers {
			if container.Name == initContainerName {
				extensionAlreadyInstalled = true
				for j, env := range container.Env {
					if env.Name == "EXTENSION_VERSION" && env.Value != string(latestVersion) {
						fmt.Printf("Upgrading %s from %s to %s\n", extension, env.Value, latestVersion)
						b.newValues.InitContainers[i].Env[j].Value = string(latestVersion)
						break
					} else if env.Name == "EXTENSION_VERSION" && env.Value == string(latestVersion) {
						fmt.Printf("Extension %s is already configured for version %s\n", extension, latestVersion)
					}
				}
				break
			}
		}

		if extensionAlreadyInstalled {
			continue
		}

		b.newValues.InitContainers = append(b.newValues.InitContainers, InitContainer{
			Name:            initContainerName,
			Image:           installerImage,
			ImagePullPolicy: installerImagePullPolicy,
			Env: []Environment{
				{
					Name:  "EXTENSION_NAME",
					Value: string(extension),
				},
				{
					Name:  "EXTENSION_VERSION",
					Value: string(latestVersion),
				},
				{
					Name:  "INSTALL_PATH",
					Value: fmt.Sprintf("/app/extensionList/%s", extension),
				},
			},
			VolumeMounts: []VolumeMount{
				{
					Name:      b.newValues.ExtensionStorage.Name,
					MountPath: "/app/extensionList",
				},
			},
		})
	}
}

func (b *InteractiveUOValueBuilder) selectExtensionsHandler() error {
	// Create an extension installer tool
	installerTool := extensions.NewExtensionInstallerBuilder().Token(b.token)

	// Get a list of currently installed extensions
	installedExtensions := b.cacheCurrentlyInstalledExtensions()

	// Set up the installer tool
	installerTool.SetExtensionsToInstall(installedExtensions)

	// Prompt the user to select which extensions to install
	err := installerTool.PromptForExtensions()
	if err != nil {
		return fmt.Errorf("failed to prompt for extensions: %s", err)
	}

	// Get the list of extensions to install
	extensionsToInstall := installerTool.GetExtensionsToInstall()

	// Set the init containers for the extension installer
	b.setExtensionInstallerInitContainers(extensionsToInstall)

	// Return to the auth menu
	if b.exitAfterPrompt {
		return nil
	}

	return b.MainMenu()
}
