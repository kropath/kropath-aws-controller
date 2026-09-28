// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

// outcomeValueSets is the closed set of label values internal/metrics itself
// owns, keyed by metric name (spec §2.1, §2.4, §2.5, §2.7). These are the
// label-family closed sets that do not live on a reason const in a reconciler
// package: outcome, trigger, field, and the bucketed kind set.
//
// kropath_policydocument_ref_resolutions_total carries three label families
// (outcome, kind, field) on one metric, so its entry is their union — the
// §10.2.1 label-value check does not need to know which label a literal
// belongs to, only that it is a member of some closed set.
var outcomeValueSets = map[string][]string{
	"kropath_policydocument_ref_resolutions_total": {
		// outcome (§2.4, M-11)
		"resolved", "pending", "crd_absent", "error",
		// kind — bucketed against internal/registry/entries.go's policyDocumentRefGVKs (§2.4)
		"AWSIAMRole", "AWSS3Bucket", "AWSLambdaFunction", "AWSSQSQueue",
		"AWSKMSKey", "AWSSecretsManagerSecret", "other",
		// field — bucketed against the resolve.go switch (§2.4)
		"predictedArn", "arn", "unsupported",
	},
	"kropath_cascade_mapfunc_errors_total": {
		"kropathconfig", "familyconfig", "namespace",
	},
	"kropath_labeloperator_group_discovery_total": {
		"discovered", "empty", "error",
	},
	"kropath_labeloperator_patches_total": {
		"patched", "not_found", "error",
	},
	"kropath_registry_crd_watch_events_total": {
		"activated", "store_miss", "not_servable", "cast_failed", "already_active",
	},
	"kropath_metrics_collect_errors_total": {
		"policydocument_documents", "policydocument_unresolved_refs",
		"cascade_effective_config_withheld", "kropathconfigstatus_configs",
		"namespaceplacement_namespaces",
	},
	// kropath_namespaceplacement_namespaces{status=...} and
	// kropath_namespaceplacement_transitions_total{to=...} share the same
	// 5-value closed set (spec §2.5): util.ResolvePlacement's 4 failure
	// reasons plus "ok".
	"kropath_namespaceplacement_namespaces": {
		"ok", "TeamAnnotationUnsupported", "MissingAccountAnnotation",
		"InvalidAccountAnnotation", "MissingRegionAnnotation",
	},
	"kropath_namespaceplacement_transitions_total": {
		"ok", "TeamAnnotationUnsupported", "MissingAccountAnnotation",
		"InvalidAccountAnnotation", "MissingRegionAnnotation",
	},
}

// OutcomeValues returns the closed set of label values internal/metrics owns
// for the given metric name (spec §2.1). It returns nil for a metric whose
// closed set lives elsewhere (a reconciler package's Reason consts) or that
// carries no closed-set label at all.
func OutcomeValues(metric string) []string {
	values := outcomeValueSets[metric]
	if values == nil {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

// OutcomeMetrics returns every metric name OutcomeValues recognizes. The
// §10.2.1 label-value check uses this to union every internal/metrics closed
// set without hardcoding the metric list a second time.
func OutcomeMetrics() []string {
	names := make([]string, 0, len(outcomeValueSets))
	for name := range outcomeValueSets {
		names = append(names, name)
	}
	return names
}
