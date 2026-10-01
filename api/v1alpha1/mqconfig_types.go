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

type MQConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MQConfigSpec   `json:"spec,omitempty"`
	Status MQConfigStatus `json:"status,omitempty"`
}

type MQConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MQConfig `json:"items"`
}

type MQConfigSpec struct {
	Mandatory cascade.MQConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.MQConfigSection `json:"defaults,omitempty"`
}

type MQConfigStatus struct {
	EffectiveConfig    EffectiveMQConfig  `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string             `json:"syncedTimestamp,omitempty"`
}

// EffectiveMQConfig is the fully merged governance document written by the controller
// into MQConfig.status.effectiveConfig.
type EffectiveMQConfig struct {
	AWS       ProviderIdentity            `json:"aws,omitempty"`
	Mandatory cascade.EffectiveMQSection  `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveMQSection  `json:"defaults,omitempty"`
}

func (in *MQConfig) DeepCopyInto(out *MQConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)

	// Spec.Mandatory — *bool fields
	out.Spec = in.Spec
	if in.Spec.Mandatory.PubliclyAccessible != nil {
		v := *in.Spec.Mandatory.PubliclyAccessible
		out.Spec.Mandatory.PubliclyAccessible = &v
	}
	if in.Spec.Mandatory.AutoMinorVersionUpgrade != nil {
		v := *in.Spec.Mandatory.AutoMinorVersionUpgrade
		out.Spec.Mandatory.AutoMinorVersionUpgrade = &v
	}
	if in.Spec.Mandatory.LogsGeneral != nil {
		v := *in.Spec.Mandatory.LogsGeneral
		out.Spec.Mandatory.LogsGeneral = &v
	}
	if in.Spec.Mandatory.LogsAudit != nil {
		v := *in.Spec.Mandatory.LogsAudit
		out.Spec.Mandatory.LogsAudit = &v
	}
	if in.Spec.Mandatory.EncryptionUseAWSOwnedKey != nil {
		v := *in.Spec.Mandatory.EncryptionUseAWSOwnedKey
		out.Spec.Mandatory.EncryptionUseAWSOwnedKey = &v
	}
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

	// Spec.Defaults — *bool fields
	if in.Spec.Defaults.PubliclyAccessible != nil {
		v := *in.Spec.Defaults.PubliclyAccessible
		out.Spec.Defaults.PubliclyAccessible = &v
	}
	if in.Spec.Defaults.AutoMinorVersionUpgrade != nil {
		v := *in.Spec.Defaults.AutoMinorVersionUpgrade
		out.Spec.Defaults.AutoMinorVersionUpgrade = &v
	}
	if in.Spec.Defaults.LogsGeneral != nil {
		v := *in.Spec.Defaults.LogsGeneral
		out.Spec.Defaults.LogsGeneral = &v
	}
	if in.Spec.Defaults.LogsAudit != nil {
		v := *in.Spec.Defaults.LogsAudit
		out.Spec.Defaults.LogsAudit = &v
	}
	if in.Spec.Defaults.EncryptionUseAWSOwnedKey != nil {
		v := *in.Spec.Defaults.EncryptionUseAWSOwnedKey
		out.Spec.Defaults.EncryptionUseAWSOwnedKey = &v
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

	// Status
	out.Status = in.Status
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}

	// Status.EffectiveConfig.Mandatory — *bool fields
	if in.Status.EffectiveConfig.Mandatory.PubliclyAccessible != nil {
		v := *in.Status.EffectiveConfig.Mandatory.PubliclyAccessible
		out.Status.EffectiveConfig.Mandatory.PubliclyAccessible = &v
	}
	if in.Status.EffectiveConfig.Mandatory.AutoMinorVersionUpgrade != nil {
		v := *in.Status.EffectiveConfig.Mandatory.AutoMinorVersionUpgrade
		out.Status.EffectiveConfig.Mandatory.AutoMinorVersionUpgrade = &v
	}
	if in.Status.EffectiveConfig.Mandatory.LogsGeneral != nil {
		v := *in.Status.EffectiveConfig.Mandatory.LogsGeneral
		out.Status.EffectiveConfig.Mandatory.LogsGeneral = &v
	}
	if in.Status.EffectiveConfig.Mandatory.LogsAudit != nil {
		v := *in.Status.EffectiveConfig.Mandatory.LogsAudit
		out.Status.EffectiveConfig.Mandatory.LogsAudit = &v
	}
	if in.Status.EffectiveConfig.Mandatory.EncryptionUseAWSOwnedKey != nil {
		v := *in.Status.EffectiveConfig.Mandatory.EncryptionUseAWSOwnedKey
		out.Status.EffectiveConfig.Mandatory.EncryptionUseAWSOwnedKey = &v
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

	// Status.EffectiveConfig.Defaults — *bool fields
	if in.Status.EffectiveConfig.Defaults.PubliclyAccessible != nil {
		v := *in.Status.EffectiveConfig.Defaults.PubliclyAccessible
		out.Status.EffectiveConfig.Defaults.PubliclyAccessible = &v
	}
	if in.Status.EffectiveConfig.Defaults.AutoMinorVersionUpgrade != nil {
		v := *in.Status.EffectiveConfig.Defaults.AutoMinorVersionUpgrade
		out.Status.EffectiveConfig.Defaults.AutoMinorVersionUpgrade = &v
	}
	if in.Status.EffectiveConfig.Defaults.LogsGeneral != nil {
		v := *in.Status.EffectiveConfig.Defaults.LogsGeneral
		out.Status.EffectiveConfig.Defaults.LogsGeneral = &v
	}
	if in.Status.EffectiveConfig.Defaults.LogsAudit != nil {
		v := *in.Status.EffectiveConfig.Defaults.LogsAudit
		out.Status.EffectiveConfig.Defaults.LogsAudit = &v
	}
	if in.Status.EffectiveConfig.Defaults.EncryptionUseAWSOwnedKey != nil {
		v := *in.Status.EffectiveConfig.Defaults.EncryptionUseAWSOwnedKey
		out.Status.EffectiveConfig.Defaults.EncryptionUseAWSOwnedKey = &v
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

func (in *MQConfig) DeepCopy() *MQConfig {
	if in == nil {
		return nil
	}
	out := new(MQConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *MQConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *MQConfigList) DeepCopyInto(out *MQConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]MQConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *MQConfigList) DeepCopy() *MQConfigList {
	if in == nil {
		return nil
	}
	out := new(MQConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *MQConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s MQConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
