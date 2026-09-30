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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kropath/kropath-aws-controller/internal/cascade"
)

type ManagedPrometheusConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedPrometheusConfigSpec   `json:"spec,omitempty"`
	Status ManagedPrometheusConfigStatus `json:"status,omitempty"`
}

type ManagedPrometheusConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ManagedPrometheusConfig `json:"items"`
}

type ManagedPrometheusConfigSpec struct {
	Mandatory cascade.ManagedPrometheusConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.ManagedPrometheusConfigSection `json:"defaults,omitempty"`
}

type ManagedPrometheusConfigStatus struct {
	EffectiveConfig    EffectiveManagedPrometheusConfig `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition               `json:"conditions,omitempty"`
	ObservedGeneration int64                            `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string                           `json:"syncedTimestamp,omitempty"`
}

type EffectiveManagedPrometheusConfig struct {
	AWS       ProviderIdentity                         `json:"aws,omitempty"`
	Mandatory cascade.EffectiveManagedPrometheusSection `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveManagedPrometheusSection `json:"defaults,omitempty"`
}

func (in *ManagedPrometheusConfig) DeepCopyInto(out *ManagedPrometheusConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *ManagedPrometheusConfig) DeepCopy() *ManagedPrometheusConfig {
	if in == nil {
		return nil
	}
	out := new(ManagedPrometheusConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *ManagedPrometheusConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ManagedPrometheusConfigList) DeepCopyInto(out *ManagedPrometheusConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]ManagedPrometheusConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *ManagedPrometheusConfigList) DeepCopy() *ManagedPrometheusConfigList {
	if in == nil {
		return nil
	}
	out := new(ManagedPrometheusConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *ManagedPrometheusConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ManagedPrometheusConfigSpec) DeepCopyInto(out *ManagedPrometheusConfigSpec) {
	*out = *in
	if in.Mandatory.SyncedLabels != nil {
		out.Mandatory.SyncedLabels = make(map[string]string, len(in.Mandatory.SyncedLabels))
		for k, v := range in.Mandatory.SyncedLabels {
			out.Mandatory.SyncedLabels[k] = v
		}
	}
	if in.Mandatory.SyncedAnnotations != nil {
		out.Mandatory.SyncedAnnotations = make(map[string]string, len(in.Mandatory.SyncedAnnotations))
		for k, v := range in.Mandatory.SyncedAnnotations {
			out.Mandatory.SyncedAnnotations[k] = v
		}
	}
	if in.Mandatory.Tags != nil {
		out.Mandatory.Tags = make(map[string]string, len(in.Mandatory.Tags))
		for k, v := range in.Mandatory.Tags {
			out.Mandatory.Tags[k] = v
		}
	}
	if in.Defaults.SyncedLabels != nil {
		out.Defaults.SyncedLabels = make(map[string]string, len(in.Defaults.SyncedLabels))
		for k, v := range in.Defaults.SyncedLabels {
			out.Defaults.SyncedLabels[k] = v
		}
	}
	if in.Defaults.SyncedAnnotations != nil {
		out.Defaults.SyncedAnnotations = make(map[string]string, len(in.Defaults.SyncedAnnotations))
		for k, v := range in.Defaults.SyncedAnnotations {
			out.Defaults.SyncedAnnotations[k] = v
		}
	}
	if in.Defaults.Tags != nil {
		out.Defaults.Tags = make(map[string]string, len(in.Defaults.Tags))
		for k, v := range in.Defaults.Tags {
			out.Defaults.Tags[k] = v
		}
	}
}

func (in *ManagedPrometheusConfigStatus) DeepCopyInto(out *ManagedPrometheusConfigStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		copy(out.Conditions, in.Conditions)
	}
	if in.EffectiveConfig.Mandatory.SyncedLabels != nil {
		out.EffectiveConfig.Mandatory.SyncedLabels = make(map[string]string, len(in.EffectiveConfig.Mandatory.SyncedLabels))
		for k, v := range in.EffectiveConfig.Mandatory.SyncedLabels {
			out.EffectiveConfig.Mandatory.SyncedLabels[k] = v
		}
	}
	if in.EffectiveConfig.Mandatory.SyncedAnnotations != nil {
		out.EffectiveConfig.Mandatory.SyncedAnnotations = make(map[string]string, len(in.EffectiveConfig.Mandatory.SyncedAnnotations))
		for k, v := range in.EffectiveConfig.Mandatory.SyncedAnnotations {
			out.EffectiveConfig.Mandatory.SyncedAnnotations[k] = v
		}
	}
	if in.EffectiveConfig.Mandatory.Tags != nil {
		out.EffectiveConfig.Mandatory.Tags = make(map[string]string, len(in.EffectiveConfig.Mandatory.Tags))
		for k, v := range in.EffectiveConfig.Mandatory.Tags {
			out.EffectiveConfig.Mandatory.Tags[k] = v
		}
	}
	if in.EffectiveConfig.Defaults.SyncedLabels != nil {
		out.EffectiveConfig.Defaults.SyncedLabels = make(map[string]string, len(in.EffectiveConfig.Defaults.SyncedLabels))
		for k, v := range in.EffectiveConfig.Defaults.SyncedLabels {
			out.EffectiveConfig.Defaults.SyncedLabels[k] = v
		}
	}
	if in.EffectiveConfig.Defaults.SyncedAnnotations != nil {
		out.EffectiveConfig.Defaults.SyncedAnnotations = make(map[string]string, len(in.EffectiveConfig.Defaults.SyncedAnnotations))
		for k, v := range in.EffectiveConfig.Defaults.SyncedAnnotations {
			out.EffectiveConfig.Defaults.SyncedAnnotations[k] = v
		}
	}
	if in.EffectiveConfig.Defaults.Tags != nil {
		out.EffectiveConfig.Defaults.Tags = make(map[string]string, len(in.EffectiveConfig.Defaults.Tags))
		for k, v := range in.EffectiveConfig.Defaults.Tags {
			out.EffectiveConfig.Defaults.Tags[k] = v
		}
	}
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s ManagedPrometheusConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
