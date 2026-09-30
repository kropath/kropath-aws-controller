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

type DynamoDBConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DynamoDBConfigSpec   `json:"spec,omitempty"`
	Status DynamoDBConfigStatus `json:"status,omitempty"`
}

type DynamoDBConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DynamoDBConfig `json:"items"`
}

type DynamoDBConfigSpec struct {
	Mandatory cascade.DynamoDBConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.DynamoDBConfigSection `json:"defaults,omitempty"`
}

type DynamoDBConfigStatus struct {
	EffectiveConfig    EffectiveDynamoDBConfig `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition      `json:"conditions,omitempty"`
	ObservedGeneration int64                   `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string                  `json:"syncedTimestamp,omitempty"`
}

type EffectiveDynamoDBConfig struct {
	AWS       ProviderIdentity                 `json:"aws,omitempty"`
	Mandatory cascade.EffectiveDynamoDBSection `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveDynamoDBSection `json:"defaults,omitempty"`
}

func deepCopyDynamoDBConfigSection(dst *cascade.DynamoDBConfigSection, src *cascade.DynamoDBConfigSection) {
	if src.EncryptionEnabled != nil {
		b := *src.EncryptionEnabled
		dst.EncryptionEnabled = &b
	}
	if src.DeletionProtectionEnabled != nil {
		b := *src.DeletionProtectionEnabled
		dst.DeletionProtectionEnabled = &b
	}
	if src.PointInTimeRecoveryEnabled != nil {
		b := *src.PointInTimeRecoveryEnabled
		dst.PointInTimeRecoveryEnabled = &b
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
	if src.Tags != nil {
		dst.Tags = make(map[string]string, len(src.Tags))
		for k, v := range src.Tags {
			dst.Tags[k] = v
		}
	}
}

func deepCopyEffectiveDynamoDBSection(dst *cascade.EffectiveDynamoDBSection, src *cascade.EffectiveDynamoDBSection) {
	if src.EncryptionEnabled != nil {
		b := *src.EncryptionEnabled
		dst.EncryptionEnabled = &b
	}
	if src.DeletionProtectionEnabled != nil {
		b := *src.DeletionProtectionEnabled
		dst.DeletionProtectionEnabled = &b
	}
	if src.PointInTimeRecoveryEnabled != nil {
		b := *src.PointInTimeRecoveryEnabled
		dst.PointInTimeRecoveryEnabled = &b
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
	if src.Tags != nil {
		dst.Tags = make(map[string]string, len(src.Tags))
		for k, v := range src.Tags {
			dst.Tags[k] = v
		}
	}
}

func (in *DynamoDBConfig) DeepCopyInto(out *DynamoDBConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	deepCopyDynamoDBConfigSection(&out.Spec.Mandatory, &in.Spec.Mandatory)
	deepCopyDynamoDBConfigSection(&out.Spec.Defaults, &in.Spec.Defaults)
	out.Status = in.Status
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
	deepCopyEffectiveDynamoDBSection(&out.Status.EffectiveConfig.Mandatory, &in.Status.EffectiveConfig.Mandatory)
	deepCopyEffectiveDynamoDBSection(&out.Status.EffectiveConfig.Defaults, &in.Status.EffectiveConfig.Defaults)
}

func (in *DynamoDBConfig) DeepCopy() *DynamoDBConfig {
	if in == nil {
		return nil
	}
	out := new(DynamoDBConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *DynamoDBConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *DynamoDBConfigList) DeepCopyInto(out *DynamoDBConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]DynamoDBConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *DynamoDBConfigList) DeepCopy() *DynamoDBConfigList {
	if in == nil {
		return nil
	}
	out := new(DynamoDBConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *DynamoDBConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s DynamoDBConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
