// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

// Package registry binds each entry in features.All to the code that constructs
// and registers its controller. It is the sole place that imports both
// internal/features and the reconciler packages; all other packages must stay
// import-cycle-free.
package registry

import (
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/kropath/kropath-controller/internal/metrics"
)

// BuildCtx carries what a Build function needs to construct and register its controller.
type BuildCtx struct {
	Manager ctrl.Manager
	Log     logr.Logger
}

// Entry describes one reconciler and how to register it with a running manager.
//
// Required and Optional carry full GroupVersionKinds (not bare kind strings) so that
// labeloperator (which spans three provider API groups) and policydocument (whose ref
// kinds are all in aws.kropath.run) can be expressed without ambiguity.
type Entry struct {
	// Package is the Go sub-package under internal/reconciler/. It must match the
	// Package field in the corresponding features.Reconciler — this is the stable
	// machine identifier that joins the two tables.
	Package string

	// Required lists the GVKs that must be servable before this entry is registered.
	// A startup gate registers the reconciler as active only when all Required GVKs
	// are present; otherwise the entry stays pending.
	Required []schema.GroupVersionKind

	// Optional lists the GVKs that are watched when present and attached later via
	// AddKindWatch if they appear after the controller is already active.
	Optional []schema.GroupVersionKind

	// Build constructs and registers the reconciler's controller against the manager.
	// It must use ctrl.NewControllerManagedBy(bctx.Manager)...Build(r) (not .Complete)
	// so that the returned controller.Controller handle can be retained for later
	// AddKindWatch calls. Build may return (nil, nil) for reconcilers that create
	// multiple controllers internally (e.g. labeloperator).
	// servedOptional lists the Optional GVKs that are currently served; the reconciler
	// may pre-register watches for them at build time.
	Build func(bctx BuildCtx, servedOptional []schema.GroupVersionKind) (controller.Controller, error)

	// AddKindWatch attaches an optional GVK to an already-running controller. Called
	// by the CRD watcher (KRO-849) when a previously-absent optional kind becomes
	// servable. May be nil if Optional is empty.
	AddKindWatch func(c controller.Controller, gvk schema.GroupVersionKind) error
}

// entryState holds the runtime state for a single entry.
type entryState struct {
	entry            Entry
	handle           controller.Controller // non-nil once active (may be nil for multi-controller entries)
	active           bool
	missingKindNames []string                         // kind names of Required GVKs not yet served
	attachedOptional map[schema.GroupVersionKind]bool // optional kinds already watched

	// pendingTimestampSet is true once this process has recorded a
	// kropath_registry_reconciler_pending_since_timestamp_seconds child for
	// this entry (M-17). Guards against RunGate overwriting an
	// already-recorded pending timestamp on a later call, and is reset to
	// false when the entry activates so a later re-parking (were that ever
	// possible) would record a fresh timestamp.
	pendingTimestampSet bool
}

// Coordinator owns all registry state: per-entry active/pending status, per-optional-kind
// attached/not, and the retained controller handles. It is the sole owner of this state —
// the startup gate and the CRD watcher (KRO-849) are thin drivers over it.
//
// Access is mutex-guarded: the startup gate runs on the main goroutine and the KRO-849
// watcher will run as a manager runnable; both mutate the same table.
type Coordinator struct {
	mu         sync.Mutex
	entries    []entryState
	servedGVKs map[schema.GroupVersionKind]bool // grows as CRDs are discovered at runtime
}

// Add appends an entry to the coordinator. Must be called before any Run method.
func (c *Coordinator) Add(e Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, entryState{entry: e})
}

// RunUnconditional registers every entry unconditionally without any CRD availability
// gate. This is the commit-1 behaviour, identical to the hand-written SetupWithManager
// blocks in the previous main.go. It is used until the startup gate (RunGate) is
// introduced in the next commit.
func (c *Coordinator) RunUnconditional(bctx BuildCtx) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.entries {
		es := &c.entries[i]
		if es.active {
			continue
		}
		handle, err := es.entry.Build(bctx, nil)
		if err != nil {
			return fmt.Errorf("registry: building %s: %w", es.entry.Package, err)
		}
		es.handle = handle
		es.active = true
	}
	return nil
}

// OnGVKServable is called by the CRD watcher when a CRD becomes established. It:
//   - Marks the GVK as served in the coordinator's set.
//   - For pending entries whose Required GVKs are now all served, builds the reconciler.
//   - For active entries that declare the GVK as Optional and haven't attached it yet,
//     calls AddKindWatch to attach the watch to the running controller.
func (c *Coordinator) OnGVKServable(bctx BuildCtx, gvk schema.GroupVersionKind) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.servedGVKs == nil {
		c.servedGVKs = make(map[schema.GroupVersionKind]bool)
	}
	c.servedGVKs[gvk] = true

	for i := range c.entries {
		es := &c.entries[i]

		if es.active {
			// Already active — check if this GVK is an unattached Optional.
			if es.entry.AddKindWatch == nil {
				continue
			}
			// Optional == nil means wildcard: any GVK may trigger AddKindWatch.
			// This is used by labeloperator which registers a controller per provider GVK.
			isOptional := es.entry.Optional == nil
			for _, opt := range es.entry.Optional {
				if opt == gvk {
					isOptional = true
					break
				}
			}
			if !isOptional {
				continue
			}
			if es.attachedOptional != nil && es.attachedOptional[gvk] {
				continue
			}
			if err := es.entry.AddKindWatch(es.handle, gvk); err != nil {
				crdWatchErrorsTotal.WithLabelValues(errReasonWatchFailed).Inc()
				return fmt.Errorf("registry: AddKindWatch for %v on %s: %w", gvk, es.entry.Package, err)
			}
			if es.attachedOptional == nil {
				es.attachedOptional = make(map[schema.GroupVersionKind]bool)
			}
			es.attachedOptional[gvk] = true
			metrics.OptionalKindsAttached(es.entry.Package, len(es.attachedOptional))
			continue
		}

		// Not yet active — check if all Required GVKs are now served.
		allServed := true
		for _, req := range es.entry.Required {
			if !c.servedGVKs[req] {
				allServed = false
				break
			}
		}
		if !allServed {
			continue
		}

		// All Required served — compute which Optional are also served.
		var servedOptional []schema.GroupVersionKind
		for _, opt := range es.entry.Optional {
			if c.servedGVKs[opt] {
				servedOptional = append(servedOptional, opt)
			}
		}

		handle, err := es.entry.Build(bctx, servedOptional)
		if err != nil {
			crdWatchErrorsTotal.WithLabelValues(errReasonBuildFailed).Inc()
			return fmt.Errorf("registry: building %s on CRD event: %w", es.entry.Package, err)
		}
		es.handle = handle
		es.active = true
		es.missingKindNames = nil

		if len(servedOptional) > 0 {
			es.attachedOptional = make(map[schema.GroupVersionKind]bool, len(servedOptional))
			for _, opt := range servedOptional {
				es.attachedOptional[opt] = true
			}
			metrics.OptionalKindsAttached(es.entry.Package, len(es.attachedOptional))
		}

		reconcilerActivationsTotal.WithLabelValues(es.entry.Package).Inc()
		reconcilerActive.WithLabelValues(es.entry.Package).Set(1)
		reconcilerMissingKinds.WithLabelValues(es.entry.Package).Set(0)
		metrics.ReconcilerPendingClear(es.entry.Package) // M-17: runtime-activation path
		es.pendingTimestampSet = false

		bctx.Log.Info("reconciler activated by CRD watcher",
			"package", es.entry.Package,
			"trigger", gvk.Kind)
	}

	return nil
}

// ReconcilerState returns the runtime state for the given package name.
// Implements features.StateReader.
func (c *Coordinator) ReconcilerState(pkg string) (state string, missingKinds []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, es := range c.entries {
		if es.entry.Package != pkg {
			continue
		}
		if es.active {
			return "active", nil
		}
		return "pending", append([]string(nil), es.missingKindNames...)
	}
	return "", nil
}

// Entries returns a snapshot of all entries (for testing and the /features handler).
func (c *Coordinator) Entries() []Entry {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Entry, len(c.entries))
	for i, es := range c.entries {
		out[i] = es.entry
	}
	return out
}

// ActivePackages returns the set of package names currently marked active (for testing).
func (c *Coordinator) ActivePackages() map[string]bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	m := make(map[string]bool, len(c.entries))
	for _, es := range c.entries {
		m[es.entry.Package] = es.active
	}
	return m
}
