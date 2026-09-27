// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	reconcilerActive = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kropath_reconciler_active",
		Help: "1 if the reconciler is active (all required CRDs served), 0 if pending. For reconcilers with optional or wildcard kinds this does not imply any kind is being watched — see kropath_labeloperator_watched_kinds.",
	}, []string{"package"})

	reconcilerMissingKinds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kropath_reconciler_missing_kinds",
		Help: "Number of required GVKs not yet served for this reconciler.",
	}, []string{"package"})

	reconcilerActivationsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kropath_reconciler_activations_total",
		Help: "Total number of times a reconciler was activated by the CRD watcher.",
	}, []string{"package"})

	crdWatchErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kropath_crd_watch_errors_total",
		Help: "Total number of runtime CRD watch errors by reason.",
	}, []string{"reason"})
)

func init() {
	ctrlmetrics.Registry.MustRegister(
		reconcilerActive,
		reconcilerMissingKinds,
		reconcilerActivationsTotal,
		crdWatchErrorsTotal,
	)
}
