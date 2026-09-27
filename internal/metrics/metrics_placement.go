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

// placementResolutionsTotal and configProfileResolutionsTotal are the two
// shared-helper metrics (design §7.1, spec §1.1): each is emitted from the one
// chokepoint every cascade reconciler and namespaceplacement already calls,
// so a newly-added family is instrumented automatically.
var (
	placementResolutionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kropath_placement_resolutions_total",
		Help: "Total placement resolutions by outcome reason, across every <Family>Config reconciler and namespaceplacement.",
	}, []string{"reason"})

	configProfileResolutionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kropath_config_profile_resolutions_total",
		Help: "Total global-tier config profile lookups by family and outcome reason; ProfileFallthrough means the requested profile was absent and the default was substituted.",
	}, []string{"family", "reason"})
)

func init() {
	ctrlmetrics.Registry.MustRegister(placementResolutionsTotal, configProfileResolutionsTotal)
}

// PlacementResolved records a placement resolution outcome (spec §2.2). reason
// must be one of util.PlacementReasons().
func PlacementResolved(reason string) {
	placementResolutionsTotal.WithLabelValues(reason).Inc()
}

// ConfigProfileResolved records a global-tier config profile lookup outcome
// (spec §2.6). family is the features.All package name (lowercase); reason
// must be one of util.ConfigProfileReasons().
func ConfigProfileResolved(family, reason string) {
	configProfileResolutionsTotal.WithLabelValues(family, reason).Inc()
}
