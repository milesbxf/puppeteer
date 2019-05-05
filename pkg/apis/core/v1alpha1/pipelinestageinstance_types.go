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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PipelineStageInstanceSpec defines the desired state of PipelineStageInstance
type PipelineStageInstanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

type PipelineStageInstancePhase string

const (
	PipelineStageInstanceInProgress PipelineStageInstancePhase = "InProgress"
	PipelineStageInstanceError      PipelineStageInstancePhase = "Error"
	PipelineStageInstanceComplete   PipelineStageInstancePhase = "Complete"
)

// PipelineStageInstanceStatus defines the observed state of PipelineStageInstance
type PipelineStageInstanceStatus struct {
	Phase PipelineStageInstancePhase `json:"phase,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PipelineStageInstance is the Schema for the pipelinestageinstances API
// +k8s:openapi-gen=true
type PipelineStageInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineStageInstanceSpec   `json:"spec,omitempty"`
	Status PipelineStageInstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PipelineStageInstanceList contains a list of PipelineStageInstance
type PipelineStageInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PipelineStageInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PipelineStageInstance{}, &PipelineStageInstanceList{})
}
