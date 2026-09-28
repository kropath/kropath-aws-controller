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

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// tierConfigsCollectorLabel is this collector's kropath_metrics_collect_errors_total{collector=...} value.
const tierConfigsCollectorLabel = "kropathconfigstatus_configs"

// tierConfigsCollector implements the spec §3.4 collector: how many
// KropathConfig objects are currently in each tier-reference reason
// (internal/reconciler/kropathconfigstatus's Reconciled condition), plus the
// family_kinds_unavailable gauge that qualifies it.
//
// This package cannot import internal/reconciler/kropathconfigstatus for its
// ReconciledConditionType / Reason* constants: that package imports
// internal/reconciler/util, which imports internal/metrics, so importing it
// here would cycle. reconciledConditionType (defined in
// withheld_config_collector.go) is the shared literal.
type tierConfigsCollector struct {
	apiReader client.Reader
}

func init() {
	RegisterCollectorFactory(func(_, apiReader client.Reader) prometheus.Collector {
		return &tierConfigsCollector{apiReader: apiReader}
	})
}

func (c *tierConfigsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- kropathconfigstatusConfigs
}

func (c *tierConfigsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := CollectContext(context.Background())
	defer cancel()

	var list v1alpha1.KropathConfigList
	if err := c.apiReader.List(ctx, &list); err != nil {
		CollectError(tierConfigsCollectorLabel)
		return
	}

	unavailable, err := countUnavailableFamilyKinds(ctx, c.apiReader)
	if err != nil {
		CollectError(tierConfigsCollectorLabel)
		return
	}

	counts := map[string]int{}
	for _, kpc := range list.Items {
		for _, cond := range kpc.Status.Conditions {
			if cond.Type == reconciledConditionType {
				counts[cond.Reason]++
				break
			}
		}
	}

	KropathConfigFamilyKindsUnavailable(unavailable)
	for reason, n := range counts {
		ch <- prometheus.MustNewConstMetric(kropathconfigstatusConfigs, prometheus.GaugeValue, float64(n), reason)
	}
}
