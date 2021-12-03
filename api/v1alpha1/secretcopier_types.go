/*
Copyright 2021.

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

// SecretCopierSpec defines the desired state of SecretCopier
type SecretCopierSpec struct {
	// +kubebuilder:validation:Required
	SecretLabel string `json:"secretLabel"`
}

type Phase string

const (
	RunningPhase Phase  = "RUNNING"
	PenndingPhase Phase	= "PENDING"
	ErrorPhase Phase	= "ERROR"
)
// SecretCopierStatus defines the observed state of SecretCopier
type SecretCopierStatus struct {
	Phase 			Phase 			`json:"phase"`
	
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SecretCopier is the Schema for the secretcopiers API
type SecretCopier struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretCopierSpec   `json:"spec,omitempty"`
	Status SecretCopierStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecretCopierList contains a list of SecretCopier
type SecretCopierList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretCopier `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretCopier{}, &SecretCopierList{})
}
