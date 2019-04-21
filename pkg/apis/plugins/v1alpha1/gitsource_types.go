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

type GitRepositoryOptions struct {
	URL    string `json:"url,omitempty"`
	Branch string `json:"branch,omitempty"`
}

type GitCloneOptions struct {
	Shallow bool `json:"shallow,omitempty"`
}

type PollOptions struct {
	IntervalMinutes *int32 `json:"interval_minutes,omitempty"`
}

// GitSourceSpec defines the desired state of GitSource
type GitSourceSpec struct {
	Repository GitRepositoryOptions `json:"repository,omitempty"`
	Clone      GitCloneOptions      `json:"clone,omitempty"`
	Poll       PollOptions          `json:"poll,omitempty"`
}

// GitSourceStatus defines the observed state of GitSource
type GitSourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitSource is the Schema for the gitsources API
// +k8s:openapi-gen=true
type GitSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitSourceSpec   `json:"spec,omitempty"`
	Status GitSourceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GitSourceList contains a list of GitSource
type GitSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitSource{}, &GitSourceList{})
}
