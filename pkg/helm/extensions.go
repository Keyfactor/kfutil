package helm

import (
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"io"
	"net/http"
	"strings"
)

func (b *UniversalOrchestratorHelmValueBuilder) selectExtensionsHandler() error {
	ghTool := NewGithubReleaseFetcher(b.token)
	extensions, err := ghTool.GetExtensionList()
	if err != nil {
		return err
	}

	// Get a list of currently installed extensions
	installedExtensions := make(Extensions)
	for _, container := range b.values.InitContainers {
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
		for i, container := range b.values.InitContainers {
			if container.Name == initContainerName {
				extensionAlreadyInstalled = true
				for j, env := range container.Env {
					if env.Name == "EXTENSION_VERSION" && env.Value != latestVersion {
						fmt.Printf("Upgrading %s from %s to %s\n", extension, env.Value, latestVersion)
						b.values.InitContainers[i].Env[j].Value = latestVersion
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

		b.values.InitContainers = append(b.values.InitContainers, InitContainer{
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
					Name:      b.values.ExtensionStorage.Name,
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
	page := 1

	for {
		// Prevent rate limiting
		if page > 10 {
			break
		}

		// Ask https://api.github.com/orgs/keyfactor/repos for the list of repos
		// Unmarshal the body into a slice of gitHubRepo structs
		var repos []gitHubRepo
		err := g.Get(fmt.Sprintf("https://api.github.com/orgs/keyfactor/repos?page=%d&per_page=100", page), &repos)
		if err != nil {
			return nil, err
		}

		// If the length of the repos slice is 0, we've reached the end of the list
		if len(repos) == 0 {
			break
		}

		// Loop through the repos and add them to the orchestratorList slice
		for _, repo := range repos {
			if strings.Contains(repo.Name, "-orchestrator") {
				orchestratorList = append(orchestratorList, repo.Name)
			}
		}

		page++
	}

	return orchestratorList, nil
}

func (g *GithubReleaseFetcher) GetExtensionList() (Extensions, error) {
	extensionNameList, err := g.getOrchestratorNames()
	if err != nil {
		return nil, err
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
	client := http.DefaultClient

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Add Authorization header
	if g.token != "" {
		req.Header.Set("Authorization", "Bearer "+g.token)
	}

	get, err := client.Do(req)
	if err != nil {
		return err
	}

	// Read the body of the response
	body, err := io.ReadAll(get.Body)
	if err != nil {
		return err
	}

	// Unmarshal the body
	// TODO this could panic if the body is not valid JSON
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}

	return nil
}
