/*
Copyright 2025.

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

// SimpleapiSpec defines the desired state of Simpleapi
type SimpleapiSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Simpleapi. Edit simpleapi_types.go to remove/update
	Image    string `json:"image"`
	Version  string `json:"version"`
	Port     int32  `json:"port"`
	Replicas *int32 `json:"replicas"`
}

// SimpleapiStatus defines the observed state of Simpleapi
type SimpleapiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Simpleapi is the Schema for the simpleapis API
type Simpleapi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimpleapiSpec   `json:"spec,omitempty"`
	Status SimpleapiStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SimpleapiList contains a list of Simpleapi
type SimpleapiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Simpleapi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Simpleapi{}, &SimpleapiList{})
}
