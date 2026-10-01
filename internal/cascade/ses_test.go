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

var (
	zeroSESKropath = cascade.SESKropathSection{}
	zeroSESCfg     = cascade.SESConfigSection{}
)

// mergeSESAll calls MergeSESCascade with all eight inputs.
func mergeSESAll(
	globalKropathMandatory,
	localKropathMandatory cascade.SESKropathSection,
	globalSESCfgMandatory,
	localSESCfgMandatory,
	localSESCfgDefaults,
	globalSESCfgDefaults cascade.SESConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.SESKropathSection,
) cascade.EffectiveSESConfig {
	return cascade.MergeSESCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalSESCfgMandatory,
		localSESCfgMandatory,
		localSESCfgDefaults,
		globalSESCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeSESCascade_MandatoryTagsFromKropathConfig — AC-3 controller unit test.
// globalKropathConfig.mandatory.tags win over SESConfig mandatory tags (level 1 beats level 4).
func TestMergeSESCascade_MandatoryTagsFromKropathConfig(t *testing.T) {
	got := mergeSESAll(
		cascade.SESKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // level 1
		zeroSESKropath, // level 2
		zeroSESCfg,     // level 3
		cascade.SESConfigSection{Tags: map[string]string{"cost-centre": "team"}}, // level 4 — overridden by L1
		zeroSESCfg,     // level 6
		zeroSESCfg,     // level 7
		zeroSESKropath, // level 8
		zeroSESKropath, // level 9
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-3: mandatory.tags[cost-centre] = %q, want %q (level 1 wins)", got.Mandatory.Tags["cost-centre"], "infra")
	}
	if len(got.Defaults.Tags) != 0 {
		t.Errorf("AC-3: mandatory tags must not bleed into defaults; got defaults.tags = %v", got.Defaults.Tags)
	}
}

// TestMergeSESCascade_DefaultsNamingTemplate — AC-4 controller unit test.
// localSESCfgDefaults.NamingTemplate at level 6 propagates to effCfg.defaults.namingTemplate.
func TestMergeSESCascade_DefaultsNamingTemplate(t *testing.T) {
	tmpl := "{namespace}-{name}"
	got := mergeSESAll(
		zeroSESKropath, // level 1
		zeroSESKropath, // level 2
		zeroSESCfg,     // level 3
		zeroSESCfg,     // level 4
		cascade.SESConfigSection{NamingTemplate: tmpl}, // level 6
		zeroSESCfg,     // level 7
		zeroSESKropath, // level 8
		zeroSESKropath, // level 9
	)

	if got.Defaults.NamingTemplate != tmpl {
		t.Errorf("AC-4: defaults.namingTemplate = %q, want %q (level 6)", got.Defaults.NamingTemplate, tmpl)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-4: defaults namingTemplate must not bleed into mandatory; got mandatory.namingTemplate = %q", got.Mandatory.NamingTemplate)
	}
}

// TestMergeSESCascade_MandatoryTagsMerge — AC-5 controller unit test.
// Tags from SESConfig mandatory and KropathConfig mandatory are additive union;
// when a key conflict exists, the higher-priority (lower level number) wins.
func TestMergeSESCascade_MandatoryTagsMerge(t *testing.T) {
	got := mergeSESAll(
		cascade.SESKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // level 1
		zeroSESKropath, // level 2
		zeroSESCfg,     // level 3
		cascade.SESConfigSection{Tags: map[string]string{"email-type": "transactional", "cost-centre": "team"}}, // level 4
		cascade.SESConfigSection{Tags: map[string]string{"team": "platform"}}, // level 6
		zeroSESCfg,     // level 7
		zeroSESKropath, // level 8
		zeroSESKropath, // level 9
	)

	// Both keys from KropathConfig (L1) and SESConfig mandatory (L4) must be present.
	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-5: mandatory.tags[cost-centre] = %q, want %q (L1 wins on conflict)", got.Mandatory.Tags["cost-centre"], "infra")
	}
	if got.Mandatory.Tags["email-type"] != "transactional" {
		t.Errorf("AC-5: mandatory.tags[email-type] = %q, want %q (L4 unique key present)", got.Mandatory.Tags["email-type"], "transactional")
	}
	// Defaults source (L6) must not bleed into mandatory.
	if got.Mandatory.Tags["team"] != "" {
		t.Errorf("AC-5: defaults tag must not bleed into mandatory; mandatory.tags[team] = %q", got.Mandatory.Tags["team"])
	}
	// Defaults must include L6 tag.
	if got.Defaults.Tags["team"] != "platform" {
		t.Errorf("AC-5: defaults.tags[team] = %q, want %q", got.Defaults.Tags["team"], "platform")
	}
}

// TestMergeSESCascade_MandatorySyncedLabels — AC-6 controller unit test.
// syncedLabels from SESConfig mandatory and KropathConfig mandatory are additive union.
func TestMergeSESCascade_MandatorySyncedLabels(t *testing.T) {
	got := mergeSESAll(
		cascade.SESKropathSection{SyncedLabels: map[string]string{"org": "kropath"}}, // level 1
		zeroSESKropath, // level 2
		zeroSESCfg,     // level 3
		cascade.SESConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "org": "ses-team"}}, // level 4
		zeroSESCfg,     // level 6
		zeroSESCfg,     // level 7
		zeroSESKropath, // level 8
		zeroSESKropath, // level 9
	)

	// L1 wins on "org" key conflict.
	if got.Mandatory.SyncedLabels["org"] != "kropath" {
		t.Errorf("AC-6: mandatory.syncedLabels[org] = %q, want %q (L1 wins)", got.Mandatory.SyncedLabels["org"], "kropath")
	}
	// L4 unique key must be present.
	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-6: mandatory.syncedLabels[data-class] = %q, want %q", got.Mandatory.SyncedLabels["data-class"], "internal")
	}
	if len(got.Defaults.SyncedLabels) != 0 {
		t.Errorf("AC-6: mandatory syncedLabels must not bleed into defaults; got defaults.syncedLabels = %v", got.Defaults.SyncedLabels)
	}
}

// TestMergeSESCascade_MandatorySyncedAnnotations — AC-7 controller unit test.
// syncedAnnotations from SESConfig mandatory and KropathConfig mandatory are additive union.
func TestMergeSESCascade_MandatorySyncedAnnotations(t *testing.T) {
	got := mergeSESAll(
		cascade.SESKropathSection{SyncedAnnotations: map[string]string{"compliance": "hipaa"}}, // level 1
		zeroSESKropath, // level 2
		zeroSESCfg,     // level 3
		cascade.SESConfigSection{SyncedAnnotations: map[string]string{"owner": "email-team", "compliance": "ses"}}, // level 4
		zeroSESCfg,     // level 6
		zeroSESCfg,     // level 7
		zeroSESKropath, // level 8
		zeroSESKropath, // level 9
	)

	// L1 wins on "compliance" key conflict.
	if got.Mandatory.SyncedAnnotations["compliance"] != "hipaa" {
		t.Errorf("AC-7: mandatory.syncedAnnotations[compliance] = %q, want %q (L1 wins)", got.Mandatory.SyncedAnnotations["compliance"], "hipaa")
	}
	// L4 unique key must be present.
	if got.Mandatory.SyncedAnnotations["owner"] != "email-team" {
		t.Errorf("AC-7: mandatory.syncedAnnotations[owner] = %q, want %q", got.Mandatory.SyncedAnnotations["owner"], "email-team")
	}
}

// TestMergeSESCascade_AllMapFields — AC-8 controller unit test.
// All three map fields (tags, syncedLabels, syncedAnnotations) present in both tiers.
func TestMergeSESCascade_AllMapFields(t *testing.T) {
	got := mergeSESAll(
		zeroSESKropath, // level 1
		zeroSESKropath, // level 2
		zeroSESCfg,     // level 3
		cascade.SESConfigSection{ // level 4
			Tags:              map[string]string{"env": "prod"},
			SyncedLabels:      map[string]string{"tier": "email"},
			SyncedAnnotations: map[string]string{"provisioner": "kropath"},
		},
		cascade.SESConfigSection{ // level 6
			Tags:              map[string]string{"team": "platform"},
			SyncedLabels:      map[string]string{"tier": "shared"},
			SyncedAnnotations: map[string]string{"provisioner": "self-service"},
		},
		zeroSESCfg,     // level 7
		zeroSESKropath, // level 8
		zeroSESKropath, // level 9
	)

	// Mandatory maps from L4 must be present.
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("AC-8: mandatory.tags[env] = %q, want %q", got.Mandatory.Tags["env"], "prod")
	}
	if got.Mandatory.SyncedLabels["tier"] != "email" {
		t.Errorf("AC-8: mandatory.syncedLabels[tier] = %q, want %q", got.Mandatory.SyncedLabels["tier"], "email")
	}
	if got.Mandatory.SyncedAnnotations["provisioner"] != "kropath" {
		t.Errorf("AC-8: mandatory.syncedAnnotations[provisioner] = %q, want %q", got.Mandatory.SyncedAnnotations["provisioner"], "kropath")
	}
	// Defaults maps from L6 must be present.
	if got.Defaults.Tags["team"] != "platform" {
		t.Errorf("AC-8: defaults.tags[team] = %q, want %q", got.Defaults.Tags["team"], "platform")
	}
	// L6 wins over L7 on key conflict for defaults.
	if got.Defaults.SyncedLabels["tier"] != "shared" {
		t.Errorf("AC-8: defaults.syncedLabels[tier] = %q, want %q (L6 wins)", got.Defaults.SyncedLabels["tier"], "shared")
	}
	if got.Defaults.SyncedAnnotations["provisioner"] != "self-service" {
		t.Errorf("AC-8: defaults.syncedAnnotations[provisioner] = %q, want %q (L6 wins)", got.Defaults.SyncedAnnotations["provisioner"], "self-service")
	}
}

// TestMergeSESCascade_MandatoryNamingTemplateSESConfigOnly verifies that namingTemplate
// in the mandatory tier comes only from SESConfig levels (L3, L4), not KropathConfig.
func TestMergeSESCascade_MandatoryNamingTemplateSESConfigOnly(t *testing.T) {
	got := mergeSESAll(
		zeroSESKropath, // level 1
		zeroSESKropath, // level 2
		cascade.SESConfigSection{NamingTemplate: "corp-{name}"}, // level 3
		zeroSESCfg,     // level 4
		zeroSESCfg,     // level 6
		zeroSESCfg,     // level 7
		zeroSESKropath, // level 8
		zeroSESKropath, // level 9
	)

	if got.Mandatory.NamingTemplate != "corp-{name}" {
		t.Errorf("mandatory.namingTemplate = %q, want %q (L3 provides mandatory namingTemplate)", got.Mandatory.NamingTemplate, "corp-{name}")
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("mandatory namingTemplate must not bleed into defaults; got defaults.namingTemplate = %q", got.Defaults.NamingTemplate)
	}
}

// TestMergeSESCascade_DefaultsPriorityOrder verifies that L6 wins over L7, L8, L9 for defaults.
func TestMergeSESCascade_DefaultsPriorityOrder(t *testing.T) {
	got := mergeSESAll(
		zeroSESKropath, // level 1
		zeroSESKropath, // level 2
		zeroSESCfg,     // level 3
		zeroSESCfg,     // level 4
		cascade.SESConfigSection{Tags: map[string]string{"owner": "local-ses"}}, // level 6 (wins)
		cascade.SESConfigSection{Tags: map[string]string{"owner": "global-ses"}}, // level 7
		cascade.SESKropathSection{Tags: map[string]string{"owner": "local-kpc"}}, // level 8
		cascade.SESKropathSection{Tags: map[string]string{"owner": "global-kpc"}}, // level 9
	)

	if got.Defaults.Tags["owner"] != "local-ses" {
		t.Errorf("defaults priority: defaults.tags[owner] = %q, want %q (L6 wins)", got.Defaults.Tags["owner"], "local-ses")
	}
}

// TestMergeSESCascade_EmptySourcesYieldEmptyResult verifies that all-zero inputs
// produce an empty EffectiveSESConfig with no panics.
func TestMergeSESCascade_EmptySourcesYieldEmptyResult(t *testing.T) {
	got := mergeSESAll(
		zeroSESKropath, zeroSESKropath,
		zeroSESCfg, zeroSESCfg, zeroSESCfg, zeroSESCfg,
		zeroSESKropath, zeroSESKropath,
	)

	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("empty inputs: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("empty inputs: mandatory.tags = %v, want nil/empty", got.Mandatory.Tags)
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("empty inputs: defaults.namingTemplate = %q, want empty", got.Defaults.NamingTemplate)
	}
}
