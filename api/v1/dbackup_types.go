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

package v1

import (
	kubebatchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DbackupSpec defines the desired state of Dbackup
type DbackupSpec struct {
	//+kubebuilder:validation:MinLength=0
	Schedule string `json:"schedule"`

	//+kubebuilder:validation:MinLength=0
	DatabaseTarget string `json:"DatabaseTarget"`

	// v1 will support only AWS
	//+kubebuilder:validation:Enum=AWS
	CloudProvider string `json:"cloudProvider"`

	//+kubebuilder:validation:MinLength=0
	BucketEndpoint string `json:"bucketEndpoint"`

	// v1 will only support Postgres
	//+kubebuilder:validation:Enum=Postgres
	DatabaseType string `json:"databaseType"`

	BackupTemplate kubebatchv1.JobTemplateSpec `json:"backupTemplate"`

	// +optional
	ConcurrencyPolicy Policy `json:"concurrencyPolicy,omitempty"`
}

// +kubebuilder:validation:Enum=Allow;Forbid;Replace
type Policy string

const (
	// Allow job to run togther
	Allow Policy = "Allow"

	// Forbid jobs to run togther
	Forbid Policy = "Forbid"

	// ReplaceConcurrent cancels currently running job and replaces it with a new one.
	Replace Policy = "Replace"
)

// DbackupStatus defines the observed state of Dbackup
type DbackupStatus struct {
	// +optional
	Active []corev1.ObjectReference `json:"active,omitempty"`

	// +optional
	LastBackupTime *metav1.Time `json:"lastScheduleTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Dbackup is the Schema for the dbackups API
type Dbackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DbackupSpec   `json:"spec,omitempty"`
	Status DbackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DbackupList contains a list of Dbackup
type DbackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dbackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dbackup{}, &DbackupList{})
}
