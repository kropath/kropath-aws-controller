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

type S3Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   S3ConfigSpec   `json:"spec,omitempty"`
	Status S3ConfigStatus `json:"status,omitempty"`
}

type S3ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []S3Config `json:"items"`
}

type S3ConfigSpec struct {
	Mandatory cascade.S3ConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.S3ConfigSection `json:"defaults,omitempty"`
}

// EffectiveS3Config is the v1alpha1 wrapper around the cascade merge result.
// It adds the AWS provider identity (ADR-010 D-3) so RGDs have a single read point.
type EffectiveS3Config struct {
	AWS       ProviderIdentity         `json:"aws,omitempty"`
	Mandatory cascade.EffectiveS3Section `json:"mandatory"`
	Defaults  cascade.EffectiveS3Section `json:"defaults"`
}

type S3ConfigStatus struct {
	EffectiveConfig    EffectiveS3Config  `json:"effectiveConfig,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string             `json:"syncedTimestamp,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

func (in *S3Config) DeepCopyInto(out *S3Config) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	// Spec deep copy — S3ConfigSection contains maps
	out.Spec = S3ConfigSpec{}
	deepCopyS3ConfigSection(&in.Spec.Mandatory, &out.Spec.Mandatory)
	deepCopyS3ConfigSection(&in.Spec.Defaults, &out.Spec.Defaults)
	// Status deep copy
	out.Status = S3ConfigStatus{
		ObservedGeneration: in.Status.ObservedGeneration,
		SyncedTimestamp:    in.Status.SyncedTimestamp,
	}
	deepCopyEffectiveS3Config(&in.Status.EffectiveConfig, &out.Status.EffectiveConfig)
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
}

func deepCopyS3ConfigSection(in, out *cascade.S3ConfigSection) {
	*out = *in
	if in.SyncedLabels != nil {
		out.SyncedLabels = make(map[string]string, len(in.SyncedLabels))
		for k, v := range in.SyncedLabels {
			out.SyncedLabels[k] = v
		}
	}
	if in.SyncedAnnotations != nil {
		out.SyncedAnnotations = make(map[string]string, len(in.SyncedAnnotations))
		for k, v := range in.SyncedAnnotations {
			out.SyncedAnnotations[k] = v
		}
	}
	if in.Tags != nil {
		out.Tags = make(map[string]string, len(in.Tags))
		for k, v := range in.Tags {
			out.Tags[k] = v
		}
	}
}

func deepCopyEffectiveS3Config(in, out *EffectiveS3Config) {
	out.AWS = in.AWS
	deepCopyEffectiveS3Section(&in.Mandatory, &out.Mandatory)
	deepCopyEffectiveS3Section(&in.Defaults, &out.Defaults)
}

func deepCopyEffectiveS3Section(in, out *cascade.EffectiveS3Section) {
	*out = *in
	if in.SyncedLabels != nil {
		out.SyncedLabels = make(map[string]string, len(in.SyncedLabels))
		for k, v := range in.SyncedLabels {
			out.SyncedLabels[k] = v
		}
	}
	if in.SyncedAnnotations != nil {
		out.SyncedAnnotations = make(map[string]string, len(in.SyncedAnnotations))
		for k, v := range in.SyncedAnnotations {
			out.SyncedAnnotations[k] = v
		}
	}
	if in.Tags != nil {
		out.Tags = make(map[string]string, len(in.Tags))
		for k, v := range in.Tags {
			out.Tags[k] = v
		}
	}
}

func (in *S3Config) DeepCopy() *S3Config {
	if in == nil {
		return nil
	}
	out := new(S3Config)
	in.DeepCopyInto(out)
	return out
}

func (in *S3Config) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *S3ConfigList) DeepCopyInto(out *S3ConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]S3Config, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *S3ConfigList) DeepCopy() *S3ConfigList {
	if in == nil {
		return nil
	}
	out := new(S3ConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *S3ConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s S3ConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
