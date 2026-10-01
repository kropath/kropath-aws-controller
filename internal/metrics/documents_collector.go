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

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// documentsCollectorName is this collector's kropath_metrics_collect_errors_total{collector=...} value (spec §2.5).
const documentsCollectorName = "policydocument_documents"

// documentsCollector reports the number of PolicyDocument objects currently
// in each Ready condition reason, sampled at scrape time (spec §3.1). A
// reason with zero current documents emits no series at all, which is also
// what makes a deleted object's series disappear on the next scrape with no
// separate deletion bookkeeping: this collector recomputes from a fresh List
// every call and never carries state between scrapes.
type documentsCollector struct {
	reader client.Reader
}

func newDocumentsCollector(reader, _ client.Reader) prometheus.Collector {
	return &documentsCollector{reader: reader}
}

func init() {
	RegisterCollectorFactory(newDocumentsCollector)
}

func (c *documentsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- policydocumentDocumentsDesc
}

func (c *documentsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := CollectContext(context.Background())
	defer cancel()

	var docs v1alpha1.PolicyDocumentList
	if err := c.reader.List(ctx, &docs); err != nil {
		CollectError(documentsCollectorName)
		return
	}

	counts := make(map[string]int)
	for i := range docs.Items {
		reason := readyReason(&docs.Items[i])
		if reason == "" {
			continue
		}
		counts[reason]++
	}

	for reason, n := range counts {
		ch <- prometheus.MustNewConstMetric(policydocumentDocumentsDesc, prometheus.GaugeValue, float64(n), reason)
	}
}

// readyReason returns doc's Ready condition Reason, or "" if the document
// has not yet reconciled and carries no Ready condition.
func readyReason(doc *v1alpha1.PolicyDocument) string {
	for _, cond := range doc.Status.Conditions {
		if cond.Type == v1alpha1.ConditionReady {
			return cond.Reason
		}
	}
	return ""
}
