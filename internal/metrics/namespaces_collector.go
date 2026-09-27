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

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// namespacesCollectorLabel is this collector's kropath_metrics_collect_errors_total{collector=...} value.
const namespacesCollectorLabel = "namespaceplacement_namespaces"

// These two annotation keys mirror util.GlobalConfigNamespaceAnnotation and
// util.PlacementStatusAnnotation exactly. They are duplicated as literals
// rather than imported because internal/reconciler/util imports
// internal/metrics, so importing util here would cycle.
const (
	globalConfigNamespaceAnnotation = "aws.kropath.run/global-config-namespace"
	placementStatusAnnotation       = "aws.kropath.run/placement-status"
)

// namespacesCollector implements the spec §3.5 collector: how many resource
// namespaces currently carry each placement-status annotation value.
// Governance-only namespaces (no globalConfigNamespaceAnnotation) are never
// evaluated by namespaceplacement and so never carry the status annotation --
// they are excluded here by construction, the same classification the
// reconciler itself uses, without an extra Get per namespace.
type namespacesCollector struct {
	reader client.Reader
}

func init() {
	RegisterCollectorFactory(func(reader, _ client.Reader) prometheus.Collector {
		return &namespacesCollector{reader: reader}
	})
}

func (c *namespacesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- namespaceplacementNamespaces
}

func (c *namespacesCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := CollectContext(context.Background())
	defer cancel()

	var list corev1.NamespaceList
	if err := c.reader.List(ctx, &list); err != nil {
		CollectError(namespacesCollectorLabel)
		return
	}

	counts := map[string]int{}
	for _, ns := range list.Items {
		if ns.Annotations[globalConfigNamespaceAnnotation] == "" {
			continue
		}
		status := ns.Annotations[placementStatusAnnotation]
		if status == "" {
			continue
		}
		counts[status]++
	}

	for status, n := range counts {
		ch <- prometheus.MustNewConstMetric(namespaceplacementNamespaces, prometheus.GaugeValue, float64(n), status)
	}
}
