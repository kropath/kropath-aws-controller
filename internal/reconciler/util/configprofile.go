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

package util

import (
	"context"
	"fmt"
	"strings"

	"github.com/kropath/kropath-aws-controller/internal/metrics"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// DefaultConfigProfile is the well-known fallthrough profile name for
	// <ResourceFamily>Config lookups (ADR-015 §5.9, ADR-002 D-8). It mirrors the
	// "general-policy" default every resource instance CRD declares for
	// spec.configRef (ADR-015 §3.4) — the two must stay in sync, which is why
	// every reconciler sources the fallthrough target from this single constant
	// instead of a per-family literal.
	DefaultConfigProfile = "general-policy"

	// ConfigProfileResolvedConditionType names the condition kropath-aws-controller
	// publishes on a <ResourceFamily>Config CR describing whether the global-tier
	// profile lookup resolved directly, via the ADR-015 §5.9 fallthrough hop, or
	// not at all. ADR-015 §5.9: "skipping is legal; skipping invisibly is not."
	ConfigProfileResolvedConditionType = "GlobalProfileResolved"
)

// Reason codes for global-tier config profile resolution (spec §2.1, §2.6).
// These strings are normative: they appear verbatim as
// ConfigProfileResolvedCondition's Reason and are the closed set
// ConfigProfileReasons() returns.
const (
	ReasonProfileFound       = "ProfileFound"
	ReasonProfileFallthrough = "ProfileFallthrough"
	ReasonProfileUnresolved  = "ProfileUnresolved"
)

// ConfigProfileReasons returns the closed set of reason values
// kropath_config_profile_resolutions_total can carry (spec §2.1), in the
// order LoadConfigWithFallthrough evaluates them.
func ConfigProfileReasons() []string {
	return []string{ReasonProfileFound, ReasonProfileFallthrough, ReasonProfileUnresolved}
}

// objectWithGVK is client.Object plus the SetGroupVersionKind setter that
// metav1.TypeMeta promotes onto every kropath config CR type. Generic method
// calls only see methods declared on the constraint interface itself, so
// client.Object alone is not enough even though every concrete type has it.
type objectWithGVK interface {
	client.Object
	SetGroupVersionKind(gvk schema.GroupVersionKind)
}

// LoadConfigWithFallthrough looks up <kind>/<requestedProfile> in namespace. If it
// does not exist, it looks up <kind>/<fallthroughProfile> exactly once (ADR-015
// §5.9: "exactly one fallthrough hop, the same at both tiers"). If neither exists,
// it returns a non-nil zero-value object with found=false, so callers can keep
// merging an empty tier exactly as they did before this rule existed — the only
// change is that a typo'd or unpopulated profile no longer silently discards the
// org-wide baseline when that baseline does exist under the fallthrough name.
//
// T is the config CR's struct type (e.g. v1alpha1.S3Config); PT is its pointer
// type, which must implement client.Object and SetGroupVersionKind.
func LoadConfigWithFallthrough[T any, PT interface {
	*T
	objectWithGVK
}](ctx context.Context, c client.Client, gvk schema.GroupVersionKind, namespace, requestedProfile, fallthroughProfile string) (obj PT, found bool, viaFallthrough bool, err error) {
	family := strings.ToLower(gvk.Kind)

	obj = newTypedConfig[T, PT](gvk)
	getErr := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: requestedProfile}, obj)
	if getErr == nil {
		metrics.ConfigProfileResolved(family, ReasonProfileFound)
		return obj, true, false, nil
	}
	if client.IgnoreNotFound(getErr) != nil {
		return nil, false, false, getErr
	}

	if requestedProfile != fallthroughProfile {
		fallback := newTypedConfig[T, PT](gvk)
		getErr = c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: fallthroughProfile}, fallback)
		if getErr == nil {
			metrics.ConfigProfileResolved(family, ReasonProfileFallthrough)
			return fallback, true, true, nil
		}
		if client.IgnoreNotFound(getErr) != nil {
			return nil, false, false, getErr
		}
	}

	metrics.ConfigProfileResolved(family, ReasonProfileUnresolved)
	return newTypedConfig[T, PT](gvk), false, false, nil
}

func newTypedConfig[T any, PT interface {
	*T
	objectWithGVK
}](gvk schema.GroupVersionKind) PT {
	obj := PT(new(T))
	obj.SetGroupVersionKind(gvk)
	return obj
}

// ConfigProfileResolvedCondition reports whether a global-tier <ResourceFamily>Config
// profile lookup (ADR-015 §5.9) resolved directly, via the fallthrough hop, or not at
// all. requestedProfile is the profile the local tier asked for (the name of the
// <ResourceFamily>Config CR being reconciled).
func ConfigProfileResolvedCondition(requestedProfile string, found, viaFallthrough bool, observedGeneration int64, now metav1.Time) metav1.Condition {
	switch {
	case found && !viaFallthrough:
		return metav1.Condition{
			Type:               ConfigProfileResolvedConditionType,
			Status:             metav1.ConditionTrue,
			Reason:             ReasonProfileFound,
			Message:            fmt.Sprintf("global config found for profile %q", requestedProfile),
			ObservedGeneration: observedGeneration,
			LastTransitionTime: now,
		}
	case found && viaFallthrough:
		return metav1.Condition{
			Type:               ConfigProfileResolvedConditionType,
			Status:             metav1.ConditionTrue,
			Reason:             ReasonProfileFallthrough,
			Message:            fmt.Sprintf("global config not found for profile %q; fell through to %q", requestedProfile, DefaultConfigProfile),
			ObservedGeneration: observedGeneration,
			LastTransitionTime: now,
		}
	default:
		return metav1.Condition{
			Type:               ConfigProfileResolvedConditionType,
			Status:             metav1.ConditionFalse,
			Reason:             ReasonProfileUnresolved,
			Message:            fmt.Sprintf("global config not found for profile %q or fallthrough profile %q", requestedProfile, DefaultConfigProfile),
			ObservedGeneration: observedGeneration,
			LastTransitionTime: now,
		}
	}
}
