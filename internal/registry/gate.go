// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kropath/kropath-aws-controller/internal/metrics"
)

// KropathConfigGVK is the GVK that must be served before any reconciler starts.
// It is an install prerequisite — the operator exits non-zero if absent.
var KropathConfigGVK = schema.GroupVersionKind{
	Group:   "aws.kropath.run",
	Version: "v1alpha1",
	Kind:    "KropathConfig",
}

// ResourceLister is the subset of discovery.DiscoveryInterface used by GatherServedGVKs.
// Using a narrow interface makes the function easy to stub in tests.
type ResourceLister interface {
	ServerPreferredResources() ([]*metav1.APIResourceList, error)
}

// GatherServedGVKs queries the API server and returns the set of GVKs that are
// currently served (i.e. have an established CRD with this version available).
// Partial failures from ServerPreferredResources are ignored as long as at
// least one resource list is returned.
func GatherServedGVKs(dc ResourceLister) (map[schema.GroupVersionKind]bool, error) {
	resourceLists, err := dc.ServerPreferredResources()
	if err != nil && len(resourceLists) == 0 {
		return nil, fmt.Errorf("discovery: listing server resources: %w", err)
	}
	served := make(map[schema.GroupVersionKind]bool, len(resourceLists)*10)
	for _, rl := range resourceLists {
		gv, parseErr := schema.ParseGroupVersion(rl.GroupVersion)
		if parseErr != nil {
			continue
		}
		for _, r := range rl.APIResources {
			served[schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: r.Kind}] = true
		}
	}
	return served, nil
}

// RunGate registers each entry whose Required GVKs are all present in servedGVKs,
// leaving the rest pending. Before the per-entry loop it checks the KropathConfig
// precondition: if KropathConfig is not served, RunGate returns an error — the
// caller should treat this as fatal and exit non-zero.
func (c *Coordinator) RunGate(bctx BuildCtx, servedGVKs map[schema.GroupVersionKind]bool) error {
	if !servedGVKs[KropathConfigGVK] {
		return fmt.Errorf("startup gate: %s/%s %s is not served — "+
			"install the KropathConfig CRD before starting kropath-aws-controller",
			KropathConfigGVK.Group, KropathConfigGVK.Version, KropathConfigGVK.Kind)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Seed the coordinator's runtime GVK set from the startup discovery snapshot.
	if c.servedGVKs == nil {
		c.servedGVKs = make(map[schema.GroupVersionKind]bool, len(servedGVKs))
	}
	for gvk, ok := range servedGVKs {
		if ok {
			c.servedGVKs[gvk] = true
		}
	}

	for i := range c.entries {
		es := &c.entries[i]
		if es.active {
			continue
		}

		var missingKindNames []string
		for _, gvk := range es.entry.Required {
			if !servedGVKs[gvk] {
				missingKindNames = append(missingKindNames, gvk.Kind)
			}
		}
		es.missingKindNames = missingKindNames

		reconcilerMissingKinds.WithLabelValues(es.entry.Package).Set(float64(len(missingKindNames)))

		if len(missingKindNames) > 0 {
			reconcilerActive.WithLabelValues(es.entry.Package).Set(0)
			if !es.pendingTimestampSet {
				metrics.ReconcilerPendingSince(es.entry.Package, time.Now()) // M-17: startup-gate path
				es.pendingTimestampSet = true
			}
			bctx.Log.Info("reconciler pending: required CRDs not yet served",
				"package", es.entry.Package,
				"missingCount", len(missingKindNames),
				"missing", missingKindNames)
			continue
		}

		// All Required served — compute which Optional are also served.
		var servedOptional []schema.GroupVersionKind
		for _, opt := range es.entry.Optional {
			if servedGVKs[opt] {
				servedOptional = append(servedOptional, opt)
			}
		}

		handle, err := es.entry.Build(bctx, servedOptional)
		if err != nil {
			return fmt.Errorf("registry: building %s: %w", es.entry.Package, err)
		}
		es.handle = handle
		es.active = true

		if len(servedOptional) > 0 {
			es.attachedOptional = make(map[schema.GroupVersionKind]bool, len(servedOptional))
			for _, opt := range servedOptional {
				es.attachedOptional[opt] = true
			}
			metrics.OptionalKindsAttached(es.entry.Package, len(es.attachedOptional))
		}

		// Pre-seed wildcard-Optional entries: discovery-lag can cause a GVK miss in
		// labeloperator.Setup; registering missed controllers here (before mgr.Start)
		// keeps them in the initial cache sync rather than adding informers post-startup,
		// which would flip /readyz to 500 until the new informer catches up (KRO-857).
		if es.entry.Optional == nil && es.entry.AddKindWatch != nil {
			es.attachedOptional = make(map[schema.GroupVersionKind]bool, len(c.servedGVKs))
			for gvk := range c.servedGVKs {
				if err := es.entry.AddKindWatch(es.handle, gvk); err != nil {
					return fmt.Errorf("registry: pre-seeding AddKindWatch for %v on %s: %w", gvk, es.entry.Package, err)
				}
				es.attachedOptional[gvk] = true
			}
			metrics.OptionalKindsAttached(es.entry.Package, len(es.attachedOptional))
		}

		reconcilerActive.WithLabelValues(es.entry.Package).Set(1)
		metrics.ReconcilerPendingClear(es.entry.Package) // M-17: startup-gate path
		es.pendingTimestampSet = false
	}
	return nil
}
