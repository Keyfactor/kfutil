package helm

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"kfutil/pkg/cmdutil"
	"log"
	"strings"
)

func (b *InteractiveUOValueBuilder) selectExtensionsHandler() error {
	ghTool := NewGithubReleaseFetcher(b.token)
	extensions, err := ghTool.GetExtensionList()
	if err != nil {
		cmdutil.PrintError(err)
		return b.MainMenu()
	}

	// Get a list of currently installed extensions
	installedExtensions := make(Extensions)
	for _, container := range b.newValues.InitContainers {
		extensionName := ""
		extensionVersion := ""
		for _, env := range container.Env {
			if env.Name == "EXTENSION_NAME" {
				extensionName = env.Value
			}
			if env.Name == "EXTENSION_VERSION" {
				extensionVersion = env.Value
			}
		}
		if extensionName != "" {
			installedExtensions[extensionName] = append(installedExtensions[extensionName], extensionVersion)
		}
	}

	extensionNameList := make([]string, 0)
	for extensionName, _ := range extensions {
		extensionNameList = append(extensionNameList, extensionName)
	}

	descriptionFunc := func(value string, index int) string {
		for extensionName, versions := range extensions {
			if extensionName == value {
				description := fmt.Sprintf("%s", versions[0])

				if installed, ok := installedExtensions[extensionName]; ok {
					description = fmt.Sprintf("%s (currently %s)", description, installed[0])
				}

				return description
			}
		}
		return ""
	}

	// TODO make selection process declarative - i.e. extensions not in the extensionsToInstall slice are not installed
	var extensionsToInstall []string
	prompt := &survey.MultiSelect{
		Message:     "Select the extensions to install - the most recent versions are displayed",
		Options:     alphabetize(extensionNameList),
		Description: descriptionFunc,
	}
	err = survey.AskOne(prompt, &extensionsToInstall)
	if err != nil {
		return err
	}

	for _, extension := range extensionsToInstall {
		// extensionsToInstall is a subset of extensions since user was prompted to select from extensions
		// So we can safely assume that extension is in extensions
		versions := extensions[extension]
		// Also, extensions is only populated if there are releases for the extension
		latestVersion := versions[0]

		initContainerName := fmt.Sprintf("%s-installer", extension)
		extensionAlreadyInstalled := false

		// If the extension is already covered in the Init Containers, upgrade it
		for i, container := range b.newValues.InitContainers {
			if container.Name == initContainerName {
				extensionAlreadyInstalled = true
				for j, env := range container.Env {
					if env.Name == "EXTENSION_VERSION" && env.Value != latestVersion {
						fmt.Printf("Upgrading %s from %s to %s\n", extension, env.Value, latestVersion)
						b.newValues.InitContainers[i].Env[j].Value = latestVersion
						break
					} else if env.Name == "EXTENSION_VERSION" && env.Value == latestVersion {
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
					Value: extension,
				},
				{
					Name:  "EXTENSION_VERSION",
					Value: latestVersion,
				},
				{
					Name:  "INSTALL_PATH",
					Value: fmt.Sprintf("/app/extensions/%s", extension),
				},
			},
			VolumeMounts: []VolumeMount{
				{
					Name:      b.newValues.ExtensionStorage.Name,
					MountPath: "/app/extensions",
				},
			},
		})
	}

	// Return to the main menu
	return b.MainMenu()
}

type Extensions map[string][]string

type GithubReleaseFetcher struct {
	token string
}

func NewGithubReleaseFetcher(token string) *GithubReleaseFetcher {
	return &GithubReleaseFetcher{
		token: token,
	}
}

func (g *GithubReleaseFetcher) getOrchestratorNames() ([]string, error) {
	orchestratorList := make([]string, 0)

	for page := 1; page < 100; page++ {
		// Prevent rate limiting by setting upper bound to 100

		// Ask https://api.github.com/orgs/keyfactor/repos for the list of repos
		// Unmarshal the body into a slice of gitHubRepo structs
		var repos []gitHubRepo
		err := g.Get(fmt.Sprintf("https://api.github.com/orgs/keyfactor/repos?type=public&page=%d&per_page=100", page), &repos)
		if err != nil {
			return nil, err
		}

		// If the length of the repos slice is 0, we've reached the end of the list
		if len(repos) == 0 {
			break
		}

		// Loop through the repos and add them to the orchestratorList slice
		for _, repo := range repos {
			// If the repo ends with "-orchestrator" or "-pam, add it to the list
			if strings.HasSuffix(repo.Name, "-orchestrator") || strings.HasSuffix(repo.Name, "-pam") {
				orchestratorList = append(orchestratorList, repo.Name)
			}
		}
	}

	return orchestratorList, nil
}

func (g *GithubReleaseFetcher) GetExtensionList() (Extensions, error) {
	extensionNameList, err := g.getOrchestratorNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get list of extensions: %s", err)
	}

	extensions := make(Extensions)

	for _, extensionName := range extensionNameList {
		// Ask https://api.github.com/repos/keyfactor/{name}/releases for the list of releases
		// Unmarshal the body into a slice of gitHubRelease structs
		var releases []gitHubRelease
		err = g.Get(fmt.Sprintf("https://api.github.com/repos/keyfactor/%s/releases", extensionName), &releases)
		if err != nil {
			return nil, err
		}

		// Add the extension to the list of extensions
		versions := make([]string, 0)
		for _, release := range releases {
			if !release.Prerelease {
				versions = append(versions, release.TagName)
			}
		}
		if len(versions) > 0 {
			extensions[extensionName] = versions
		}
	}

	return extensions, nil
}

func (g *GithubReleaseFetcher) Get(url string, v any) error {
	body, err := cmdutil.NewSimpleRestClient().Get(url)
	if err != nil {
		return err
	}

	// Unmarshal the body
	err = json.Unmarshal(body, v)
	if err != nil {
		message := GithubMessage{}
		err = json.Unmarshal(body, &message)
		if err != nil {
			log.Printf("Failed to unmarshal JSON: %s", err)
			return err
		}

		return fmt.Errorf("failed to get %s: %s (%s)", url, message.Message, message.DocumentationUrl)
	}

	return nil
}
