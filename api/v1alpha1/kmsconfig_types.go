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

type KMSConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KMSConfigSpec   `json:"spec,omitempty"`
	Status KMSConfigStatus `json:"status,omitempty"`
}

type KMSConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []KMSConfig `json:"items"`
}

type KMSConfigSpec struct {
	Mandatory cascade.KMSConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.KMSConfigSection `json:"defaults,omitempty"`
}

type KMSConfigStatus struct {
	EffectiveConfig    EffectiveKMSConfig `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string             `json:"syncedTimestamp,omitempty"`
}

type EffectiveKMSConfig struct {
	AWS       ProviderIdentity            `json:"aws,omitempty"`
	Mandatory cascade.EffectiveKMSSection `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveKMSSection `json:"defaults,omitempty"`
}

func (in *KMSConfig) DeepCopyInto(out *KMSConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	if in.Spec.Mandatory.AllowedKeySpecs != nil {
		out.Spec.Mandatory.AllowedKeySpecs = make([]string, len(in.Spec.Mandatory.AllowedKeySpecs))
		copy(out.Spec.Mandatory.AllowedKeySpecs, in.Spec.Mandatory.AllowedKeySpecs)
	}
	if in.Spec.Mandatory.AllowedGrantOperations != nil {
		out.Spec.Mandatory.AllowedGrantOperations = make([]string, len(in.Spec.Mandatory.AllowedGrantOperations))
		copy(out.Spec.Mandatory.AllowedGrantOperations, in.Spec.Mandatory.AllowedGrantOperations)
	}
	if in.Spec.Mandatory.Tags != nil {
		out.Spec.Mandatory.Tags = make(map[string]string, len(in.Spec.Mandatory.Tags))
		for k, v := range in.Spec.Mandatory.Tags {
			out.Spec.Mandatory.Tags[k] = v
		}
	}
	if in.Spec.Defaults.AllowedKeySpecs != nil {
		out.Spec.Defaults.AllowedKeySpecs = make([]string, len(in.Spec.Defaults.AllowedKeySpecs))
		copy(out.Spec.Defaults.AllowedKeySpecs, in.Spec.Defaults.AllowedKeySpecs)
	}
	if in.Spec.Defaults.AllowedGrantOperations != nil {
		out.Spec.Defaults.AllowedGrantOperations = make([]string, len(in.Spec.Defaults.AllowedGrantOperations))
		copy(out.Spec.Defaults.AllowedGrantOperations, in.Spec.Defaults.AllowedGrantOperations)
	}
	if in.Spec.Defaults.Tags != nil {
		out.Spec.Defaults.Tags = make(map[string]string, len(in.Spec.Defaults.Tags))
		for k, v := range in.Spec.Defaults.Tags {
			out.Spec.Defaults.Tags[k] = v
		}
	}
	out.Status = in.Status
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
	if in.Status.EffectiveConfig.Mandatory.AllowedKeySpecs != nil {
		out.Status.EffectiveConfig.Mandatory.AllowedKeySpecs = make([]string, len(in.Status.EffectiveConfig.Mandatory.AllowedKeySpecs))
		copy(out.Status.EffectiveConfig.Mandatory.AllowedKeySpecs, in.Status.EffectiveConfig.Mandatory.AllowedKeySpecs)
	}
	if in.Status.EffectiveConfig.Mandatory.AllowedGrantOperations != nil {
		out.Status.EffectiveConfig.Mandatory.AllowedGrantOperations = make([]string, len(in.Status.EffectiveConfig.Mandatory.AllowedGrantOperations))
		copy(out.Status.EffectiveConfig.Mandatory.AllowedGrantOperations, in.Status.EffectiveConfig.Mandatory.AllowedGrantOperations)
	}
	if in.Status.EffectiveConfig.Mandatory.Tags != nil {
		out.Status.EffectiveConfig.Mandatory.Tags = make(map[string]string, len(in.Status.EffectiveConfig.Mandatory.Tags))
		for k, v := range in.Status.EffectiveConfig.Mandatory.Tags {
			out.Status.EffectiveConfig.Mandatory.Tags[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.AllowedKeySpecs != nil {
		out.Status.EffectiveConfig.Defaults.AllowedKeySpecs = make([]string, len(in.Status.EffectiveConfig.Defaults.AllowedKeySpecs))
		copy(out.Status.EffectiveConfig.Defaults.AllowedKeySpecs, in.Status.EffectiveConfig.Defaults.AllowedKeySpecs)
	}
	if in.Status.EffectiveConfig.Defaults.AllowedGrantOperations != nil {
		out.Status.EffectiveConfig.Defaults.AllowedGrantOperations = make([]string, len(in.Status.EffectiveConfig.Defaults.AllowedGrantOperations))
		copy(out.Status.EffectiveConfig.Defaults.AllowedGrantOperations, in.Status.EffectiveConfig.Defaults.AllowedGrantOperations)
	}
	if in.Status.EffectiveConfig.Defaults.Tags != nil {
		out.Status.EffectiveConfig.Defaults.Tags = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.Tags))
		for k, v := range in.Status.EffectiveConfig.Defaults.Tags {
			out.Status.EffectiveConfig.Defaults.Tags[k] = v
		}
	}
}

func (in *KMSConfig) DeepCopy() *KMSConfig {
	if in == nil {
		return nil
	}
	out := new(KMSConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *KMSConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *KMSConfigList) DeepCopyInto(out *KMSConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]KMSConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *KMSConfigList) DeepCopy() *KMSConfigList {
	if in == nil {
		return nil
	}
	out := new(KMSConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *KMSConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s KMSConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
