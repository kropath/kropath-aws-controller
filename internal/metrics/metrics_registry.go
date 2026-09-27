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

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// registryReconcilerPendingSinceTimestampSeconds, registryCRDWatchEventsTotal
// and registryOptionalKindsAttached close the three gaps in the registry's
// existing metrics (design §8.7, spec §1.7): since-when a reconciler has been
// pending, the outcomes the CRD watcher's bare-return arms don't record
// today, and how many optional kinds are actually attached.
var (
	registryReconcilerPendingSinceTimestampSeconds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kropath_registry_reconciler_pending_since_timestamp_seconds",
		Help: "Unix time at which this process first observed the reconciler as pending. The child is deleted on activation, so an absent series means not pending, and the value resets on restart because no process can know when the CRD went missing. Aggregate with min by (package), never max.",
	}, []string{"package"})

	registryCRDWatchEventsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kropath_registry_crd_watch_events_total",
		Help: "Total CRD watch events by outcome, including the store misses, non-servable CRDs and cast failures the error counter does not record.",
	}, []string{"outcome"})

	registryOptionalKindsAttached = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kropath_registry_optional_kinds_attached",
		Help: "Number of optional GVKs currently attached to this reconciler's running controller.",
	}, []string{"package"})
)

func init() {
	ctrlmetrics.Registry.MustRegister(
		registryReconcilerPendingSinceTimestampSeconds,
		registryCRDWatchEventsTotal,
		registryOptionalKindsAttached,
	)
}

// ReconcilerPendingSince records the instant this process first observed
// pkg's reconciler as pending (design M-17, spec §8.7). Must be paired with
// ReconcilerPendingClear on activation, or the series pins the §9 alert on
// forever once it fires.
func ReconcilerPendingSince(pkg string, t time.Time) {
	registryReconcilerPendingSinceTimestampSeconds.WithLabelValues(pkg).Set(float64(t.Unix()))
}

// ReconcilerPendingClear deletes pkg's pending-since child so an absent
// series means "not pending" (M-17). A Set(0) would instead read as
// "pending since the Unix epoch" and fire the §9 alert permanently. Named
// rather than an inline DeleteLabelValues so the registry's startup-gate and
// runtime-activation call sites read symmetrically and the M-17 deletion
// contract is greppable.
func ReconcilerPendingClear(pkg string) {
	registryReconcilerPendingSinceTimestampSeconds.DeleteLabelValues(pkg)
}

// CRDWatchEvent records one CRD-watch event loop outcome. outcome must be one
// of "activated", "store_miss", "not_servable", "cast_failed", "already_active".
func CRDWatchEvent(outcome string) {
	registryCRDWatchEventsTotal.WithLabelValues(outcome).Inc()
}

// OptionalKindsAttached records how many optional GVKs are currently attached
// to pkg's running controller.
func OptionalKindsAttached(pkg string, n int) {
	registryOptionalKindsAttached.WithLabelValues(pkg).Set(float64(n))
}
