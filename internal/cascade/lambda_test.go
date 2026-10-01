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

package cascade_test

import (
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/cascade"
)

// zeroLambdaKropath is a zero-value LambdaKropathSection (absent KropathConfig source).
var zeroLambdaKropath = cascade.LambdaKropathSection{}

// zeroLambdaCfg is a zero-value LambdaConfigSection (absent LambdaConfig source).
var zeroLambdaCfg = cascade.LambdaConfigSection{}

// mergeLambdaAll calls MergeLambdaCascade with all eight inputs.
func mergeLambdaAll(
	globalKropathMandatory,
	localKropathMandatory cascade.LambdaKropathSection,
	globalLambdaCfgMandatory,
	localLambdaCfgMandatory,
	localLambdaCfgDefaults,
	globalLambdaCfgDefaults cascade.LambdaConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.LambdaKropathSection,
) cascade.EffectiveLambdaConfig {
	return cascade.MergeLambdaCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalLambdaCfgMandatory,
		localLambdaCfgMandatory,
		localLambdaCfgDefaults,
		globalLambdaCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeLambdaCascade_MandatoryRuntimeAtL1 — globalKropathConfig.mandatory.lambda.runtime
// set at level 1 propagates to effCfg.mandatory.runtime.
func TestMergeLambdaCascade_MandatoryRuntimeAtL1(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{Runtime: "python3.12"},
		zeroLambdaKropath,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.Runtime != "python3.12" {
		t.Errorf("mandatory.runtime = %q, want python3.12", got.Mandatory.Runtime)
	}
	if got.Defaults.Runtime != "" {
		t.Error("defaults.runtime must not bleed from mandatory")
	}
}

// TestMergeLambdaCascade_MandatoryRuntimeAtL1WinsOverL3 — level-1 KropathConfig wins over
// level-3 LambdaConfig when both set runtime.
func TestMergeLambdaCascade_MandatoryRuntimeAtL1WinsOverL3(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{Runtime: "python3.12"}, // level 1
		zeroLambdaKropath,
		cascade.LambdaConfigSection{Runtime: "nodejs20.x"}, // level 3
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.Runtime != "python3.12" {
		t.Errorf("mandatory.runtime = %q, want python3.12 (level-1 must win)", got.Mandatory.Runtime)
	}
}

// TestMergeLambdaCascade_MandatoryCeilingMemorySize — KropathConfig mandatory memorySize
// at level 1 acts as a ceiling over LambdaConfig level 3.
func TestMergeLambdaCascade_MandatoryCeilingMemorySize(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{MemorySize: 512}, // level 1 ceiling
		zeroLambdaKropath,
		cascade.LambdaConfigSection{MemorySize: 1024}, // level 3 — lower priority
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.MemorySize != 512 {
		t.Errorf("mandatory.memorySize = %d, want 512 (level-1 ceiling must win)", got.Mandatory.MemorySize)
	}
}

// TestMergeLambdaCascade_MandatoryCeilingTimeout — KropathConfig mandatory timeout
// at level 1 wins over LambdaConfig level 3.
func TestMergeLambdaCascade_MandatoryCeilingTimeout(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{Timeout: 30}, // level 1 ceiling
		zeroLambdaKropath,
		cascade.LambdaConfigSection{Timeout: 900}, // level 3 — lower priority
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.Timeout != 30 {
		t.Errorf("mandatory.timeout = %d, want 30 (level-1 ceiling must win)", got.Mandatory.Timeout)
	}
}

// TestMergeLambdaCascade_MandatoryTracingMode — KropathConfig mandatory tracingMode
// at level 1 propagates to effCfg.mandatory.tracingMode.
func TestMergeLambdaCascade_MandatoryTracingMode(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{TracingMode: "Active"},
		zeroLambdaKropath,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.TracingMode != "Active" {
		t.Errorf("mandatory.tracingMode = %q, want Active", got.Mandatory.TracingMode)
	}
}

// TestMergeLambdaCascade_MandatoryKmsKeyArn — KropathConfig mandatory kmsKeyArn at level 1
// propagates to effCfg.mandatory.kmsKeyArn.
func TestMergeLambdaCascade_MandatoryKmsKeyArn(t *testing.T) {
	arn := "arn:aws:kms:ap-southeast-2:123456789012:key/abc-123"
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{KmsKeyArn: arn},
		zeroLambdaKropath,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.KmsKeyArn != arn {
		t.Errorf("mandatory.kmsKeyArn = %q, want %q", got.Mandatory.KmsKeyArn, arn)
	}
}

// TestMergeLambdaCascade_DefaultsRuntimeAtL6 — localLambdaConfig.defaults.runtime set
// at level 6 resolves to effCfg.defaults.runtime.
func TestMergeLambdaCascade_DefaultsRuntimeAtL6(t *testing.T) {
	got := mergeLambdaAll(
		zeroLambdaKropath,
		zeroLambdaKropath,
		zeroLambdaCfg,
		zeroLambdaCfg,
		cascade.LambdaConfigSection{Runtime: "nodejs20.x"}, // level 6
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.Runtime != "" {
		t.Errorf("mandatory.runtime = %q, want empty (only defaults set)", got.Mandatory.Runtime)
	}
	if got.Defaults.Runtime != "nodejs20.x" {
		t.Errorf("defaults.runtime = %q, want nodejs20.x", got.Defaults.Runtime)
	}
}

// TestMergeLambdaCascade_DefaultsFallbackToL9 — globalKropathConfig.defaults.lambda.runtime
// at level 9 is used when no higher-priority defaults source sets runtime.
func TestMergeLambdaCascade_DefaultsFallbackToL9(t *testing.T) {
	got := mergeLambdaAll(
		zeroLambdaKropath,
		zeroLambdaKropath,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		cascade.LambdaKropathSection{Runtime: "python3.12"}, // level 9
	)

	if got.Defaults.Runtime != "python3.12" {
		t.Errorf("defaults.runtime = %q, want python3.12 (level-9 fallback)", got.Defaults.Runtime)
	}
}

// TestMergeLambdaCascade_DefaultsL6WinsOverL9 — level 6 wins over level 9 for defaults.
func TestMergeLambdaCascade_DefaultsL6WinsOverL9(t *testing.T) {
	got := mergeLambdaAll(
		zeroLambdaKropath,
		zeroLambdaKropath,
		zeroLambdaCfg,
		zeroLambdaCfg,
		cascade.LambdaConfigSection{Runtime: "nodejs20.x"},  // level 6 wins
		zeroLambdaCfg,
		zeroLambdaKropath,
		cascade.LambdaKropathSection{Runtime: "python3.12"}, // level 9
	)

	if got.Defaults.Runtime != "nodejs20.x" {
		t.Errorf("defaults.runtime = %q, want nodejs20.x (level-6 must win over level-9)", got.Defaults.Runtime)
	}
}

// TestMergeLambdaCascade_NamingTemplateLambdaConfigOnly — namingTemplate flows only from
// LambdaConfig levels (not from KropathConfig lambda section).
func TestMergeLambdaCascade_NamingTemplateLambdaConfigOnly(t *testing.T) {
	got := mergeLambdaAll(
		zeroLambdaKropath,
		zeroLambdaKropath,
		cascade.LambdaConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 3
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("mandatory.namingTemplate = %q, want {namespace}-{name}", got.Mandatory.NamingTemplate)
	}
}

// TestMergeLambdaCascade_NamingTemplateDefaultsAtL6 — defaults namingTemplate from level 6.
func TestMergeLambdaCascade_NamingTemplateDefaultsAtL6(t *testing.T) {
	got := mergeLambdaAll(
		zeroLambdaKropath,
		zeroLambdaKropath,
		zeroLambdaCfg,
		zeroLambdaCfg,
		cascade.LambdaConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 6
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.NamingTemplate != "" {
		t.Error("mandatory.namingTemplate must be empty when only defaults level set")
	}
	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("defaults.namingTemplate = %q, want {namespace}-{name}", got.Defaults.NamingTemplate)
	}
}

// TestMergeLambdaCascade_TagsMergeAC10 — AC-10: mandatory.tags merges KropathConfig org-wide
// tags (level 1) with LambdaConfig mandatory tags (level 3/4). Both must appear in the result.
func TestMergeLambdaCascade_TagsMergeAC10(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{Tags: map[string]string{"cost-centre": "infra"}},       // level 1
		zeroLambdaKropath,
		cascade.LambdaConfigSection{Tags: map[string]string{"service": "lambda"}}, // level 3
		zeroLambdaCfg,
		cascade.LambdaConfigSection{Tags: map[string]string{"team": "platform"}},  // level 6 (defaults)
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["service"] != "lambda" {
		t.Errorf("mandatory.tags[service] = %q, want lambda", got.Mandatory.Tags["service"])
	}
	if got.Defaults.Tags["team"] != "platform" {
		t.Errorf("defaults.tags[team] = %q, want platform", got.Defaults.Tags["team"])
	}
}

// TestMergeLambdaCascade_TagsKropathWinsOnConflict — when KropathConfig (L1) and LambdaConfig (L3)
// both set the same tag key, KropathConfig mandatory wins.
func TestMergeLambdaCascade_TagsKropathWinsOnConflict(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{Tags: map[string]string{"env": "prod"}},   // level 1
		zeroLambdaKropath,
		cascade.LambdaConfigSection{Tags: map[string]string{"env": "staging"}}, // level 3
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("mandatory.tags[env] = %q, want prod (level-1 must win on key conflict)", got.Mandatory.Tags["env"])
	}
}

// TestMergeLambdaCascade_SyncedLabelsAC11 — AC-11: mandatory.syncedLabels from LambdaConfig (L3)
// and KropathConfig (L1) are merged; both must appear in the result.
func TestMergeLambdaCascade_SyncedLabelsAC11(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{SyncedLabels: map[string]string{"org-class": "internal"}}, // level 1
		zeroLambdaKropath,
		cascade.LambdaConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // level 3
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("mandatory.syncedLabels[data-class] = %q, want internal", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["org-class"] != "internal" {
		t.Errorf("mandatory.syncedLabels[org-class] = %q, want internal", got.Mandatory.SyncedLabels["org-class"])
	}
}

// TestMergeLambdaCascade_SyncedAnnotationsMerge — syncedAnnotations merge across sources.
func TestMergeLambdaCascade_SyncedAnnotationsMerge(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{SyncedAnnotations: map[string]string{"org-policy": "strict"}},
		zeroLambdaKropath,
		cascade.LambdaConfigSection{SyncedAnnotations: map[string]string{"lambda-policy": "secure"}},
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.SyncedAnnotations["org-policy"] != "strict" {
		t.Errorf("mandatory.syncedAnnotations[org-policy] = %q, want strict", got.Mandatory.SyncedAnnotations["org-policy"])
	}
	if got.Mandatory.SyncedAnnotations["lambda-policy"] != "secure" {
		t.Errorf("mandatory.syncedAnnotations[lambda-policy] = %q, want secure", got.Mandatory.SyncedAnnotations["lambda-policy"])
	}
}

// TestMergeLambdaCascade_AllAbsent — when all sources are zero, effectiveConfig fields
// are all zero (permissive; no governance enforced).
func TestMergeLambdaCascade_AllAbsent(t *testing.T) {
	got := mergeLambdaAll(
		zeroLambdaKropath, zeroLambdaKropath,
		zeroLambdaCfg, zeroLambdaCfg, zeroLambdaCfg, zeroLambdaCfg,
		zeroLambdaKropath, zeroLambdaKropath,
	)

	if got.Mandatory.Runtime != "" {
		t.Errorf("all-absent: mandatory.runtime = %q, want empty", got.Mandatory.Runtime)
	}
	if got.Mandatory.MemorySize != 0 {
		t.Errorf("all-absent: mandatory.memorySize = %d, want 0", got.Mandatory.MemorySize)
	}
	if got.Mandatory.Timeout != 0 {
		t.Errorf("all-absent: mandatory.timeout = %d, want 0", got.Mandatory.Timeout)
	}
	if got.Mandatory.TracingMode != "" {
		t.Errorf("all-absent: mandatory.tracingMode = %q, want empty", got.Mandatory.TracingMode)
	}
	if got.Mandatory.KmsKeyArn != "" {
		t.Errorf("all-absent: mandatory.kmsKeyArn = %q, want empty", got.Mandatory.KmsKeyArn)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.Runtime != "" {
		t.Errorf("all-absent: defaults.runtime = %q, want empty", got.Defaults.Runtime)
	}
	if got.Defaults.MemorySize != 0 {
		t.Errorf("all-absent: defaults.memorySize = %d, want 0", got.Defaults.MemorySize)
	}
	if got.Defaults.Timeout != 0 {
		t.Errorf("all-absent: defaults.timeout = %d, want 0", got.Defaults.Timeout)
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeLambdaCascade_MandatoryIsolatedFromDefaults — mandatory fields must not
// bleed into defaults and vice versa.
func TestMergeLambdaCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeLambdaAll(
		cascade.LambdaKropathSection{Runtime: "python3.12", TracingMode: "Active"}, // level 1 mandatory
		zeroLambdaKropath,
		zeroLambdaCfg,
		zeroLambdaCfg,
		cascade.LambdaConfigSection{Runtime: "nodejs20.x", TracingMode: "PassThrough"}, // level 6 defaults
		zeroLambdaCfg,
		zeroLambdaKropath,
		zeroLambdaKropath,
	)

	if got.Mandatory.Runtime != "python3.12" {
		t.Errorf("mandatory.runtime = %q, want python3.12", got.Mandatory.Runtime)
	}
	if got.Mandatory.TracingMode != "Active" {
		t.Errorf("mandatory.tracingMode = %q, want Active", got.Mandatory.TracingMode)
	}
	if got.Defaults.Runtime != "nodejs20.x" {
		t.Errorf("defaults.runtime = %q, want nodejs20.x", got.Defaults.Runtime)
	}
	if got.Defaults.TracingMode != "PassThrough" {
		t.Errorf("defaults.tracingMode = %q, want PassThrough", got.Defaults.TracingMode)
	}
}

// TestMergeLambdaCascade_CascadeOrder — verifies the full mandatory cascade order
// (L1 > L2 > L3 > L4) and defaults cascade order (L6 > L7 > L8 > L9) for runtime.
func TestMergeLambdaCascade_CascadeOrder(t *testing.T) {
	mandatoryCases := []struct {
		name   string
		l1, l2 cascade.LambdaKropathSection
		l3, l4 cascade.LambdaConfigSection
		want   string
	}{
		{
			name: "L1-wins",
			l1:   cascade.LambdaKropathSection{Runtime: "L1"},
			l2:   cascade.LambdaKropathSection{Runtime: "L2"},
			l3:   cascade.LambdaConfigSection{Runtime: "L3"},
			l4:   cascade.LambdaConfigSection{Runtime: "L4"},
			want: "L1",
		},
		{
			name: "L2-wins-when-L1-absent",
			l1:   zeroLambdaKropath,
			l2:   cascade.LambdaKropathSection{Runtime: "L2"},
			l3:   cascade.LambdaConfigSection{Runtime: "L3"},
			l4:   cascade.LambdaConfigSection{Runtime: "L4"},
			want: "L2",
		},
		{
			name: "L3-wins-when-L1-L2-absent",
			l1:   zeroLambdaKropath,
			l2:   zeroLambdaKropath,
			l3:   cascade.LambdaConfigSection{Runtime: "L3"},
			l4:   cascade.LambdaConfigSection{Runtime: "L4"},
			want: "L3",
		},
		{
			name: "L4-wins-when-L1-L2-L3-absent",
			l1:   zeroLambdaKropath,
			l2:   zeroLambdaKropath,
			l3:   zeroLambdaCfg,
			l4:   cascade.LambdaConfigSection{Runtime: "L4"},
			want: "L4",
		},
	}

	for _, tc := range mandatoryCases {
		t.Run("mandatory-"+tc.name, func(t *testing.T) {
			got := mergeLambdaAll(
				tc.l1, tc.l2,
				tc.l3, tc.l4,
				zeroLambdaCfg, zeroLambdaCfg,
				zeroLambdaKropath, zeroLambdaKropath,
			)
			if got.Mandatory.Runtime != tc.want {
				t.Errorf("mandatory.runtime = %q, want %q", got.Mandatory.Runtime, tc.want)
			}
		})
	}

	defaultsCases := []struct {
		name   string
		l6, l7 cascade.LambdaConfigSection
		l8, l9 cascade.LambdaKropathSection
		want   string
	}{
		{
			name: "L6-wins",
			l6:   cascade.LambdaConfigSection{Runtime: "L6"},
			l7:   cascade.LambdaConfigSection{Runtime: "L7"},
			l8:   cascade.LambdaKropathSection{Runtime: "L8"},
			l9:   cascade.LambdaKropathSection{Runtime: "L9"},
			want: "L6",
		},
		{
			name: "L7-wins-when-L6-absent",
			l6:   zeroLambdaCfg,
			l7:   cascade.LambdaConfigSection{Runtime: "L7"},
			l8:   cascade.LambdaKropathSection{Runtime: "L8"},
			l9:   cascade.LambdaKropathSection{Runtime: "L9"},
			want: "L7",
		},
		{
			name: "L8-wins-when-L6-L7-absent",
			l6:   zeroLambdaCfg,
			l7:   zeroLambdaCfg,
			l8:   cascade.LambdaKropathSection{Runtime: "L8"},
			l9:   cascade.LambdaKropathSection{Runtime: "L9"},
			want: "L8",
		},
		{
			name: "L9-fallback",
			l6:   zeroLambdaCfg,
			l7:   zeroLambdaCfg,
			l8:   zeroLambdaKropath,
			l9:   cascade.LambdaKropathSection{Runtime: "L9"},
			want: "L9",
		},
	}

	for _, tc := range defaultsCases {
		t.Run("defaults-"+tc.name, func(t *testing.T) {
			got := mergeLambdaAll(
				zeroLambdaKropath, zeroLambdaKropath,
				zeroLambdaCfg, zeroLambdaCfg,
				tc.l6, tc.l7, tc.l8, tc.l9,
			)
			if got.Defaults.Runtime != tc.want {
				t.Errorf("defaults.runtime = %q, want %q", got.Defaults.Runtime, tc.want)
			}
		})
	}
}
