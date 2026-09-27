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

// kropathconfigstatusConfigs is collector-emitted by tierConfigsCollector
// (spec §3.4), not registered directly: its value only exists at scrape time.
var kropathconfigstatusConfigs = prometheus.NewDesc(
	"kropath_kropathconfigstatus_configs",
	`Number of KropathConfig objects currently in each tier-reference reason, sampled at scrape time; reason="Unreferenced" means no <Family>Config resolves this config.`,
	[]string{"reason"}, nil)

var kropathconfigstatusFamilyKindsUnavailable = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "kropath_kropathconfigstatus_family_kinds_unavailable",
	Help: "Number of <Family>Config kinds whose CRD was absent at the last consumer count, which can understate consumers and flip a KropathConfig to Unreferenced.",
})

func init() {
	ctrlmetrics.Registry.MustRegister(kropathconfigstatusFamilyKindsUnavailable)
}

// KropathConfigFamilyKindsUnavailable sets the number of <Family>Config kinds
// whose CRD was absent (NoKindMatchError) on the tierConfigsCollector's last
// pass (spec §3.4). It is a plain gauge rather than collector-emitted because
// the count is derived as a side effect of that collector's Collect, not from
// an independent list of its own.
func KropathConfigFamilyKindsUnavailable(n int) {
	kropathconfigstatusFamilyKindsUnavailable.Set(float64(n))
}
