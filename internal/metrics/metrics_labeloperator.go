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
	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// labeloperatorWatchedKinds, labeloperatorGroupDiscoveryTotal and
// labeloperatorPatchesTotal answer what kropath_reconciler_active cannot for
// labeloperator (design M-9, spec §1.4, §8.4): whether any kind under a
// provider API group actually has a label-injection controller registered,
// the outcome of every discovery attempt, and the outcome of every patch.
var (
	labeloperatorWatchedKinds = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kropath_labeloperator_watched_kinds",
		Help: "Number of GVKs in this API group that have a label-injection controller registered; unlike kropath_reconciler_active this is 0 when discovery returned nothing.",
	}, []string{"group"})

	labeloperatorGroupDiscoveryTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kropath_labeloperator_group_discovery_total",
		Help: "Total API group discovery attempts by group and outcome; error and empty both leave label injection off for the whole group until the pod restarts.",
	}, []string{"group", "outcome"})

	labeloperatorPatchesTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kropath_labeloperator_patches_total",
		Help: "Total resource-name label patch attempts by API group and outcome.",
	}, []string{"group", "outcome"})
)

func init() {
	ctrlmetrics.Registry.MustRegister(
		labeloperatorWatchedKinds,
		labeloperatorGroupDiscoveryTotal,
		labeloperatorPatchesTotal,
	)
}

// LabelOperatorWatchedKinds records how many GVKs under group currently have
// a label-injection controller registered (spec §8.4).
func LabelOperatorWatchedKinds(group string, n int) {
	labeloperatorWatchedKinds.WithLabelValues(group).Set(float64(n))
}

// LabelOperatorGroupDiscovery records the outcome of one API group discovery
// attempt. outcome must be one of "discovered", "empty", "error" (spec §2.5).
func LabelOperatorGroupDiscovery(group, outcome string) {
	labeloperatorGroupDiscoveryTotal.WithLabelValues(group, outcome).Inc()
}

// LabelOperatorPatch records the outcome of one resource-name label patch
// attempt. outcome must be one of "patched", "not_found", "error" (spec §2.5).
func LabelOperatorPatch(group, outcome string) {
	labeloperatorPatchesTotal.WithLabelValues(group, outcome).Inc()
}
