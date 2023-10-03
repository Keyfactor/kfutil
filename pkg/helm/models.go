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
