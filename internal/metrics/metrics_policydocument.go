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

// policydocumentDocumentsDesc and policydocumentUnresolvedRefsDesc are
// collector-emitted (documents_collector.go, unresolved_refs_collector.go),
// not pushed by a typed helper -- they are scrape-time samples over the
// PolicyDocument objects that currently exist (spec §1.2, §3.1, §3.2).
var (
	policydocumentDocumentsDesc = prometheus.NewDesc(
		"kropath_policydocument_documents",
		"Number of PolicyDocument objects currently in each Ready condition reason, sampled at scrape time.",
		[]string{"reason"}, nil)

	policydocumentUnresolvedRefsDesc = prometheus.NewDesc(
		"kropath_policydocument_unresolved_refs",
		"Number of unresolved policy source references across all PolicyDocument objects, by referenced kind and status field, sampled at scrape time.",
		[]string{"kind", "field"}, nil)
)

var (
	policydocumentRefResolutionsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kropath_policydocument_ref_resolutions_total",
		Help: "Total policy source reference resolution attempts by referenced kind, status field and outcome; crd_absent means the referenced kind is not installed, pending means it is installed but the field is not populated yet.",
	}, []string{"kind", "field", "outcome"})

	policydocumentSidConflictsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "kropath_policydocument_sid_conflicts_total",
		Help: "Total times a PolicyDocument merge was rejected because two statements shared a Sid.",
	})
)

func init() {
	ctrlmetrics.Registry.MustRegister(policydocumentRefResolutionsTotal, policydocumentSidConflictsTotal)
}

// PolicyDocumentRefResolution records a policy source reference resolution
// attempt (spec §2.4). kind and field must already be bucketed against their
// closed sets by the caller -- resolveRef's bucketKind for kind, and the
// resolve.go field switch's own branch literal for field -- so a tenant's
// unbounded ref.kind or ref.field value never reaches a label.
func PolicyDocumentRefResolution(kind, field, outcome string) {
	policydocumentRefResolutionsTotal.WithLabelValues(kind, field, outcome).Inc()
}

// PolicyDocumentSidConflict records that a PolicyDocument merge was rejected
// because two statements shared a Sid.
func PolicyDocumentSidConflict() {
	policydocumentSidConflictsTotal.Inc()
}
