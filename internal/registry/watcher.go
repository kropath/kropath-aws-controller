// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/kropath/kropath-aws-controller/internal/metrics"
)

// Error reason labels for crdWatchErrorsTotal.
const (
	errReasonBuildFailed          = "build_failed"
	errReasonWatchFailed          = "watch_failed"
	errReasonPredicateUnparseable = "predicate_unparseable"
)

// crdGVR is the GVR for the apiextensions CRD type.
var crdGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

// Watcher is a manager.Runnable that lists and watches all CRDs and calls
// coord.OnGVKServable whenever a CRD transitions into the Established+served state.
// It implements LeaderElectionRunnable so that it only runs on the leader replica.
type Watcher struct {
	coord *Coordinator
	bctx  BuildCtx
}

// NewWatcher constructs a Watcher. The dynamic client is derived from the manager's
// REST config in Start.
func NewWatcher(coord *Coordinator, bctx BuildCtx) *Watcher {
	return &Watcher{coord: coord, bctx: bctx}
}

// NeedLeaderElection implements manager.LeaderElectionRunnable.
// The watcher mutates Coordinator state that must converge on one replica only.
func (w *Watcher) NeedLeaderElection() bool { return true }

// Start implements manager.Runnable. It creates a dynamic informer for CRDs,
// feeds add/update events through a rate-limited workqueue, and dispatches to
// coord.OnGVKServable for each event that matches the activation predicate.
func (w *Watcher) Start(ctx context.Context) error {
	dc, err := dynamic.NewForConfig(w.bctx.Manager.GetConfig())
	if err != nil {
		crdWatchErrorsTotal.WithLabelValues(errReasonWatchFailed).Inc()
		return fmt.Errorf("watcher: dynamic client: %w", err)
	}

	factory := dynamicinformer.NewDynamicSharedInformerFactory(dc, 0)
	informer := factory.ForResource(crdGVR).Informer()

	// Strip spec.versions[].schema before caching to reduce memory pressure.
	if err := informer.SetTransform(stripCRDSchema); err != nil {
		crdWatchErrorsTotal.WithLabelValues(errReasonWatchFailed).Inc()
		return fmt.Errorf("watcher: SetTransform: %w", err)
	}

	q := workqueue.NewTypedRateLimitingQueue(
		workqueue.DefaultTypedControllerRateLimiter[string]())
	defer q.ShutDown()

	enqueue := func(obj interface{}) {
		u, ok := obj.(*unstructured.Unstructured)
		if !ok {
			return
		}
		q.Add(u.GetName())
	}

	if _, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    enqueue,
		UpdateFunc: func(_, newObj interface{}) { enqueue(newObj) },
	}); err != nil {
		crdWatchErrorsTotal.WithLabelValues(errReasonWatchFailed).Inc()
		return fmt.Errorf("watcher: AddEventHandler: %w", err)
	}

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())

	w.bctx.Log.Info("CRD watcher started")

	defer runtime.HandleCrash()

	// activatedGVKs tracks which GVKs this process has already successfully
	// handed to OnGVKServable, so a repeat watch event on the same GVK (e.g.
	// an Update after the CRD is already Established) is distinguishable from
	// the first time it became servable (design §2.5's activated vs
	// already_active outcomes). Only ever touched from this single loop
	// goroutine, so it needs no locking.
	activatedGVKs := make(map[schema.GroupVersionKind]bool)

	for {
		item, shutdown := q.Get()
		if shutdown {
			return nil
		}
		func() {
			defer q.Done(item)
			// Fetch from the informer's local store (no API round-trip).
			rawObj, exists, err := informer.GetStore().GetByKey(item)
			if err != nil || !exists {
				metrics.CRDWatchEvent("store_miss")
				return
			}
			u, ok := rawObj.(*unstructured.Unstructured)
			if !ok {
				metrics.CRDWatchEvent("cast_failed")
				return
			}
			gvk, servable := crdServable(u)
			if !servable {
				metrics.CRDWatchEvent("not_servable")
				return
			}
			if callErr := w.coord.OnGVKServable(w.bctx, gvk); callErr != nil {
				w.bctx.Log.Error(callErr, "OnGVKServable failed", "gvk", gvk)
				crdWatchErrorsTotal.WithLabelValues(errReasonBuildFailed).Inc()
				return
			}
			if activatedGVKs[gvk] {
				metrics.CRDWatchEvent("already_active")
				return
			}
			activatedGVKs[gvk] = true
			metrics.CRDWatchEvent("activated")
		}()
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
}

// crdServable inspects a CRD unstructured object and returns the GVK that it
// serves plus true when BOTH of the following hold:
//  1. The CRD has a condition Established=True.
//  2. It has at least one version entry with name="v1alpha1" and served=true.
func crdServable(u *unstructured.Unstructured) (schema.GroupVersionKind, bool) {
	spec := u.Object["spec"]
	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return schema.GroupVersionKind{}, false
	}

	group, _, _ := unstructured.NestedString(u.Object, "spec", "group")
	kind, _, _ := unstructured.NestedString(u.Object, "spec", "names", "kind")
	if group == "" || kind == "" {
		return schema.GroupVersionKind{}, false
	}

	// Check Established=True condition.
	conditions, _, _ := unstructured.NestedSlice(u.Object, "status", "conditions")
	established := false
	for _, c := range conditions {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if cm["type"] == "Established" && cm["status"] == "True" {
			established = true
			break
		}
	}
	if !established {
		return schema.GroupVersionKind{}, false
	}

	// Check that v1alpha1 is served.
	versions, ok := specMap["versions"].([]interface{})
	if !ok {
		return schema.GroupVersionKind{}, false
	}
	for _, v := range versions {
		vm, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := vm["name"].(string)
		served, _ := vm["served"].(bool)
		if name == "v1alpha1" && served {
			return schema.GroupVersionKind{Group: group, Version: "v1alpha1", Kind: kind}, true
		}
	}
	return schema.GroupVersionKind{}, false
}

// stripCRDSchema removes spec.versions[].schema from each CRD before it enters
// the informer cache. OpenAPI schemas can be hundreds of kilobytes per CRD.
func stripCRDSchema(obj interface{}) (interface{}, error) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return obj, nil
	}
	versions, found, err := unstructured.NestedSlice(u.Object, "spec", "versions")
	if err != nil || !found {
		return obj, nil
	}
	stripped := make([]interface{}, len(versions))
	for i, v := range versions {
		vm, ok := v.(map[string]interface{})
		if !ok {
			stripped[i] = v
			continue
		}
		cp := make(map[string]interface{}, len(vm))
		for k, val := range vm {
			if k == "schema" {
				continue
			}
			cp[k] = val
		}
		stripped[i] = cp
	}
	if err := unstructured.SetNestedSlice(u.Object, stripped, "spec", "versions"); err != nil {
		return obj, fmt.Errorf("stripCRDSchema: %w", err)
	}
	return u, nil
}
