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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ArtifactPhase string

const (
	InvalidArtifact    ArtifactPhase = "Invalid"
	UnresolvedArtifact ArtifactPhase = "Unresolved"
	ResolvedArtifact   ArtifactPhase = "Resolved"
)

type StorageResponse struct {
	Status               string `json:"status"`
	GroupVersionResource string `json:"group_version_resource"`
	Reference            string `json:"reference"`
}

type ArtifactSource struct {
	Type   string `json:"type,omitempty"`
	Config string `json:"config,omitempty"`
}

// ArtifactSpec defines the desired state of Artifact
type ArtifactSpec struct {
	Source               ArtifactSource `json:"source,omitempty"`
	GroupVersionResource string         `json:"group_version_resource,omitempty"`
	Reference            string         `json:"reference,omitempty"`
}

// ArtifactStatus defines the observed state of Artifact
type ArtifactStatus struct {
	Phase ArtifactPhase `json:"phase,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Artifact is the Schema for the artifacts API
// +k8s:openapi-gen=true
type Artifact struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArtifactSpec   `json:"spec,omitempty"`
	Status ArtifactStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArtifactList contains a list of Artifact
type ArtifactList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Artifact `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Artifact{}, &ArtifactList{})
}
