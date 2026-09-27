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
	"strings"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// unresolvedRefsCollectorName is this collector's kropath_metrics_collect_errors_total{collector=...} value (spec §2.5).
const unresolvedRefsCollectorName = "policydocument_unresolved_refs"

// policyDocumentRefKinds mirrors internal/registry/entries.go:85
// (policyDocumentRefGVKs) and policydocument.defaultRefGVKs() -- the six
// installable kinds a PolicyDocument ref may name. This is a third mirror of
// the same closed set: internal/metrics cannot import
// internal/reconciler/policydocument (which imports internal/metrics for the
// ref-resolution counter, so the reverse import would cycle), and this
// collector must bucket the same unbounded ref.Kind value the reconciler
// does, independently, at scrape time (spec §2.4).
var policyDocumentRefKinds = map[string]bool{
	"AWSIAMRole":              true,
	"AWSS3Bucket":             true,
	"AWSLambdaFunction":       true,
	"AWSSQSQueue":             true,
	"AWSKMSKey":               true,
	"AWSSecretsManagerSecret": true,
}

// bucketRefKind buckets kind against policyDocumentRefKinds; anything outside
// the closed set costs one "other" series, not one series per typo.
func bucketRefKind(kind string) string {
	if policyDocumentRefKinds[kind] {
		return kind
	}
	return "other"
}

// bucketRefField buckets a ref's Field against the resolve.go switch's three
// values; anything else is "unsupported" -- the same closed set resolveRef's
// default branch reports (spec §2.4).
func bucketRefField(field string) string {
	switch strings.TrimSpace(field) {
	case "", "predictedArn":
		return "predictedArn"
	case "arn":
		return "arn"
	default:
		return "unsupported"
	}
}

// unresolvedRefsCollector reports, per bucketed {kind, field}, the number of
// policy source references across all PolicyDocument objects that are
// currently unresolved, sampled at scrape time (spec §3.2). It independently
// re-derives resolution status via its own Get against the manager cache --
// the same check resolveRef makes -- because the CR itself does not persist
// which of a document's refs are pending (design is behaviour-preserving:
// same Ready condition, same RequeueAfter, same status.resolvedDocumentJSON,
// no new status field).
type unresolvedRefsCollector struct {
	reader client.Reader
}

func newUnresolvedRefsCollector(reader, _ client.Reader) prometheus.Collector {
	return &unresolvedRefsCollector{reader: reader}
}

func init() {
	RegisterCollectorFactory(newUnresolvedRefsCollector)
}

func (c *unresolvedRefsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- policydocumentUnresolvedRefsDesc
}

type refBucketKey struct {
	kind  string
	field string
}

func (c *unresolvedRefsCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := CollectContext(context.Background())
	defer cancel()

	var docs v1alpha1.PolicyDocumentList
	if err := c.reader.List(ctx, &docs); err != nil {
		CollectError(unresolvedRefsCollectorName)
		return
	}

	counts := make(map[refBucketKey]int)
	for i := range docs.Items {
		doc := &docs.Items[i]
		for _, stmt := range doc.Spec.Statements {
			for _, p := range stmt.Principals {
				c.countIfUnresolved(ctx, doc.Namespace, p.Ref, counts)
			}
			for _, r := range stmt.Resources {
				c.countIfUnresolved(ctx, doc.Namespace, r.Ref, counts)
			}
		}
	}

	for key, n := range counts {
		ch <- prometheus.MustNewConstMetric(policydocumentUnresolvedRefsDesc, prometheus.GaugeValue, float64(n), key.kind, key.field)
	}
}

func (c *unresolvedRefsCollector) countIfUnresolved(ctx context.Context, namespace string, ref *v1alpha1.PolicyRef, counts map[refBucketKey]int) {
	if ref == nil || strings.TrimSpace(ref.Kind) == "" || strings.TrimSpace(ref.Name) == "" {
		return
	}

	if !c.refResolved(ctx, namespace, ref) {
		key := refBucketKey{kind: bucketRefKind(ref.Kind), field: bucketRefField(ref.Field)}
		counts[key]++
	}
}

// refResolved mirrors policydocument.resolveRef's resolution check against
// the collector's own reader. It never returns an error: a bad or
// unreachable single ref costs that ref a "not resolved" count, not a failed
// scrape.
func (c *unresolvedRefsCollector) refResolved(ctx context.Context, namespace string, ref *v1alpha1.PolicyRef) bool {
	field := bucketRefField(ref.Field)
	if field == "unsupported" {
		return false
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: ref.Kind})
	if err := c.reader.Get(ctx, types.NamespacedName{Namespace: namespace, Name: ref.Name}, obj); err != nil {
		if apimeta.IsNoMatchError(err) || client.IgnoreNotFound(err) == nil {
			return false
		}
		return false
	}

	value, found, err := unstructured.NestedString(obj.Object, "status", field)
	if err != nil || !found || strings.TrimSpace(value) == "" {
		return false
	}
	return true
}
