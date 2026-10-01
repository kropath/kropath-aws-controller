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

type BackupConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupConfigSpec   `json:"spec,omitempty"`
	Status BackupConfigStatus `json:"status,omitempty"`
}

type BackupConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []BackupConfig `json:"items"`
}

type BackupConfigSpec struct {
	Mandatory cascade.BackupConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.BackupConfigSection `json:"defaults,omitempty"`
}

type BackupConfigStatus struct {
	EffectiveConfig    EffectiveBackupConfig `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition    `json:"conditions,omitempty"`
	ObservedGeneration int64                 `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string                `json:"syncedTimestamp,omitempty"`
}

type EffectiveBackupConfig struct {
	AWS       ProviderIdentity               `json:"aws,omitempty"`
	Mandatory cascade.EffectiveBackupSection `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveBackupSection `json:"defaults,omitempty"`
}

func (in *BackupConfig) DeepCopyInto(out *BackupConfig) {
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
	if in.Status.EffectiveConfig.Mandatory.SyncedLabels != nil {
		out.Status.EffectiveConfig.Mandatory.SyncedLabels = make(map[string]string, len(in.Status.EffectiveConfig.Mandatory.SyncedLabels))
		for k, v := range in.Status.EffectiveConfig.Mandatory.SyncedLabels {
			out.Status.EffectiveConfig.Mandatory.SyncedLabels[k] = v
		}
	}
	if in.Status.EffectiveConfig.Mandatory.SyncedAnnotations != nil {
		out.Status.EffectiveConfig.Mandatory.SyncedAnnotations = make(map[string]string, len(in.Status.EffectiveConfig.Mandatory.SyncedAnnotations))
		for k, v := range in.Status.EffectiveConfig.Mandatory.SyncedAnnotations {
			out.Status.EffectiveConfig.Mandatory.SyncedAnnotations[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.Tags != nil {
		out.Status.EffectiveConfig.Defaults.Tags = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.Tags))
		for k, v := range in.Status.EffectiveConfig.Defaults.Tags {
			out.Status.EffectiveConfig.Defaults.Tags[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.SyncedLabels != nil {
		out.Status.EffectiveConfig.Defaults.SyncedLabels = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.SyncedLabels))
		for k, v := range in.Status.EffectiveConfig.Defaults.SyncedLabels {
			out.Status.EffectiveConfig.Defaults.SyncedLabels[k] = v
		}
	}
	if in.Status.EffectiveConfig.Defaults.SyncedAnnotations != nil {
		out.Status.EffectiveConfig.Defaults.SyncedAnnotations = make(map[string]string, len(in.Status.EffectiveConfig.Defaults.SyncedAnnotations))
		for k, v := range in.Status.EffectiveConfig.Defaults.SyncedAnnotations {
			out.Status.EffectiveConfig.Defaults.SyncedAnnotations[k] = v
		}
	}
}

func (in *BackupConfig) DeepCopy() *BackupConfig {
	if in == nil {
		return nil
	}
	out := new(BackupConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *BackupConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *BackupConfigList) DeepCopyInto(out *BackupConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]BackupConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *BackupConfigList) DeepCopy() *BackupConfigList {
	if in == nil {
		return nil
	}
	out := new(BackupConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *BackupConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s BackupConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
