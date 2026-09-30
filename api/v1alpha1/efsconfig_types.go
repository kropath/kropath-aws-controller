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

type EFSConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EFSConfigSpec   `json:"spec,omitempty"`
	Status EFSConfigStatus `json:"status,omitempty"`
}

type EFSConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []EFSConfig `json:"items"`
}

type EFSConfigSpec struct {
	Mandatory cascade.EFSConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.EFSConfigSection `json:"defaults,omitempty"`
}

type EFSConfigStatus struct {
	EffectiveConfig    EffectiveEFSConfig `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string             `json:"syncedTimestamp,omitempty"`
}

type EffectiveEFSConfig struct {
	AWS       ProviderIdentity              `json:"aws,omitempty"`
	Mandatory cascade.EffectiveEFSSection   `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveEFSSection   `json:"defaults,omitempty"`
}

func deepCopyEFSConfigSection(dst *cascade.EFSConfigSection, src *cascade.EFSConfigSection) {
	if src.Encrypted != nil {
		b := *src.Encrypted
		dst.Encrypted = &b
	}
	if src.BackupEnabled != nil {
		b := *src.BackupEnabled
		dst.BackupEnabled = &b
	}
	if src.Tags != nil {
		dst.Tags = make(map[string]string, len(src.Tags))
		for k, v := range src.Tags {
			dst.Tags[k] = v
		}
	}
	if src.SyncedLabels != nil {
		dst.SyncedLabels = make(map[string]string, len(src.SyncedLabels))
		for k, v := range src.SyncedLabels {
			dst.SyncedLabels[k] = v
		}
	}
	if src.SyncedAnnotations != nil {
		dst.SyncedAnnotations = make(map[string]string, len(src.SyncedAnnotations))
		for k, v := range src.SyncedAnnotations {
			dst.SyncedAnnotations[k] = v
		}
	}
}

func deepCopyEffectiveEFSSection(dst *cascade.EffectiveEFSSection, src *cascade.EffectiveEFSSection) {
	if src.Encrypted != nil {
		b := *src.Encrypted
		dst.Encrypted = &b
	}
	if src.BackupEnabled != nil {
		b := *src.BackupEnabled
		dst.BackupEnabled = &b
	}
	if src.Tags != nil {
		dst.Tags = make(map[string]string, len(src.Tags))
		for k, v := range src.Tags {
			dst.Tags[k] = v
		}
	}
	if src.SyncedLabels != nil {
		dst.SyncedLabels = make(map[string]string, len(src.SyncedLabels))
		for k, v := range src.SyncedLabels {
			dst.SyncedLabels[k] = v
		}
	}
	if src.SyncedAnnotations != nil {
		dst.SyncedAnnotations = make(map[string]string, len(src.SyncedAnnotations))
		for k, v := range src.SyncedAnnotations {
			dst.SyncedAnnotations[k] = v
		}
	}
}

func (in *EFSConfig) DeepCopyInto(out *EFSConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	deepCopyEFSConfigSection(&out.Spec.Mandatory, &in.Spec.Mandatory)
	deepCopyEFSConfigSection(&out.Spec.Defaults, &in.Spec.Defaults)
	out.Status = in.Status
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
	deepCopyEffectiveEFSSection(&out.Status.EffectiveConfig.Mandatory, &in.Status.EffectiveConfig.Mandatory)
	deepCopyEffectiveEFSSection(&out.Status.EffectiveConfig.Defaults, &in.Status.EffectiveConfig.Defaults)
}

func (in *EFSConfig) DeepCopy() *EFSConfig {
	if in == nil {
		return nil
	}
	out := new(EFSConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *EFSConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *EFSConfigList) DeepCopyInto(out *EFSConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]EFSConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *EFSConfigList) DeepCopy() *EFSConfigList {
	if in == nil {
		return nil
	}
	out := new(EFSConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *EFSConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s EFSConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
