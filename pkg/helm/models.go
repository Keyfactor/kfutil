package helm

import "time"

type InitContainer struct {
	Name            string        `yaml:"name"`
	Image           string        `yaml:"image"`
	ImagePullPolicy string        `yaml:"imagePullPolicy"`
	Env             []Environment `yaml:"env"`
	VolumeMounts    []VolumeMount `yaml:"volumeMounts"`
}

type Environment struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
	SubPath   string `yaml:"subPath"`
	ReadOnly  bool   `yaml:"readOnly"`
}

type Volume struct {
	ConfigMap struct {
		Items []ConfigMapItem `yaml:"items"`
		Name  string          `yaml:"name"`
	} `yaml:"configMap"`
	Name string `yaml:"name"`
}

type ConfigMapItem struct {
	Key  string `yaml:"key"`
	Path string `yaml:"path"`
}

type UniversalOrchestratorHelmValues struct {
	BaseOrchestratorName string `yaml:"baseOrchestratorName"`
	CompleteName         string `yaml:"completeName"`
	ReplicaCount         int    `yaml:"replicaCount"`
	Image                struct {
		Repository string `yaml:"repository"`
		PullPolicy string `yaml:"pullPolicy"`
		Tag        string `yaml:"tag"`
	} `yaml:"image"`
	CommandAgentURL string `yaml:"commandAgentUrl"`
	Auth            struct {
		SecretName             string `yaml:"secretName"`
		UseOauthAuthentication bool   `yaml:"useOauthAuthentication"`
	} `yaml:"auth"`
	LogLevel         string          `yaml:"logLevel"`
	ImagePullSecrets []interface{}   `yaml:"imagePullSecrets"`
	InitContainers   []InitContainer `yaml:"initContainers"`
	ServiceAccount   struct {
		Create      bool   `yaml:"create"`
		Name        string `yaml:"name"`
		Annotations struct {
		} `yaml:"annotations"`
	} `yaml:"serviceAccount"`
	ExtensionStorage struct {
		Name         string `yaml:"name"`
		StorageClass string `yaml:"storageClass"`
		Size         string `yaml:"size"`
		AccessMode   string `yaml:"accessMode"`
		Annotations  struct {
		} `yaml:"annotations"`
	} `yaml:"extensionStorage"`
	Volumes         []Volume      `yaml:"volumes"`
	VolumeMounts    []VolumeMount `yaml:"volumeMounts"`
	ExtraContainers interface{}   `yaml:"extraContainers"`
	PostStart       []string      `yaml:"postStart"`
}

type gitHubRepo struct {
	ID          int    `json:"id"`
	NodeID      string `json:"node_id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
}

type gitHubRelease struct {
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
	TagName         string    `json:"tag_name"`
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
