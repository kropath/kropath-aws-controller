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

// cascadeEffectiveConfigWithheld is collector-emitted by withheldConfigCollector
// (spec §3.3), not registered directly: its value only exists at scrape time.
var cascadeEffectiveConfigWithheld = prometheus.NewDesc(
	"kropath_cascade_effective_config_withheld",
	"Number of <Family>Config objects currently publishing no status.effectiveConfig, by family and withholding reason, sampled at scrape time.",
	[]string{"family", "reason"}, nil)

var cascadeMapfuncErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "kropath_cascade_mapfunc_errors_total",
	Help: "Total times a cascade reconciler's watch map function dropped a whole batch of change events after a List or placement failure, by family and trigger.",
}, []string{"family", "trigger"})

func init() {
	ctrlmetrics.Registry.MustRegister(cascadeMapfuncErrorsTotal)
}

// CascadeMapFuncError records that a cascade reconciler's watch map function
// dropped a whole batch of change events (spec §2.7). family is the
// features.All package name (lowercase); trigger is one of "kropathconfig",
// "familyconfig", "namespace".
func CascadeMapFuncError(family, trigger string) {
	cascadeMapfuncErrorsTotal.WithLabelValues(family, trigger).Inc()
}
