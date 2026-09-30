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

package quicksightconfig

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/kropath/kropath-aws-controller/internal/cascade"
	"github.com/kropath/kropath-aws-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const resourceNameLabel = "aws.kropath.run/resource-name"

type Reconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cfg := &v1alpha1.QuickSightConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("QuickSightConfig"))
	if err := r.Client.Get(ctx, req.NamespacedName, cfg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.ensureLabel(ctx, cfg); err != nil {
		return ctrl.Result{}, err
	}

	updated, result, reconcileErr := r.reconcile(ctx, cfg)
	if updated {
		if err := r.Client.Status().Update(ctx, cfg); err != nil {
			return ctrl.Result{}, err
		}
	}
	if reconcileErr != nil {
		return ctrl.Result{}, reconcileErr
	}
	return result, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, err := r.BuildWithManager(mgr)
	return err
}

func (r *Reconciler) BuildWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.QuickSightConfig{}).
		Watches(
			&v1alpha1.QuickSightConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForQuickSightConfigChange),
		).
		Watches(
			&v1alpha1.KropathConfig{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForKropathConfigChange),
		).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForNamespaceChange),
		).
		Build(r)
}

func (r *Reconciler) reconcile(ctx context.Context, cfg *v1alpha1.QuickSightConfig) (bool, ctrl.Result, error) {
	now := metav1.Now()
	placement, err := util.ResolveFamilyPlacement(ctx, r.Client, cfg.Namespace, cfg.Generation, now)
	if err != nil {
		reconciledCond, placementCond := util.NamespaceUnreadableCondition(cfg.Namespace, err, cfg.Generation, now)
		updated := r.withholdEffectiveConfig(cfg, reconciledCond, &placementCond)
		return updated, ctrl.Result{}, err
	}
	if placement.ReconciledOverride != nil {
		updated := r.withholdEffectiveConfig(cfg, *placement.ReconciledOverride, placement.PlacementCondition)
		return updated, ctrl.Result{}, nil
	}
	globalNS := placement.GlobalNamespace
	globalKropath, err := r.loadKropathConfig(ctx, globalNS, util.KropathConfigName)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	localKropath, err := r.loadKropathConfig(ctx, cfg.Namespace, util.KropathConfigName)
	if err != nil {
		return false, ctrl.Result{}, err
	}
	globalQS, globalQSFound, globalQSViaFallthrough, err := util.LoadConfigWithFallthrough[v1alpha1.QuickSightConfig](
		ctx, r.Client, v1alpha1.GroupVersion.WithKind("QuickSightConfig"), globalNS, cfg.Name, util.DefaultConfigProfile)
	if err != nil {
		return false, ctrl.Result{}, err
	}

	cfg.Status.ObservedGeneration = cfg.Generation

	// Build QuickSightKropathSections from KropathConfig.mandatory.quicksight / .defaults.quicksight
	// PLUS the tier-level tags/syncedLabels/syncedAnnotations.
	globalKropathMandatoryQS := cascade.QuickSightKropathSection{
		ImportMode:        globalKropath.Spec.Mandatory.QuickSight.ImportMode,
		Tags:              globalKropath.Spec.Mandatory.Tags,
		SyncedLabels:      globalKropath.Spec.Mandatory.SyncedLabels,
		SyncedAnnotations: globalKropath.Spec.Mandatory.SyncedAnnotations,
	}
	localKropathMandatoryQS := cascade.QuickSightKropathSection{
		ImportMode:        localKropath.Spec.Mandatory.QuickSight.ImportMode,
		Tags:              localKropath.Spec.Mandatory.Tags,
		SyncedLabels:      localKropath.Spec.Mandatory.SyncedLabels,
		SyncedAnnotations: localKropath.Spec.Mandatory.SyncedAnnotations,
	}
	localKropathDefaultsQS := cascade.QuickSightKropathSection{
		ImportMode:        localKropath.Spec.Defaults.QuickSight.ImportMode,
		Tags:              localKropath.Spec.Defaults.Tags,
		SyncedLabels:      localKropath.Spec.Defaults.SyncedLabels,
		SyncedAnnotations: localKropath.Spec.Defaults.SyncedAnnotations,
	}
	globalKropathDefaultsQS := cascade.QuickSightKropathSection{
		ImportMode:        globalKropath.Spec.Defaults.QuickSight.ImportMode,
		Tags:              globalKropath.Spec.Defaults.Tags,
		SyncedLabels:      globalKropath.Spec.Defaults.SyncedLabels,
		SyncedAnnotations: globalKropath.Spec.Defaults.SyncedAnnotations,
	}

	eff := cascade.MergeQuickSightCascade(
		globalKropathMandatoryQS,
		localKropathMandatoryQS,
		globalQS.Spec.Mandatory,
		cfg.Spec.Mandatory,
		cfg.Spec.Defaults,
		globalQS.Spec.Defaults,
		localKropathDefaultsQS,
		globalKropathDefaultsQS,
	)

	newCond := metav1.Condition{
		Type:               "Reconciled",
		Status:             metav1.ConditionTrue,
		Reason:             "CascadeMerged",
		Message:            "",
		ObservedGeneration: cfg.Generation,
		LastTransitionTime: now,
	}
	profileCond := util.ConfigProfileResolvedCondition(cfg.Name, globalQSFound, globalQSViaFallthrough, cfg.Generation, now)
	newEffConfig := v1alpha1.EffectiveQuickSightConfig{
		AWS:       placement.Identity,
		Mandatory: eff.Mandatory,
		Defaults:  eff.Defaults,
	}

	if !conditionNeedsUpdate(cfg.Status.Conditions, newCond) &&
		!conditionNeedsUpdate(cfg.Status.Conditions, *placement.PlacementCondition) &&
		!conditionNeedsUpdate(cfg.Status.Conditions, profileCond) &&
		reflect.DeepEqual(cfg.Status.EffectiveConfig, newEffConfig) {
		return false, ctrl.Result{}, nil
	}

	cfg.Status.Conditions = setCondition(cfg.Status.Conditions, newCond)
	cfg.Status.Conditions = setCondition(cfg.Status.Conditions, *placement.PlacementCondition)
	cfg.Status.Conditions = setCondition(cfg.Status.Conditions, profileCond)
	cfg.Status.EffectiveConfig = newEffConfig
	cfg.Status.SyncedTimestamp = now.UTC().Format(time.RFC3339)

	return true, ctrl.Result{}, nil
}

// withholdEffectiveConfig publishes reconciledCond (and placementCond, when
// non-nil) and clears status.effectiveConfig entirely -- a placement failure or
// a governance-only namespace must never leave a stale or partial identity
// published (spec §4.2, §6.4, AC-14). It reports whether anything changed so
// the caller can skip an unnecessary status write.
func (r *Reconciler) withholdEffectiveConfig(cfg *v1alpha1.QuickSightConfig, reconciledCond metav1.Condition, placementCond *metav1.Condition) bool {
	changed := conditionNeedsUpdate(cfg.Status.Conditions, reconciledCond)
	cfg.Status.Conditions = setCondition(cfg.Status.Conditions, reconciledCond)
	if placementCond != nil {
		if conditionNeedsUpdate(cfg.Status.Conditions, *placementCond) {
			changed = true
		}
		cfg.Status.Conditions = setCondition(cfg.Status.Conditions, *placementCond)
	}
	var zero v1alpha1.EffectiveQuickSightConfig
	if !reflect.DeepEqual(cfg.Status.EffectiveConfig, zero) {
		cfg.Status.EffectiveConfig = zero
		changed = true
	}
	if changed {
		cfg.Status.ObservedGeneration = cfg.Generation
		cfg.Status.SyncedTimestamp = reconciledCond.LastTransitionTime.UTC().Format(time.RFC3339)
	}
	return changed
}

func (r *Reconciler) loadKropathConfig(ctx context.Context, namespace, name string) (*v1alpha1.KropathConfig, error) {
	cfg := &v1alpha1.KropathConfig{}
	cfg.SetGroupVersionKind(v1alpha1.GroupVersion.WithKind("KropathConfig"))
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cfg); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return &v1alpha1.KropathConfig{}, nil
		}
		return nil, err
	}
	return cfg, nil
}

func (r *Reconciler) requestsForKropathConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	kpc, ok := obj.(*v1alpha1.KropathConfig)
	if !ok {
		return nil
	}

	var list v1alpha1.QuickSightConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		util.RecordMapFuncListError(r.Log, "quicksightconfig", "kropathconfig", err)
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		// Local tier: KropathConfig lives in the item's own namespace. Classified
		// by namespace, not name -- the name is a fixed singleton (ADR-018 D-1).
		if kpc.Namespace == item.Namespace {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
			})
			continue
		}
		// Global tier: KropathConfig lives in the item's resolved global namespace.
		_, globalNS, err := util.ResolveNamespaceRole(ctx, r.Client, item.Namespace)
		if err != nil {
			continue
		}
		if kpc.Namespace == globalNS {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
			})
		}
	}
	return requests
}

func (r *Reconciler) requestsForQuickSightConfigChange(ctx context.Context, obj client.Object) []ctrl.Request {
	trigger, ok := obj.(*v1alpha1.QuickSightConfig)
	if !ok {
		return nil
	}

	var list v1alpha1.QuickSightConfigList
	if err := r.Client.List(ctx, &list); err != nil {
		util.RecordMapFuncListError(r.Log, "quicksightconfig", "familyconfig", err)
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		if item.Namespace == trigger.Namespace {
			continue
		}
		_, globalNS, err := util.ResolveNamespaceRole(ctx, r.Client, item.Namespace)
		if err != nil {
			continue
		}
		if trigger.Namespace == globalNS && trigger.Name == item.Name {
			requests = append(requests, ctrl.Request{
				NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
			})
		}
	}
	return requests
}

func (r *Reconciler) requestsForNamespaceChange(ctx context.Context, obj client.Object) []ctrl.Request {
	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return nil
	}

	var list v1alpha1.QuickSightConfigList
	if err := r.Client.List(ctx, &list, client.InNamespace(ns.Name)); err != nil {
		util.RecordMapFuncListError(r.Log, "quicksightconfig", "namespace", err)
		return nil
	}

	requests := make([]ctrl.Request, 0, len(list.Items))
	for _, item := range list.Items {
		requests = append(requests, ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: item.Namespace, Name: item.Name},
		})
	}
	return requests
}

// ensureLabel idempotently patches aws.kropath.run/resource-name onto the QuickSightConfig CR.
func (r *Reconciler) ensureLabel(ctx context.Context, cfg *v1alpha1.QuickSightConfig) error {
	if labels := cfg.GetLabels(); labels[resourceNameLabel] == cfg.Name {
		return nil
	}
	patch := []byte(fmt.Sprintf(`{"metadata":{"labels":{%q:%q}}}`, resourceNameLabel, cfg.Name))
	return r.Client.Patch(ctx, cfg, client.RawPatch(types.MergePatchType, patch))
}

func conditionNeedsUpdate(conditions []metav1.Condition, new metav1.Condition) bool {
	for _, c := range conditions {
		if c.Type == new.Type {
			return c.Status != new.Status || c.Reason != new.Reason || c.Message != new.Message
		}
	}
	return true
}

func setCondition(conditions []metav1.Condition, new metav1.Condition) []metav1.Condition {
	for i, c := range conditions {
		if c.Type == new.Type {
			if c.Status == new.Status {
				new.LastTransitionTime = c.LastTransitionTime
			}
			conditions[i] = new
			return conditions
		}
	}
	return append(conditions, new)
}
