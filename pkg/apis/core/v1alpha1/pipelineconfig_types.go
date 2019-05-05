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

type PipelineInput struct {
	Type   string `json:"type,omitempty"`
	Config string `json:"config,omitempty"`
}

type TaskInput struct {
	From string `json:"from,omitempty"`
	Path string `json:"path,omitempty"`
}

type TaskOutput struct {
	Type   string `json:"type,omitempty"`
	Config string `json:"config,omitempty"`
}

type PipelineTask struct {
	Image   string                `json:"image,omitempty"`
	Inputs  map[string]TaskInput  `json:"inputs,omitempty"`
	Shell   string                `json:"shell,omitempty"`
	Outputs map[string]TaskOutput `json:"outputs,omitempty"`
}

type PipelineStage struct {
	Name  string                  `json:"name,omitempty"`
	Tasks map[string]PipelineTask `json:"tasks,omitempty"`
}

type Workflow struct {
	Inputs map[string]PipelineInput `json:"inputs,omitempty"`
	Stages []PipelineStage          `json:"stages,omitempty"`
}

// PipelineConfigSpec defines the desired state of PipelineConfig
type PipelineConfigSpec struct {
	Workflow `json:"workflow,omitempty"`
}

// PipelineConfigStatus defines the observed state of PipelineConfig
type PipelineConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PipelineConfig is the Schema for the pipelineconfigs API
// +k8s:openapi-gen=true
type PipelineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineConfigSpec   `json:"spec,omitempty"`
	Status PipelineConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PipelineConfigList contains a list of PipelineConfig
type PipelineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelineConfig{}, &PipelineConfigList{})
}
