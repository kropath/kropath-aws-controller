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

// Package labeloperator watches all resources under the three provider API
// groups and ensures each carries the <provider>.kropath.run/resource-name
// label whose value equals metadata.name.
package labeloperator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/kropath/kropath-aws-controller/internal/metrics"
)

// providerGroups maps each watched API group to its resource-name label key.
var providerGroups = map[string]string{
	"aws.kropath.run":   "aws.kropath.run/resource-name",
	"gcp.kropath.run":   "gcp.kropath.run/resource-name",
	"azure.kropath.run": "azure.kropath.run/resource-name",
}

// Reconciler applies the <provider>.kropath.run/resource-name label to a
// specific resource kind. One instance is created per GVK discovered at
// startup.
type Reconciler struct {
	Client   client.Client
	Log      logr.Logger
	GVK      schema.GroupVersionKind
	LabelKey string
}

// Reconcile ensures the resource carries the correct resource-name label.
// It is a no-op when the label already matches metadata.name.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(r.GVK)
	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !needsPatch(obj, r.LabelKey) {
		r.Log.V(1).Info("label already correct, no-op", "kind", r.GVK.Kind, "name", req.Name)
		return ctrl.Result{}, nil
	}

	// JSON merge patch touching metadata.labels only — never spec or status.
	patchData := buildLabelPatch(r.LabelKey, obj.GetName())
	target := &unstructured.Unstructured{}
	target.SetGroupVersionKind(r.GVK)
	target.SetName(obj.GetName())
	target.SetNamespace(obj.GetNamespace())

	if err := r.Client.Patch(ctx, target, client.RawPatch(types.MergePatchType, patchData)); err != nil {
		if apierrors.IsNotFound(err) {
			metrics.LabelOperatorPatch(r.GVK.Group, "not_found")
			return ctrl.Result{}, nil
		}
		metrics.LabelOperatorPatch(r.GVK.Group, "error")
		return ctrl.Result{}, fmt.Errorf("patching label on %s %s/%s: %w",
			r.GVK.Kind, req.Namespace, req.Name, err)
	}

	metrics.LabelOperatorPatch(r.GVK.Group, "patched")
	r.Log.Info("patched resource-name label",
		"kind", r.GVK.Kind, "name", req.Name, "namespace", req.Namespace)
	return ctrl.Result{}, nil
}

// needsPatch reports whether the object requires a label patch.
func needsPatch(obj *unstructured.Unstructured, labelKey string) bool {
	labels := obj.GetLabels()
	if labels == nil {
		return true
	}
	v, ok := labels[labelKey]
	return !ok || v != obj.GetName()
}

// buildLabelPatch returns a JSON merge patch that sets labelKey to value.
func buildLabelPatch(labelKey, value string) []byte {
	kJSON, _ := json.Marshal(labelKey)
	vJSON, _ := json.Marshal(value)
	return []byte(fmt.Sprintf(`{"metadata":{"labels":{%s:%s}}}`, kJSON, vJSON))
}

// unchangedLabelPredicate skips Update events where the tracked label key has
// not changed, preventing unnecessary reconciliation from spec/status writes.
func unchangedLabelPredicate(labelKey string) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc:  func(event.CreateEvent) bool { return true },
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return true },
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldVal := e.ObjectOld.GetLabels()[labelKey]
			newVal := e.ObjectNew.GetLabels()[labelKey]
			return oldVal != newVal
		},
	}
}

// LabelKeyForGroup returns the resource-name label key for the given API group.
// Returns "" if the group is not a known provider group.
func LabelKeyForGroup(group string) string {
	return providerGroups[group]
}

// ControllerName returns the unique controller name for the given provider GVK.
func ControllerName(gvk schema.GroupVersionKind) string {
	return fmt.Sprintf("label-operator-%s-%s",
		strings.ReplaceAll(gvk.Group, ".", "-"),
		strings.ToLower(gvk.Kind))
}

// NewReconciler constructs a Reconciler for the given GVK, deriving LabelKey from the API group.
func NewReconciler(gvk schema.GroupVersionKind, c client.Client, log logr.Logger) *Reconciler {
	return &Reconciler{
		Client:   c,
		Log:      log.WithValues("kind", gvk.Kind),
		GVK:      gvk,
		LabelKey: providerGroups[gvk.Group],
	}
}

// RegisterController creates and registers a label-operator controller for the given GVK,
// returning its handle. Uses Build (not Complete) so the handle can be retained for future use.
func RegisterController(mgr ctrl.Manager, ctrlName string, gvk schema.GroupVersionKind, log logr.Logger) (controller.Controller, error) {
	r := NewReconciler(gvk, mgr.GetClient(), log)
	proto := &unstructured.Unstructured{}
	proto.SetGroupVersionKind(gvk)
	c, err := ctrl.NewControllerManagedBy(mgr).
		Named(ctrlName).
		For(proto).
		WithEventFilter(unchangedLabelPredicate(r.LabelKey)).
		Build(r)
	if err != nil {
		return nil, fmt.Errorf("label-operator: registering controller for %s/%s: %w", gvk.Group, gvk.Kind, err)
	}
	return c, nil
}

// Setup discovers all resource types under the three provider API groups and registers
// one controller per GVK, returning a map of handles keyed by gvk.String().
// Handles are retained for later AddKindWatch calls (runtime controller registration).
func Setup(mgr ctrl.Manager, log logr.Logger) (map[string]controller.Controller, error) {
	disc, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("label-operator: creating discovery client: %w", err)
	}
	handles := make(map[string]controller.Controller)
	for group := range providerGroups {
		if err := setupGroup(mgr, log, disc, group, handles); err != nil {
			return nil, err
		}
	}
	return handles, nil
}

func setupGroup(mgr ctrl.Manager, log logr.Logger, disc discovery.DiscoveryInterface, group string, handles map[string]controller.Controller) error {
	gv := group + "/v1alpha1"
	resources, err := disc.ServerResourcesForGroupVersion(gv)
	if err != nil {
		log.Info("label-operator: no resources found for group, skipping", "gv", gv)
		metrics.LabelOperatorGroupDiscovery(group, "error")
		metrics.LabelOperatorWatchedKinds(group, 0)
		return nil
	}
	registered := 0
	for _, res := range resources.APIResources {
		if strings.Contains(res.Name, "/") {
			continue
		}
		gvk := schema.GroupVersionKind{Group: group, Version: "v1alpha1", Kind: res.Kind}
		c, err := RegisterController(mgr, ControllerName(gvk), gvk, log)
		if err != nil {
			return err
		}
		handles[gvk.String()] = c
		registered++
	}
	if registered == 0 {
		metrics.LabelOperatorGroupDiscovery(group, "empty")
	} else {
		metrics.LabelOperatorGroupDiscovery(group, "discovered")
	}
	metrics.LabelOperatorWatchedKinds(group, registered)
	return nil
}
