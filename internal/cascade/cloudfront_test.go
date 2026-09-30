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

// zeroKropathCF is a zero-value CloudFrontKropathSection (absent source).
var zeroKropathCF = cascade.CloudFrontKropathSection{}

// zeroCFCfg is a zero-value CloudFrontConfigSection (absent source).
var zeroCFCfg = cascade.CloudFrontConfigSection{}

// mergeCFAll calls MergeCloudFrontCascade with all eight inputs.
func mergeCFAll(
	globalKropathMandatory,
	localKropathMandatory cascade.CloudFrontKropathSection,
	globalCFCfgMandatory,
	localCFCfgMandatory,
	localCFCfgDefaults,
	globalCFCfgDefaults cascade.CloudFrontConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.CloudFrontKropathSection,
) cascade.EffectiveCloudFrontConfig {
	return cascade.MergeCloudFrontCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalCFCfgMandatory,
		localCFCfgMandatory,
		localCFCfgDefaults,
		globalCFCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeCloudFrontCascade_AllAbsent — when all sources are zero, effectiveConfig is
// all-zero (permissive; no governance enforced).
func TestMergeCloudFrontCascade_AllAbsent(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF, zeroKropathCF,
		zeroCFCfg, zeroCFCfg, zeroCFCfg, zeroCFCfg,
		zeroKropathCF, zeroKropathCF,
	)

	if got.Mandatory.ViewerProtocolPolicy != "" {
		t.Errorf("all-absent: mandatory.viewerProtocolPolicy = %q, want empty", got.Mandatory.ViewerProtocolPolicy)
	}
	if got.Mandatory.WebACLRequired {
		t.Errorf("all-absent: mandatory.webACLRequired = true, want false")
	}
	if got.Mandatory.LoggingEnabled {
		t.Errorf("all-absent: mandatory.loggingEnabled = true, want false")
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("all-absent: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("all-absent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
	if got.Defaults.ViewerProtocolPolicy != "" {
		t.Errorf("all-absent: defaults.viewerProtocolPolicy = %q, want empty", got.Defaults.ViewerProtocolPolicy)
	}
	if got.Defaults.WebACLRequired {
		t.Errorf("all-absent: defaults.webACLRequired = true, want false")
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("all-absent: defaults.tags = %v, want empty", got.Defaults.Tags)
	}
}

// TestMergeCloudFrontCascade_AC1 — globalKropathConfig.mandatory.cloudfront.viewerProtocolPolicy
// at level 1 propagates to effCfg.mandatory.viewerProtocolPolicy.
func TestMergeCloudFrontCascade_AC1(t *testing.T) {
	got := mergeCFAll(
		cascade.CloudFrontKropathSection{ViewerProtocolPolicy: "https-only"}, // level 1
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.ViewerProtocolPolicy != "https-only" {
		t.Errorf("AC-1: mandatory.viewerProtocolPolicy = %q, want https-only (level 1 wins)", got.Mandatory.ViewerProtocolPolicy)
	}
	if got.Defaults.ViewerProtocolPolicy != "" {
		t.Errorf("AC-1: defaults.viewerProtocolPolicy = %q, must not bleed from mandatory", got.Defaults.ViewerProtocolPolicy)
	}
}

// TestMergeCloudFrontCascade_AC2 — localKropathConfig.mandatory.cloudfront.viewerProtocolPolicy
// at level 2 wins when level 1 is empty.
func TestMergeCloudFrontCascade_AC2(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		cascade.CloudFrontKropathSection{ViewerProtocolPolicy: "redirect-to-https"}, // level 2
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.ViewerProtocolPolicy != "redirect-to-https" {
		t.Errorf("AC-2: mandatory.viewerProtocolPolicy = %q, want redirect-to-https (level 2 wins)", got.Mandatory.ViewerProtocolPolicy)
	}
}

// TestMergeCloudFrontCascade_AC3 — globalCFConfig.mandatory.viewerProtocolPolicy at level 3
// wins when levels 1-2 are empty.
func TestMergeCloudFrontCascade_AC3(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "https-only"}, // level 3
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.ViewerProtocolPolicy != "https-only" {
		t.Errorf("AC-3: mandatory.viewerProtocolPolicy = %q, want https-only (level 3 wins)", got.Mandatory.ViewerProtocolPolicy)
	}
}

// TestMergeCloudFrontCascade_AC4 — globalKropathConfig.mandatory.cloudfront.webACLRequired=true
// at level 1 propagates to effCfg.mandatory.webACLRequired.
func TestMergeCloudFrontCascade_AC4(t *testing.T) {
	got := mergeCFAll(
		cascade.CloudFrontKropathSection{WebACLRequired: true}, // level 1
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if !got.Mandatory.WebACLRequired {
		t.Errorf("AC-4: mandatory.webACLRequired = false, want true (level 1 wins)")
	}
	if got.Defaults.WebACLRequired {
		t.Errorf("AC-4: defaults.webACLRequired = true, must not bleed from mandatory")
	}
}

// TestMergeCloudFrontCascade_AC5 — globalCFConfig.mandatory.webACLRequired=true at level 3
// wins when levels 1-2 are false.
func TestMergeCloudFrontCascade_AC5(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		cascade.CloudFrontConfigSection{WebACLRequired: true}, // level 3
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if !got.Mandatory.WebACLRequired {
		t.Errorf("AC-5: mandatory.webACLRequired = false, want true (level 3 wins)")
	}
}

// TestMergeCloudFrontCascade_AC6 — globalKropathConfig.mandatory.cloudfront.loggingEnabled=true
// at level 1 propagates to effCfg.mandatory.loggingEnabled.
func TestMergeCloudFrontCascade_AC6(t *testing.T) {
	got := mergeCFAll(
		cascade.CloudFrontKropathSection{LoggingEnabled: true}, // level 1
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if !got.Mandatory.LoggingEnabled {
		t.Errorf("AC-6: mandatory.loggingEnabled = false, want true (level 1 wins)")
	}
	if got.Defaults.LoggingEnabled {
		t.Errorf("AC-6: defaults.loggingEnabled = true, must not bleed from mandatory")
	}
}

// TestMergeCloudFrontCascade_AC7 — httpVersion and sslSupportMethod are CloudFrontConfig-only;
// KropathConfig levels do not contribute these fields.
func TestMergeCloudFrontCascade_AC7(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		cascade.CloudFrontConfigSection{HttpVersion: "http2and3", SslSupportMethod: "sni-only"}, // level 3
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.HttpVersion != "http2and3" {
		t.Errorf("AC-7: mandatory.httpVersion = %q, want http2and3 (level 3)", got.Mandatory.HttpVersion)
	}
	if got.Mandatory.SslSupportMethod != "sni-only" {
		t.Errorf("AC-7: mandatory.sslSupportMethod = %q, want sni-only (level 3)", got.Mandatory.SslSupportMethod)
	}
}

// TestMergeCloudFrontCascade_AC8 — priceClass, geoRestrictionType, oacSigningBehavior are
// CloudFrontConfig-only; level 3 propagates them.
func TestMergeCloudFrontCascade_AC8(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		cascade.CloudFrontConfigSection{
			PriceClass:         "PriceClass_100",
			GeoRestrictionType: "whitelist",
			OacSigningBehavior: "always",
		}, // level 3
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.PriceClass != "PriceClass_100" {
		t.Errorf("AC-8: mandatory.priceClass = %q, want PriceClass_100 (level 3)", got.Mandatory.PriceClass)
	}
	if got.Mandatory.GeoRestrictionType != "whitelist" {
		t.Errorf("AC-8: mandatory.geoRestrictionType = %q, want whitelist (level 3)", got.Mandatory.GeoRestrictionType)
	}
	if got.Mandatory.OacSigningBehavior != "always" {
		t.Errorf("AC-8: mandatory.oacSigningBehavior = %q, want always (level 3)", got.Mandatory.OacSigningBehavior)
	}
}

// TestMergeCloudFrontCascade_AC9 — namingTemplate and loggingBucket are CloudFrontConfig-only;
// level 3 propagates them.
func TestMergeCloudFrontCascade_AC9(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		cascade.CloudFrontConfigSection{
			NamingTemplate: "{namespace}-{name}",
			LoggingBucket:  "org-logs.s3.amazonaws.com",
		}, // level 3
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-9: mandatory.namingTemplate = %q, want {namespace}-{name} (level 3)", got.Mandatory.NamingTemplate)
	}
	if got.Mandatory.LoggingBucket != "org-logs.s3.amazonaws.com" {
		t.Errorf("AC-9: mandatory.loggingBucket = %q, want org-logs.s3.amazonaws.com (level 3)", got.Mandatory.LoggingBucket)
	}
}

// TestMergeCloudFrontCascade_AC10 — defaults: localCFConfig.defaults.viewerProtocolPolicy at
// level 6 propagates; mandatory stays empty.
func TestMergeCloudFrontCascade_AC10(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "redirect-to-https"}, // level 6
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.ViewerProtocolPolicy != "" {
		t.Errorf("AC-10: mandatory.viewerProtocolPolicy = %q, want empty", got.Mandatory.ViewerProtocolPolicy)
	}
	if got.Defaults.ViewerProtocolPolicy != "redirect-to-https" {
		t.Errorf("AC-10: defaults.viewerProtocolPolicy = %q, want redirect-to-https (level 6)", got.Defaults.ViewerProtocolPolicy)
	}
}

// TestMergeCloudFrontCascade_AC11 — defaults: globalCFConfig.defaults.httpVersion at level 7
// propagates when level 6 is empty.
func TestMergeCloudFrontCascade_AC11(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		cascade.CloudFrontConfigSection{HttpVersion: "http2", PriceClass: "PriceClass_All", OacSigningBehavior: "always"}, // level 7
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Defaults.HttpVersion != "http2" {
		t.Errorf("AC-11: defaults.httpVersion = %q, want http2 (level 7)", got.Defaults.HttpVersion)
	}
	if got.Defaults.PriceClass != "PriceClass_All" {
		t.Errorf("AC-11: defaults.priceClass = %q, want PriceClass_All (level 7)", got.Defaults.PriceClass)
	}
	if got.Defaults.OacSigningBehavior != "always" {
		t.Errorf("AC-11: defaults.oacSigningBehavior = %q, want always (level 7)", got.Defaults.OacSigningBehavior)
	}
}

// TestMergeCloudFrontCascade_AC12 — defaults: globalKropathConfig.defaults.cloudfront.webACLRequired=true
// at level 9 propagates when levels 6-8 are false.
func TestMergeCloudFrontCascade_AC12(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		cascade.CloudFrontKropathSection{WebACLRequired: true}, // level 9
	)

	if !got.Defaults.WebACLRequired {
		t.Errorf("AC-12: defaults.webACLRequired = false, want true (level 9 wins when 6-8 are false)")
	}
}

// TestMergeCloudFrontCascade_AC13 — defaults: globalKropathConfig.defaults.cloudfront.loggingEnabled=true
// at level 9 propagates; mandatory stays false.
func TestMergeCloudFrontCascade_AC13(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		cascade.CloudFrontKropathSection{LoggingEnabled: true}, // level 9
	)

	if got.Mandatory.LoggingEnabled {
		t.Errorf("AC-13: mandatory.loggingEnabled = true, must not bleed from defaults")
	}
	if !got.Defaults.LoggingEnabled {
		t.Errorf("AC-13: defaults.loggingEnabled = false, want true (level 9)")
	}
}

// TestMergeCloudFrontCascade_MandatoryPriorityOrder — verifies mandatory priority order
// for viewerProtocolPolicy: level 1 > 2 > 3 > 4.
func TestMergeCloudFrontCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.CloudFrontKropathSection
		localKropathMandatory  cascade.CloudFrontKropathSection
		globalCFCfgMandatory   cascade.CloudFrontConfigSection
		localCFCfgMandatory    cascade.CloudFrontConfigSection
		wantPolicy             string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.CloudFrontKropathSection{ViewerProtocolPolicy: "level1"},
			localKropathMandatory:  cascade.CloudFrontKropathSection{ViewerProtocolPolicy: "level2"},
			globalCFCfgMandatory:   cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "level3"},
			localCFCfgMandatory:    cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "level4"},
			wantPolicy:             "level1",
		},
		{
			name:                   "level2-wins-when-1-absent",
			globalKropathMandatory: zeroKropathCF,
			localKropathMandatory:  cascade.CloudFrontKropathSection{ViewerProtocolPolicy: "level2"},
			globalCFCfgMandatory:   cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "level3"},
			localCFCfgMandatory:    cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "level4"},
			wantPolicy:             "level2",
		},
		{
			name:                   "level3-wins-when-1-2-absent",
			globalKropathMandatory: zeroKropathCF,
			localKropathMandatory:  zeroKropathCF,
			globalCFCfgMandatory:   cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "level3"},
			localCFCfgMandatory:    cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "level4"},
			wantPolicy:             "level3",
		},
		{
			name:                   "level4-wins-when-1-2-3-absent",
			globalKropathMandatory: zeroKropathCF,
			localKropathMandatory:  zeroKropathCF,
			globalCFCfgMandatory:   zeroCFCfg,
			localCFCfgMandatory:    cascade.CloudFrontConfigSection{ViewerProtocolPolicy: "level4"},
			wantPolicy:             "level4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCFAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalCFCfgMandatory,
				tc.localCFCfgMandatory,
				zeroCFCfg,
				zeroCFCfg,
				zeroKropathCF,
				zeroKropathCF,
			)
			if got.Mandatory.ViewerProtocolPolicy != tc.wantPolicy {
				t.Errorf("mandatory.viewerProtocolPolicy = %q, want %q", got.Mandatory.ViewerProtocolPolicy, tc.wantPolicy)
			}
		})
	}
}

// TestMergeCloudFrontCascade_DefaultsPriorityOrder — verifies defaults priority order
// for minimumProtocolVersion: level 6 > 7 > 8 > 9.
func TestMergeCloudFrontCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localCFCfgDefaults    cascade.CloudFrontConfigSection
		globalCFCfgDefaults   cascade.CloudFrontConfigSection
		localKropathDefaults  cascade.CloudFrontKropathSection
		globalKropathDefaults cascade.CloudFrontKropathSection
		wantVersion           string
	}{
		{
			name:                  "level6-wins",
			localCFCfgDefaults:    cascade.CloudFrontConfigSection{MinimumProtocolVersion: "TLSv1.2_2021"},
			globalCFCfgDefaults:   cascade.CloudFrontConfigSection{MinimumProtocolVersion: "TLSv1.2_2019"},
			localKropathDefaults:  cascade.CloudFrontKropathSection{MinimumProtocolVersion: "TLSv1.2_2016"},
			globalKropathDefaults: cascade.CloudFrontKropathSection{MinimumProtocolVersion: "TLSv1"},
			wantVersion:           "TLSv1.2_2021",
		},
		{
			name:                  "level7-wins-when-6-absent",
			localCFCfgDefaults:    zeroCFCfg,
			globalCFCfgDefaults:   cascade.CloudFrontConfigSection{MinimumProtocolVersion: "TLSv1.2_2019"},
			localKropathDefaults:  cascade.CloudFrontKropathSection{MinimumProtocolVersion: "TLSv1.2_2016"},
			globalKropathDefaults: cascade.CloudFrontKropathSection{MinimumProtocolVersion: "TLSv1"},
			wantVersion:           "TLSv1.2_2019",
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localCFCfgDefaults:    zeroCFCfg,
			globalCFCfgDefaults:   zeroCFCfg,
			localKropathDefaults:  cascade.CloudFrontKropathSection{MinimumProtocolVersion: "TLSv1.2_2016"},
			globalKropathDefaults: cascade.CloudFrontKropathSection{MinimumProtocolVersion: "TLSv1"},
			wantVersion:           "TLSv1.2_2016",
		},
		{
			name:                  "level9-wins-when-6-7-8-absent",
			localCFCfgDefaults:    zeroCFCfg,
			globalCFCfgDefaults:   zeroCFCfg,
			localKropathDefaults:  zeroKropathCF,
			globalKropathDefaults: cascade.CloudFrontKropathSection{MinimumProtocolVersion: "TLSv1"},
			wantVersion:           "TLSv1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCFAll(
				zeroKropathCF,
				zeroKropathCF,
				zeroCFCfg,
				zeroCFCfg,
				tc.localCFCfgDefaults,
				tc.globalCFCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.MinimumProtocolVersion != tc.wantVersion {
				t.Errorf("defaults.minimumProtocolVersion = %q, want %q", got.Defaults.MinimumProtocolVersion, tc.wantVersion)
			}
		})
	}
}

// TestMergeCloudFrontCascade_TagsUnion — KropathConfig.mandatory.tags and
// CFConfig.mandatory.tags are union-merged into effCfg.mandatory.tags.
func TestMergeCloudFrontCascade_TagsUnion(t *testing.T) {
	got := mergeCFAll(
		cascade.CloudFrontKropathSection{Tags: map[string]string{"cost-centre": "platform"}}, // level 1
		zeroKropathCF,
		zeroCFCfg,
		cascade.CloudFrontConfigSection{Tags: map[string]string{"cdn-class": "production"}}, // level 4
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.Tags["cost-centre"] != "platform" {
		t.Errorf("tags-union: mandatory.tags[cost-centre] = %q, want platform", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["cdn-class"] != "production" {
		t.Errorf("tags-union: mandatory.tags[cdn-class] = %q, want production", got.Mandatory.Tags["cdn-class"])
	}
}

// TestMergeCloudFrontCascade_TagsKeyConflict — on key conflict in mandatory tags,
// lower level number wins. Level 1 wins over level 4.
func TestMergeCloudFrontCascade_TagsKeyConflict(t *testing.T) {
	got := mergeCFAll(
		cascade.CloudFrontKropathSection{Tags: map[string]string{"env": "org-level"}},      // level 1
		zeroKropathCF,
		zeroCFCfg,
		cascade.CloudFrontConfigSection{Tags: map[string]string{"env": "config-level"}}, // level 4
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.Tags["env"] != "org-level" {
		t.Errorf("tags-conflict: mandatory.tags[env] = %q, want org-level (level 1 wins over level 4)", got.Mandatory.Tags["env"])
	}
}

// TestMergeCloudFrontCascade_DefaultsTagsUnion — defaults tags are union-merged from all four
// defaults sources; level 6 wins on key conflict.
func TestMergeCloudFrontCascade_DefaultsTagsUnion(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		cascade.CloudFrontConfigSection{Tags: map[string]string{"env": "local-cfg", "cache": "enabled"}}, // level 6
		cascade.CloudFrontConfigSection{Tags: map[string]string{"env": "global-cfg"}},                    // level 7
		cascade.CloudFrontKropathSection{Tags: map[string]string{"org": "kropath"}},                      // level 8
		zeroKropathCF,
	)

	if got.Defaults.Tags["env"] != "local-cfg" {
		t.Errorf("defaults-tags-union: defaults.tags[env] = %q, want local-cfg (level 6 wins)", got.Defaults.Tags["env"])
	}
	if got.Defaults.Tags["cache"] != "enabled" {
		t.Errorf("defaults-tags-union: defaults.tags[cache] = %q, want enabled", got.Defaults.Tags["cache"])
	}
	if got.Defaults.Tags["org"] != "kropath" {
		t.Errorf("defaults-tags-union: defaults.tags[org] = %q, want kropath", got.Defaults.Tags["org"])
	}
}

// TestMergeCloudFrontCascade_SyncedLabelsUnion — SyncedLabels from global (L3) and
// local (L4) CFConfig mandatory tiers are union-merged; L3 wins on key conflict.
func TestMergeCloudFrontCascade_SyncedLabelsUnion(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		cascade.CloudFrontConfigSection{SyncedLabels: map[string]string{"tier": "global", "data-class": "public"}}, // level 3
		cascade.CloudFrontConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "prod"}},           // level 4
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "public" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[data-class] = %q, want public", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["env"] != "prod" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[env] = %q, want prod", got.Mandatory.SyncedLabels["env"])
	}
	if got.Mandatory.SyncedLabels["tier"] != "global" {
		t.Errorf("synced-labels-union: mandatory.syncedLabels[tier] = %q, want global (L3 wins over L4)", got.Mandatory.SyncedLabels["tier"])
	}
}

// TestMergeCloudFrontCascade_SyncedLabelsDefaultsUnion — SyncedLabels from global (L7) and
// local (L6) CFConfig defaults tiers are union-merged; L6 wins on key conflict.
func TestMergeCloudFrontCascade_SyncedLabelsDefaultsUnion(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		cascade.CloudFrontConfigSection{SyncedLabels: map[string]string{"tier": "local", "env": "staging"}},  // level 6
		cascade.CloudFrontConfigSection{SyncedLabels: map[string]string{"tier": "global", "region": "apac"}}, // level 7
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Defaults.SyncedLabels["tier"] != "local" {
		t.Errorf("synced-labels-defaults: defaults.syncedLabels[tier] = %q, want local (L6 wins over L7)", got.Defaults.SyncedLabels["tier"])
	}
	if got.Defaults.SyncedLabels["env"] != "staging" {
		t.Errorf("synced-labels-defaults: defaults.syncedLabels[env] = %q, want staging", got.Defaults.SyncedLabels["env"])
	}
	if got.Defaults.SyncedLabels["region"] != "apac" {
		t.Errorf("synced-labels-defaults: defaults.syncedLabels[region] = %q, want apac", got.Defaults.SyncedLabels["region"])
	}
}

// TestMergeCloudFrontCascade_SyncedAnnotations — SyncedAnnotations from CFConfig levels only.
func TestMergeCloudFrontCascade_SyncedAnnotations(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		cascade.CloudFrontConfigSection{SyncedAnnotations: map[string]string{"compliance": "pci-dss"}}, // level 3
		cascade.CloudFrontConfigSection{SyncedAnnotations: map[string]string{"team": "platform"}},     // level 4
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.SyncedAnnotations["compliance"] != "pci-dss" {
		t.Errorf("synced-annotations: mandatory.syncedAnnotations[compliance] = %q, want pci-dss", got.Mandatory.SyncedAnnotations["compliance"])
	}
	if got.Mandatory.SyncedAnnotations["team"] != "platform" {
		t.Errorf("synced-annotations: mandatory.syncedAnnotations[team] = %q, want platform", got.Mandatory.SyncedAnnotations["team"])
	}
}

// TestMergeCloudFrontCascade_CFConfigOnlyFieldsIgnoreKropathLevels — httpVersion set in
// KropathConfig sections (via Tags only) does NOT bleed into mandatory.httpVersion.
func TestMergeCloudFrontCascade_CFConfigOnlyFieldsIgnoreKropathLevels(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.HttpVersion != "" {
		t.Errorf("cf-only-fields: mandatory.httpVersion = %q, want empty (KropathConfig has no httpVersion)", got.Mandatory.HttpVersion)
	}
	if got.Mandatory.PriceClass != "" {
		t.Errorf("cf-only-fields: mandatory.priceClass = %q, want empty", got.Mandatory.PriceClass)
	}
	if got.Mandatory.GeoRestrictionType != "" {
		t.Errorf("cf-only-fields: mandatory.geoRestrictionType = %q, want empty", got.Mandatory.GeoRestrictionType)
	}
}

// TestMergeCloudFrontCascade_VpcOriginProtocolPolicy_Mandatory — vpcOriginProtocolPolicy is
// CloudFrontConfig-only; level 3 wins over level 4 in the mandatory tier.
func TestMergeCloudFrontCascade_VpcOriginProtocolPolicy_Mandatory(t *testing.T) {
	cases := []struct {
		name                 string
		globalCFCfgMandatory cascade.CloudFrontConfigSection
		localCFCfgMandatory  cascade.CloudFrontConfigSection
		wantPolicy           string
	}{
		{
			name:                 "level3-wins",
			globalCFCfgMandatory: cascade.CloudFrontConfigSection{VpcOriginProtocolPolicy: "https-only"},
			localCFCfgMandatory:  cascade.CloudFrontConfigSection{VpcOriginProtocolPolicy: "http-only"},
			wantPolicy:           "https-only",
		},
		{
			name:                 "level4-wins-when-3-absent",
			globalCFCfgMandatory: zeroCFCfg,
			localCFCfgMandatory:  cascade.CloudFrontConfigSection{VpcOriginProtocolPolicy: "match-viewer"},
			wantPolicy:           "match-viewer",
		},
		{
			name:                 "empty-when-3-4-absent",
			globalCFCfgMandatory: zeroCFCfg,
			localCFCfgMandatory:  zeroCFCfg,
			wantPolicy:           "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCFAll(
				zeroKropathCF,
				zeroKropathCF,
				tc.globalCFCfgMandatory,
				tc.localCFCfgMandatory,
				zeroCFCfg,
				zeroCFCfg,
				zeroKropathCF,
				zeroKropathCF,
			)
			if got.Mandatory.VpcOriginProtocolPolicy != tc.wantPolicy {
				t.Errorf("mandatory.vpcOriginProtocolPolicy = %q, want %q", got.Mandatory.VpcOriginProtocolPolicy, tc.wantPolicy)
			}
		})
	}
}

// TestMergeCloudFrontCascade_VpcOriginProtocolPolicy_Defaults — vpcOriginProtocolPolicy is
// CloudFrontConfig-only; level 6 wins over level 7 in the defaults tier.
func TestMergeCloudFrontCascade_VpcOriginProtocolPolicy_Defaults(t *testing.T) {
	cases := []struct {
		name                string
		localCFCfgDefaults  cascade.CloudFrontConfigSection
		globalCFCfgDefaults cascade.CloudFrontConfigSection
		wantPolicy          string
	}{
		{
			name:                "level6-wins",
			localCFCfgDefaults:  cascade.CloudFrontConfigSection{VpcOriginProtocolPolicy: "https-only"},
			globalCFCfgDefaults: cascade.CloudFrontConfigSection{VpcOriginProtocolPolicy: "http-only"},
			wantPolicy:          "https-only",
		},
		{
			name:                "level7-wins-when-6-absent",
			localCFCfgDefaults:  zeroCFCfg,
			globalCFCfgDefaults: cascade.CloudFrontConfigSection{VpcOriginProtocolPolicy: "https-only"},
			wantPolicy:          "https-only",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeCFAll(
				zeroKropathCF,
				zeroKropathCF,
				zeroCFCfg,
				zeroCFCfg,
				tc.localCFCfgDefaults,
				tc.globalCFCfgDefaults,
				zeroKropathCF,
				zeroKropathCF,
			)
			if got.Defaults.VpcOriginProtocolPolicy != tc.wantPolicy {
				t.Errorf("defaults.vpcOriginProtocolPolicy = %q, want %q", got.Defaults.VpcOriginProtocolPolicy, tc.wantPolicy)
			}
		})
	}
}

// TestMergeCloudFrontCascade_VpcOriginProtocolPolicyIgnoresKropathLevels — vpcOriginProtocolPolicy
// is not promoted to KropathConfig (ADR-018 D-3); CloudFrontKropathSection has no such field, so
// this asserts the merge stays empty when only CFConfig-only fields on other struct types are set.
func TestMergeCloudFrontCascade_VpcOriginProtocolPolicyIgnoresKropathLevels(t *testing.T) {
	got := mergeCFAll(
		zeroKropathCF,
		zeroKropathCF,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroCFCfg,
		zeroKropathCF,
		zeroKropathCF,
	)

	if got.Mandatory.VpcOriginProtocolPolicy != "" {
		t.Errorf("vpc-origin-protocol-policy: mandatory.vpcOriginProtocolPolicy = %q, want empty (KropathConfig has no vpcOriginProtocolPolicy)", got.Mandatory.VpcOriginProtocolPolicy)
	}
	if got.Defaults.VpcOriginProtocolPolicy != "" {
		t.Errorf("vpc-origin-protocol-policy: defaults.vpcOriginProtocolPolicy = %q, want empty", got.Defaults.VpcOriginProtocolPolicy)
	}
}
