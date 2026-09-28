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

// Package namespaceplacement is the one-per-cluster namespace reconciler that
// evaluates account/region placement and publishes the verdict on the Namespace
// itself (spec §5.6, §6.2). A governance-only namespace (one lacking
// aws.kropath.run/global-config-namespace) is exempt: it is not evaluated and
// carries no placement-status annotation (spec §6.4, AC-8).
//
// This is one of the two publication sites sharing util.ResolvePlacement -- the
// other is every <ResourceFamily>Config reconciler. "One evaluation, two
// publication sites" only holds if both call the same function, which is why
// neither site re-implements the §6.2 rule chain itself.
package namespaceplacement

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/kropath/kropath-controller/internal/metrics"
	"github.com/kropath/kropath-controller/internal/reconciler/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

// Recorder is the subset of record.EventRecorder this reconciler needs, so
// tests can supply a fake without pulling in the full client-go event
// machinery.
type Recorder interface {
	Event(object runtime.Object, eventtype, reason, message string)
}

type Reconciler struct {
	Client   client.Client
	Recorder Recorder
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var ns corev1.Namespace
	if err := r.Client.Get(ctx, req.NamespacedName, &ns); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	role, _, err := util.ResolveNamespaceRole(ctx, r.Client, ns.Name)
	if err != nil {
		// The watch event that triggered this reconcile implies the read just
		// succeeded; a failure here is a transient race. Requeue rather than
		// guessing a role (spec §5.5: a read failure is an error, never a role).
		return ctrl.Result{}, err
	}

	// Governance-only namespaces are exempt from every placement rule: not
	// evaluated, and no placement-status annotation is written (spec §6.4,
	// AC-8). Nothing to heal here even if one was written under a prior release.
	if role == util.RoleGovernanceOnly {
		return ctrl.Result{}, nil
	}

	identity, placementErr, err := util.ResolvePlacement(ctx, r.Client, ns.Name)
	if err != nil {
		return ctrl.Result{}, err
	}

	var newStatus, eventType, eventReason, eventMessage string
	if placementErr != nil {
		newStatus = placementErr.Reason
		eventType = corev1.EventTypeWarning
		eventReason = placementErr.Reason
		eventMessage = placementErr.Message
	} else {
		newStatus = util.PlacementStatusOK
		eventType = corev1.EventTypeNormal
		eventReason = util.ReasonPlacementResolved
		eventMessage = util.PlacementProvenanceMessage(identity, ns.Name)
	}

	prevStatus := ns.Annotations[util.PlacementStatusAnnotation]

	if err := r.applyPlacementStatus(ctx, ns.Name, newStatus); err != nil {
		return ctrl.Result{}, err
	}

	// Events fire only on a verdict transition -- a namespace stuck in the same
	// failure must not produce an Event every resync (spec §6.2). The counter
	// shares this guard so it and the Event never drift apart (spec §2.5).
	if prevStatus != newStatus {
		metrics.NamespacePlacementTransition(newStatus)
		if r.Recorder != nil {
			r.Recorder.Event(&ns, eventType, eventReason, eventMessage)
		}
	}

	return ctrl.Result{}, nil
}

// applyPlacementStatus writes PlacementStatusAnnotation via server-side apply
// under PlacementFieldManager, so a GitOps tool managing the Namespace never
// contends with the controller over the object (spec §6.2). The controller
// never reads this annotation back as input -- it is recomputed from the other
// annotations on every evaluation.
func (r *Reconciler) applyPlacementStatus(ctx context.Context, namespace, status string) error {
	apply := corev1apply.Namespace(namespace).
		WithAnnotations(map[string]string{util.PlacementStatusAnnotation: status})
	return r.Client.Apply(ctx, apply, client.FieldOwner(util.PlacementFieldManager), client.ForceOwnership)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	_, err := r.BuildWithManager(mgr)
	return err
}

func (r *Reconciler) BuildWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Build(r)
}
