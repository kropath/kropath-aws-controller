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

type CognitoConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CognitoConfigSpec   `json:"spec,omitempty"`
	Status CognitoConfigStatus `json:"status,omitempty"`
}

type CognitoConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []CognitoConfig `json:"items"`
}

type CognitoConfigSpec struct {
	Mandatory cascade.CognitoConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.CognitoConfigSection `json:"defaults,omitempty"`
}

type CognitoConfigStatus struct {
	EffectiveConfig    EffectiveCognitoConfig `json:"effectiveConfig,omitempty"`
	Conditions         []metav1.Condition     `json:"conditions,omitempty"`
	ObservedGeneration int64                  `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string                 `json:"syncedTimestamp,omitempty"`
}

type EffectiveCognitoConfig struct {
	AWS       ProviderIdentity                `json:"aws,omitempty"`
	Mandatory cascade.EffectiveCognitoSection `json:"mandatory,omitempty"`
	Defaults  cascade.EffectiveCognitoSection `json:"defaults,omitempty"`
}

func deepCopyCognitoConfigSection(dst *cascade.CognitoConfigSection, src *cascade.CognitoConfigSection) {
	if src.PasswordPolicy.RequireLowercase != nil {
		b := *src.PasswordPolicy.RequireLowercase
		dst.PasswordPolicy.RequireLowercase = &b
	}
	if src.PasswordPolicy.RequireNumbers != nil {
		b := *src.PasswordPolicy.RequireNumbers
		dst.PasswordPolicy.RequireNumbers = &b
	}
	if src.PasswordPolicy.RequireSymbols != nil {
		b := *src.PasswordPolicy.RequireSymbols
		dst.PasswordPolicy.RequireSymbols = &b
	}
	if src.PasswordPolicy.RequireUppercase != nil {
		b := *src.PasswordPolicy.RequireUppercase
		dst.PasswordPolicy.RequireUppercase = &b
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

func deepCopyEffectiveCognitoSection(dst *cascade.EffectiveCognitoSection, src *cascade.EffectiveCognitoSection) {
	if src.PasswordPolicy.RequireLowercase != nil {
		b := *src.PasswordPolicy.RequireLowercase
		dst.PasswordPolicy.RequireLowercase = &b
	}
	if src.PasswordPolicy.RequireNumbers != nil {
		b := *src.PasswordPolicy.RequireNumbers
		dst.PasswordPolicy.RequireNumbers = &b
	}
	if src.PasswordPolicy.RequireSymbols != nil {
		b := *src.PasswordPolicy.RequireSymbols
		dst.PasswordPolicy.RequireSymbols = &b
	}
	if src.PasswordPolicy.RequireUppercase != nil {
		b := *src.PasswordPolicy.RequireUppercase
		dst.PasswordPolicy.RequireUppercase = &b
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

func (in *CognitoConfig) DeepCopyInto(out *CognitoConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	// Spec — mandatory
	deepCopyCognitoConfigSection(&out.Spec.Mandatory, &in.Spec.Mandatory)
	// Spec — defaults
	deepCopyCognitoConfigSection(&out.Spec.Defaults, &in.Spec.Defaults)
	// Status conditions
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
	// Status effectiveConfig
	deepCopyEffectiveCognitoSection(&out.Status.EffectiveConfig.Mandatory, &in.Status.EffectiveConfig.Mandatory)
	deepCopyEffectiveCognitoSection(&out.Status.EffectiveConfig.Defaults, &in.Status.EffectiveConfig.Defaults)
}

func (in *CognitoConfig) DeepCopy() *CognitoConfig {
	if in == nil {
		return nil
	}
	out := new(CognitoConfig)
	in.DeepCopyInto(out)
	return out
}

func (in *CognitoConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *CognitoConfigList) DeepCopyInto(out *CognitoConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]CognitoConfig, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *CognitoConfigList) DeepCopy() *CognitoConfigList {
	if in == nil {
		return nil
	}
	out := new(CognitoConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *CognitoConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s CognitoConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
