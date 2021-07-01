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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RyanSpec defines the desired state of Ryan
type RyanSpec struct {
	// ResourceName is the name of the resource to watch
	ResourceName string `json:"resourceName"`

	// API version of the resource to watch, reserved for future when we support more than just deployment
	//+kubebuilder:default="apps/v1"
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the resource to watch, reserved for future when we support more than just deployment
	//+kubebuilder:default="Deployment"
	Kind string `json:"kind,omitempty"`
}

// RyanStatus defines the observed state of Ryan
type RyanStatus struct {
	// ReleaseName is the name of the helm release
	ReleaseName string `json:"releaseName"`

	// AppName is the helm app name
	AppName string `json:"appName"`

	// AppVersion is the helm app version
	AppVersion string `json:"appVersion"`
}

// Ryan is the Schema for the ryans API
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="AppVersion",type=string,JSONPath=`.status.appVersion`
type Ryan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RyanSpec   `json:"spec,omitempty"`
	Status RyanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RyanList contains a list of Ryan
type RyanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ryan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Ryan{}, &RyanList{})
}
