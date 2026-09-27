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
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// CollectTimeout bounds every scrape-time collector's List call (design §6,
// spec §3). Derived from the scrape budget, not invented: the fastest
// requeue in the repo is policydocument's defaultRequeueAfter = 10s, and the
// default Prometheus scrape timeout is 10s, so five collectors at this
// timeout each cannot exhaust one scrape even if all five degrade
// simultaneously.
const CollectTimeout = 2 * time.Second

var metricsCollectErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "kropath_metrics_collect_errors_total",
	Help: "Total times a scrape-time collector could not list and therefore emitted no samples for its metric.",
}, []string{"collector"})

func init() {
	ctrlmetrics.Registry.MustRegister(metricsCollectErrorsTotal)
}

// CollectError records that collector could not list its backing kind and
// therefore emitted no samples this scrape. A collector must never emit a
// zero in this case: a zero asserts "no objects are in a bad state", which
// is a worse failure than a gap in the series (design §6, spec §3).
func CollectError(collector string) {
	metricsCollectErrorsTotal.WithLabelValues(collector).Inc()
}

// CollectContext returns a context bounded by CollectTimeout, for a
// scrape-time collector's Collect method to use for its List call.
func CollectContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, CollectTimeout)
}
