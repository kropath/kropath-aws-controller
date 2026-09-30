// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"runtime"

	"github.com/kropath/kropath-aws-controller/internal/features"
	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

func init() {
	registerMetrics(ctrlmetrics.Registry, Version, GitCommit, runtime.Version(), features.All)
}

// registerMetrics creates and registers the build-info and feature-enabled collectors
// against the given registerer. Extracted for unit-testability.
func registerMetrics(r prometheus.Registerer, ver, gitCommit, goVersion string, all []features.Reconciler) {
	bi := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kropath_build_info",
			Help: "Build information about the kropath-aws-controller binary (constant 1, labels carry the values).",
		},
		[]string{"version", "git_commit", "go_version"},
	)
	bi.With(prometheus.Labels{
		"version":    ver,
		"git_commit": gitCommit,
		"go_version": goVersion,
	}).Set(1)

	fe := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kropath_feature_enabled",
			Help: "One series per reconciler feature built into this binary (constant 1).",
		},
		[]string{"feature", "stability"},
	)
	for _, rec := range all {
		fe.With(prometheus.Labels{
			"feature":   rec.Package,
			"stability": rec.Stability,
		}).Set(1)
	}

	r.MustRegister(bi, fe)
}
