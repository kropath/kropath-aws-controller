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

// namespaceplacementNamespaces is collector-emitted by namespacesCollector
// (spec §3.5), not registered directly: its value only exists at scrape time.
var namespaceplacementNamespaces = prometheus.NewDesc(
	"kropath_namespaceplacement_namespaces",
	"Number of resource namespaces currently carrying each placement-status annotation value, sampled at scrape time.",
	[]string{"status"}, nil)

var namespaceplacementTransitionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "kropath_namespaceplacement_transitions_total",
	Help: "Total placement verdict transitions by the verdict transitioned to; Events fire only on transition, so this is the durable record.",
}, []string{"to"})

func init() {
	ctrlmetrics.Registry.MustRegister(namespaceplacementTransitionsTotal)
}

// NamespacePlacementTransition records a placement verdict transition (spec
// §2.5). to is the verdict transitioned to: util.PlacementStatusOK ("ok") or
// one of the four placement failure reasons.
func NamespacePlacementTransition(to string) {
	namespaceplacementTransitionsTotal.WithLabelValues(to).Inc()
}
