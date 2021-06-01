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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomTypeSpec defines the desired state of CustomType
type CustomTypeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Replicas int `json:"replicas"`
}

// CustomTypeStatus defines the observed state of CustomType
type CustomTypeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Replicas int      `json:"replicas"`
	PodNames []string `json:"podNames"` // 记录已经运行的 Pod 名字
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CustomType is the Schema for the customtypes API
type CustomType struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomTypeSpec   `json:"spec,omitempty"`
	Status CustomTypeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomTypeList contains a list of CustomType
type CustomTypeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomType `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomType{}, &CustomTypeList{})
}
