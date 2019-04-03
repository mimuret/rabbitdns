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
	KindDynaRule     = "DynaRule"
	KindDynaRuleList = "DynaRuleList"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DynaRule is a specification for a DynaRule resource
type DynaRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status DynaRuleStatus `json:"status"`
	Spec   DynaRuleSpec   `json:"spec"`
}

type DynaRuleStatus struct {
	Name string `json:"name"`
}

// DynaRuleSpec is the spec for a Workflow resource
type DynaRuleSpec struct {
	Name string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DynaRuleList is a list of Workflow resources
type DynaRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DynaRule `json:"items"`
}

func NewDynaRule() *DynaRule {
	return &DynaRule{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDynaRule,
			APIVersion: GroupVersionCurrent,
		},
	}
}

func NewDynaRuleList() *DynaRuleList {
	return &DynaRuleList{
		TypeMeta: metav1.TypeMeta{
			Kind:       KindDynaRuleList,
			APIVersion: GroupVersionCurrent,
		},
	}
}
