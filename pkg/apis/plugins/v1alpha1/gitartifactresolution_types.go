/*
Copyright 2019 Miles Bryant.

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

package v1alpha1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GitPipelineInputConfig struct {
	Repository string `json:"repository,omitempty"`
}

func GitPipelineInputConfigFromJSON(jsonStr string) (*GitPipelineInputConfig, error) {
	g := &GitPipelineInputConfig{}
	err := json.Unmarshal([]byte(jsonStr), g)
	return g, err
}

type GitPipelineInstanceInputConfig struct {
	Commit string `json:"commit,omitempty"`
}

func GitPipelineInstanceInputConfigFromJSON(jsonStr string) (*GitPipelineInstanceInputConfig, error) {
	g := &GitPipelineInstanceInputConfig{}
	err := json.Unmarshal([]byte(jsonStr), g)
	return g, err
}

// GitArtifactResolutionSpec defines the desired state of GitArtifactResolution
type GitArtifactResolutionSpec struct {
	RepositoryURL string `json:"repository_url,omitempty"`
	CommitSHA     string `json:"commit_sha,omitempty"`
}

func (g GitArtifactResolutionSpec) ToJSON() string {
	bytes, err := json.Marshal(g)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

// GitArtifactResolutionStatus defines the observed state of GitArtifactResolution
type GitArtifactResolutionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitArtifactResolution is the Schema for the gitartifactresolutions API
// +k8s:openapi-gen=true
type GitArtifactResolution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *GitArtifactResolutionSpec  `json:"spec,omitempty"`
	Status GitArtifactResolutionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitArtifactResolutionList contains a list of GitArtifactResolution
type GitArtifactResolutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitArtifactResolution `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitArtifactResolution{}, &GitArtifactResolutionList{})
}