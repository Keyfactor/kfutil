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

package extensions

import (
	"encoding/json"
	"fmt"
	"kfutil/pkg/cmdutil"
	"log"
	"strings"
	"time"
)

const defaultGithubOrg = "keyfactor"

type Version string
type ExtensionName string
type Extensions map[ExtensionName]Version
type Extension struct {
	Name    ExtensionName
	Version Version
}

type GithubReleaseFetcher struct {
	token string
	org   string
}

func NewGithubReleaseFetcher(org, token string) *GithubReleaseFetcher {
	if org == "" {
		org = defaultGithubOrg
	}

	return &GithubReleaseFetcher{
		org:   org,
		token: token,
	}
}

var firstExtensionCache ExtensionName

func (g *GithubReleaseFetcher) GetFirstExtension() (ExtensionName, error) {
	if firstExtensionCache != "" {
		return firstExtensionCache, nil
	}

	var name ExtensionName

	// Ask https://api.github.com/orgs/keyfactor/repos for the list of repos
	// Unmarshal the body into a slice of GithubRepo structs
	var repos []GithubRepo
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?type=public&page=1&per_page=100", g.org)
	err := g.Get(url, &repos)
	if err != nil {
		return "", err
	}

	// Loop through the repos and add them to the orchestratorList slice
	for _, repo := range repos {
		// If the repo doesn't end with "-orchestrator" or "-pam, skip it
		if !strings.HasSuffix(repo.Name, "-orchestrator") && !strings.HasSuffix(repo.Name, "-pam") {
			continue
		}

		versions, err := g.GetExtensionVersions(ExtensionName(repo.Name))
		if err != nil {
			return "", err
		}

		if len(versions) > 0 {
			name = ExtensionName(repo.Name)
			break
		}
	}

	firstExtensionCache = name

	return name, nil
}

func (g *GithubReleaseFetcher) GetExtensionNames() ([]ExtensionName, error) {
	orchestratorList := make([]ExtensionName, 0)

	for page := 1; page < 100; page++ {
		// Prevent rate limiting by setting upper bound to 100

		// Ask https://api.github.com/orgs/keyfactor/repos for the list of repos
		// Unmarshal the body into a slice of GithubRepo structs
		var repos []GithubRepo
		url := fmt.Sprintf("https://api.github.com/orgs/%s/repos?type=public&page=%d&per_page=100", g.org, page)
		err := g.Get(url, &repos)
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
				orchestratorList = append(orchestratorList, ExtensionName(repo.Name))
			}
		}
	}

	return orchestratorList, nil
}

func (g *GithubReleaseFetcher) GetExtensionVersions(extension ExtensionName) ([]Version, error) {
	// Ask https://api.github.com/repos/keyfactor/{name}/releases for the list of releases
	// Unmarshal the body into a slice of GithubRelease structs
	var releases []GithubRelease
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", g.org, extension)
	err := g.Get(url, &releases)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of releases: %s", err)
	}

	// Add the extension to the list of extensions
	// TODO order the list of extensions by version release date just in case Github changes in the future
	versions := make([]Version, 0)
	for _, release := range releases {
		if !release.Prerelease {
			versions = append(versions, release.TagName)
		}
	}

	return versions, nil
}

func (g *GithubReleaseFetcher) GetExtensionList() (Extensions, error) {
	extensionNameList, err := g.GetExtensionNames()
	if err != nil {
		return nil, fmt.Errorf("failed to get list of extensions: %s", err)
	}

	extensions := make(Extensions)

	for _, extensionName := range extensionNameList {
		versions, err := g.GetExtensionVersions(extensionName)
		if err != nil {
			return nil, err
		}

		if len(versions) > 0 {
			extensions[extensionName] = versions[0]
		}
	}

	return extensions, nil
}

func (g *GithubReleaseFetcher) ExtensionExists(name ExtensionName, version Version) (bool, error) {
	versions, err := g.GetExtensionVersions(name)
	if err != nil {
		return false, fmt.Errorf("failed to get list of versions for extension %q: %s", name, err)
	}

	for _, v := range versions {
		if v == version {
			return true, nil
		}
	}

	return false, nil
}

func (g *GithubReleaseFetcher) DownloadExtension(name ExtensionName, version Version) (*[]byte, error) {
	// Construct URL
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s_%s.zip", g.org, name, version, name, version)

	// Download the zip file
	rest := cmdutil.NewSimpleRestClient()
	rest.SetBearerToken(g.token)
	body, err := rest.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download extension %s:%s: %s", name, version, err)
	}

	return &body, nil
}

func (g *GithubReleaseFetcher) Get(url string, v any) error {
	rest := cmdutil.NewSimpleRestClient()
	rest.SetBearerToken(g.token)
	body, err := rest.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get %s: %s", url, err)
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

type GithubRepo struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
}

type GithubRelease struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         Version   `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string `json:"url"`
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}

type GithubMessage struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url"`
}
