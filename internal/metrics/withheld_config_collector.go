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
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// withheldCollectorLabel is this collector's kropath_metrics_collect_errors_total{collector=...} value.
const withheldCollectorLabel = "cascade_effective_config_withheld"

// reconciledConditionType is the Type every <ResourceFamily>Config reconciler
// publishes alongside PlacementResolved (spec §7.2). Its Reason equals the
// PlacementResolved condition's Reason in every case that also publishes a
// PlacementResolved condition, and it is the only one of the two published in
// the governance-only case (util.ResolveFamilyPlacement's RoleGovernanceOnly
// early return sets Reconciled=GlobalTierInput but no PlacementResolved
// condition at all) -- so Reconciled is the condition that is always present
// whenever effectiveConfig is withheld, and is what this collector reads.
const reconciledConditionType = "Reconciled"

// withheldConfigCollector implements the spec §3.3 collector: how many
// <Family>Config objects currently publish no status.effectiveConfig, by
// family and withholding reason.
type withheldConfigCollector struct {
	apiReader client.Reader
}

func init() {
	RegisterCollectorFactory(func(_, apiReader client.Reader) prometheus.Collector {
		return &withheldConfigCollector{apiReader: apiReader}
	})
}

func (c *withheldConfigCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cascadeEffectiveConfigWithheld
}

func (c *withheldConfigCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := CollectContext(context.Background())
	defer cancel()

	counts := map[[2]string]int{}
	for _, fk := range familyConfigKinds {
		list := fk.newList()
		if err := c.apiReader.List(ctx, list); err != nil {
			if apimeta.IsNoMatchError(err) {
				// This family's CRD is not installed -- not a collect error
				// (spec §3), so it simply contributes no series.
				continue
			}
			CollectError(withheldCollectorLabel)
			return
		}

		items, err := apimeta.ExtractList(list)
		if err != nil {
			CollectError(withheldCollectorLabel)
			return
		}
		for _, item := range items {
			reason, withheld := withheldReason(item)
			if !withheld {
				continue
			}
			counts[[2]string{fk.family, reason}]++
		}
	}

	for key, n := range counts {
		ch <- prometheus.MustNewConstMetric(cascadeEffectiveConfigWithheld, prometheus.GaugeValue, float64(n), key[0], key[1])
	}
}

// withheldReason reports the Reconciled condition's Reason for a
// <Family>Config object whose status.effectiveConfig is the Go zero value,
// via reflection: the 57 concrete types differ in their EffectiveConfig
// struct, but all share the same Status.EffectiveConfig / Status.Conditions
// field names. withheld is false when effectiveConfig is non-zero (published
// successfully) or when no Reconciled condition exists yet (not yet
// reconciled) -- neither is the withheld-and-explained state this metric
// reports.
func withheldReason(obj runtime.Object) (reason string, withheld bool) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	status := v.FieldByName("Status")
	effectiveConfig := status.FieldByName("EffectiveConfig")
	if !effectiveConfig.IsValid() || !effectiveConfig.IsZero() {
		return "", false
	}

	conditions, _ := status.FieldByName("Conditions").Interface().([]metav1.Condition)
	for _, cond := range conditions {
		if cond.Type == reconciledConditionType {
			return cond.Reason, cond.Reason != ""
		}
	}
	return "", false
}
