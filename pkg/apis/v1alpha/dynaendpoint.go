// Copyright (C) 2018 Manabu Sonoda.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	KindDynaEndpoint     = "DynaEndpoint"
	KindDynaEndpointList = "DynaEndpointList"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DynaEndpoint is a specification for a DynaEndpoint resource
type DynaEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status DynaEndpointStatus `json:"status"`
	Spec   DynaEndpointSpec   `json:"spec"`
}

type DynaEndpointStatus struct {
	Name string `json:"name"`
}

// DynaEndpointSpec is the spec for a Workflow resource
type DynaEndpointSpec struct {
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DynaEndpointList is a list of Workflow resources
type DynaEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DynaEndpoint `json:"items"`
}

func NewDynaEndpoint() *DynaEndpoint {
	return &DynaEndpoint{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDynaEndpoint,
			APIVersion: GroupVersionCurrent,
		},
	}
}

func NewDynaEndpointList() *DynaEndpointList {
	return &DynaEndpointList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDynaEndpointList,
			APIVersion: GroupVersionCurrent,
		},
	}
}
