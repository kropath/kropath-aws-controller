// Copyright 2026 kropath Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"github.com/kropath/kropath-aws-controller/internal/cascade"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type DSQLConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DSQLConfigSpec   `json:"spec,omitempty"`
	Status DSQLConfigStatus `json:"status,omitempty"`
}

type DSQLConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DSQLConfig `json:"items"`
}

type DSQLConfigSpec struct {
	Mandatory cascade.DSQLConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.DSQLConfigSection `json:"defaults,omitempty"`
}

type DSQLConfigStatus struct {
	EffectiveConfig    EffectiveDSQLConfig `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition  `json:"conditions,omitempty"`
	ObservedGeneration int64               `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string              `json:"syncedTimestamp,omitempty"`
}

type EffectiveDSQLConfig struct {
	AWS       ProviderIdentity          `json:"aws,omitempty"`
	Mandatory cascade.EffectiveDSQLSection `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveDSQLSection `json:"defaults,omitempty"`
}

func (in *DSQLConfig) DeepCopyInto(out *DSQLConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	if in.Spec.Mandatory.Tags != nil {
		out.Spec.Mandatory.Tags = make(map[string]string, len(in.Spec.Mandatory.Tags))
		for k, v := range in.Spec.Mandatory.Tags {
			out.Spec.Mandatory.Tags[k] = v
		}
	}
	if in.Spec.Mandatory.SyncedLabels != nil {
		out.Spec.Mandatory.SyncedLabels = make(map[string]string, len(in.Spec.Mandatory.SyncedLabels))
		for k, v := range in.Spec.Mandatory.SyncedLabels {
			out.Spec.Mandatory.SyncedLabels[k] = v
		}
	}
	if in.Spec.Mandatory.SyncedAnnotations != nil {
		out.Spec.Mandatory.SyncedAnnotations = make(map[string]string, len(in.Spec.Mandatory.SyncedAnnotations))
		for k, v := range in.Spec.Mandatory.SyncedAnnotations {
			out.Spec.Mandatory.SyncedAnnotations[k] = v
		}
	}
	if in.Spec.Defaults.Tags != nil {
		out.Spec.Defaults.Tags = make(map[string]string, len(in.Spec.Defaults.Tags))
		for k, v := range in.Spec.Defaults.Tags {
			out.Spec.Defaults.Tags[k] = v
		}
	}
	if in.Spec.Defaults.SyncedLabels != nil {
		out.Spec.Defaults.SyncedLabels = make(map[string]string, len(in.Spec.Defaults.SyncedLabels))
		for k, v := range in.Spec.Defaults.SyncedLabels {
			out.Spec.Defaults.SyncedLabels[k] = v
		}
	}
	if in.Spec.Defaults.SyncedAnnotations != nil {
		out.Spec.Defaults.SyncedAnnotations = make(map[string]string, len(in.Spec.Defaults.SyncedAnnotations))
		for k, v := range in.Spec.Defaults.SyncedAnnotations {
			out.Spec.Defaults.SyncedAnnotations[k] = v
		}
	}
	out.Status = in.Status
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
	if in.Status.EffectiveConfig.Mandatory.Tags != nil {
		out.Status.EffectiveConfig.Mandatory.Tags = make(map[string]string, len(in.Status.EffectiveConfig.Mandatory.Tags))
		for k, v := range in.Status.EffectiveConfig.Mandatory.Tags {
			out.Status.EffectiveConfig.Mandatory.Tags[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.Tags != nil {
		out.Status.EffectiveConfig.Defaults.Tags = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.Tags))
		for k, v := range in.Status.EffectiveConfig.Defaults.Tags {
			out.Status.EffectiveConfig.Defaults.Tags[k] = v
		}
	}
}

func (in *DSQLConfig) DeepCopy() *DSQLConfig {
	if in == nil {
		return nil
	}
	out := new(DSQLConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *DSQLConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *DSQLConfigList) DeepCopyInto(out *DSQLConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]DSQLConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *DSQLConfigList) DeepCopy() *DSQLConfigList {
	if in == nil {
		return nil
	}
	out := new(DSQLConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *DSQLConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s DSQLConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
