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

type EC2Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EC2ConfigSpec   `json:"spec,omitempty"`
	Status EC2ConfigStatus `json:"status,omitempty"`
}

type EC2ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []EC2Config `json:"items"`
}

type EC2ConfigSpec struct {
	Mandatory cascade.EC2ConfigSection `json:"mandatory,omitempty"`
	Defaults  cascade.EC2ConfigSection `json:"defaults,omitempty"`
}

type EC2ConfigStatus struct {
	EffectiveConfig    EffectiveEC2Config `json:"effectiveConfig,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	SyncedTimestamp    string             `json:"syncedTimestamp,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

type EffectiveEC2Config struct {
	AWS       ProviderIdentity            `json:"aws,omitempty"`
	Mandatory cascade.EffectiveEC2Section `json:"mandatory"`
	Defaults  cascade.EffectiveEC2Section `json:"defaults"`
}

func copyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func deepCopyEC2Section(in cascade.EC2ConfigSection) cascade.EC2ConfigSection {
	out := in
	out.Tags = copyStringMap(in.Tags)
	out.SyncedLabels = copyStringMap(in.SyncedLabels)
	out.SyncedAnnotations = copyStringMap(in.SyncedAnnotations)
	return out
}

func deepCopyEffectiveEC2Section(in cascade.EffectiveEC2Section) cascade.EffectiveEC2Section {
	out := in
	out.Tags = copyStringMap(in.Tags)
	out.SyncedLabels = copyStringMap(in.SyncedLabels)
	out.SyncedAnnotations = copyStringMap(in.SyncedAnnotations)
	return out
}

func (in *EC2Config) DeepCopyInto(out *EC2Config) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec.Mandatory = deepCopyEC2Section(in.Spec.Mandatory)
	out.Spec.Defaults = deepCopyEC2Section(in.Spec.Defaults)
	out.Status.EffectiveConfig.Mandatory = deepCopyEffectiveEC2Section(in.Status.EffectiveConfig.Mandatory)
	out.Status.EffectiveConfig.Defaults = deepCopyEffectiveEC2Section(in.Status.EffectiveConfig.Defaults)
	if in.Status.Conditions != nil {
		out.Status.Conditions = make([]metav1.Condition, len(in.Status.Conditions))
		copy(out.Status.Conditions, in.Status.Conditions)
	}
}

func (in *EC2Config) DeepCopy() *EC2Config {
	if in == nil {
		return nil
	}
	out := new(EC2Config)
	in.DeepCopyInto(out)
	return out
}

func (in *EC2Config) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *EC2ConfigList) DeepCopyInto(out *EC2ConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]EC2Config, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *EC2ConfigList) DeepCopy() *EC2ConfigList {
	if in == nil {
		return nil
	}
	out := new(EC2ConfigList)
	in.DeepCopyInto(out)
	return out
}

func (in *EC2ConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// MarshalJSON omits "effectiveConfig" from the wire payload when it is the
// Go zero value -- see marshalConfigStatus in effectiveconfig_marshal.go
// (KRO-1199).
func (s EC2ConfigStatus) MarshalJSON() ([]byte, error) {
	return marshalConfigStatus(s.EffectiveConfig, s.ObservedGeneration, s.SyncedTimestamp, s.Conditions)
}
