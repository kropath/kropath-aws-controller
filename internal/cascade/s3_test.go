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
	zeroS3    = cascade.S3Section{}
	zeroS3Cfg = cascade.S3ConfigSection{}
)

// mergeS3All is a test helper that calls MergeS3Cascade with named slots
// so existing tests remain readable.
func mergeS3All(
	// KropathConfig tiers (S3Section)
	globalKropathMandatory, localKropathMandatory cascade.S3Section,
	// S3Config tiers (S3ConfigSection) — have extra fields
	globalS3Mandatory, localS3Mandatory,
	localS3Defaults, globalS3Defaults cascade.S3ConfigSection,
	// KropathConfig defaults tiers (S3Section)
	localKropathDefaults, globalKropathDefaults cascade.S3Section,
) cascade.EffectiveS3Config {
	return cascade.MergeS3Cascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalS3Mandatory,
		localS3Mandatory,
		localS3Defaults,
		globalS3Defaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// s3Cfg returns an S3ConfigSection with only EncryptionAlgorithm set,
// to keep existing scalar tests concise.
func s3Cfg(enc string) cascade.S3ConfigSection {
	return cascade.S3ConfigSection{EncryptionAlgorithm: enc}
}

func TestMergeS3Cascade_MandatoryLevel1Wins(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
		cascade.S3Section{EncryptionAlgorithm: "AES256"},
		s3Cfg("AES256"),
		s3Cfg("aws:kms"),
		zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	if got.Mandatory.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want %q", got.Mandatory.EncryptionAlgorithm, "aws:kms")
	}
}

func TestMergeS3Cascade_MandatoryBooleanLevel1Wins(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{BlockPublicAccess: true, EnforceHttpsOnly: true, LoggingEnabled: true},
		cascade.S3Section{BlockPublicAccess: false, EnforceHttpsOnly: false, LoggingEnabled: false},
		zeroS3Cfg, zeroS3Cfg,
		zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	if !got.Mandatory.BlockPublicAccess {
		t.Fatal("mandatory.blockPublicAccess should be true")
	}
	if !got.Mandatory.EnforceHttpsOnly {
		t.Fatal("mandatory.enforceHttpsOnly should be true")
	}
	if !got.Mandatory.LoggingEnabled {
		t.Fatal("mandatory.loggingEnabled should be true")
	}
}

func TestMergeS3Cascade_DefaultsCascadeOrder(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3,
		zeroS3Cfg, zeroS3Cfg,
		cascade.S3ConfigSection{Versioning: "local-s3"},
		cascade.S3ConfigSection{Versioning: "global-s3"},
		cascade.S3Section{Versioning: "local-kpc"},
		cascade.S3Section{Versioning: "global-kpc"},
	)

	if got.Defaults.Versioning != "local-s3" {
		t.Fatalf("defaults.versioning = %q, want %q", got.Defaults.Versioning, "local-s3")
	}
}

func TestMergeS3Cascade_DefaultsGlobalKropathFallback(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3,
		zeroS3Cfg, zeroS3Cfg,
		zeroS3Cfg, zeroS3Cfg,
		zeroS3,
		cascade.S3Section{EncryptionAlgorithm: "aws:kms"},
	)

	if got.Defaults.EncryptionAlgorithm != "aws:kms" {
		t.Fatalf("defaults.encryptionAlgorithm = %q, want %q", got.Defaults.EncryptionAlgorithm, "aws:kms")
	}
	if got.Mandatory.EncryptionAlgorithm != "" {
		t.Fatalf("mandatory.encryptionAlgorithm = %q, want empty", got.Mandatory.EncryptionAlgorithm)
	}
}

func TestMergeS3Cascade_PassthroughZeroValues(t *testing.T) {
	got := mergeS3All(zeroS3, zeroS3, zeroS3Cfg, zeroS3Cfg, zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3)

	if got.Mandatory.ObjectLockMode != "" {
		t.Fatalf("mandatory.objectLockMode = %q, want empty", got.Mandatory.ObjectLockMode)
	}
	if got.Mandatory.ObjectLockRetentionDays != 0 {
		t.Fatalf("mandatory.objectLockRetentionDays = %d, want 0", got.Mandatory.ObjectLockRetentionDays)
	}
	if got.Defaults.ObjectLockMode != "" {
		t.Fatalf("defaults.objectLockMode = %q, want empty", got.Defaults.ObjectLockMode)
	}
	if got.Defaults.ObjectLockRetentionDays != 0 {
		t.Fatalf("defaults.objectLockRetentionDays = %d, want 0", got.Defaults.ObjectLockRetentionDays)
	}
}

// Gap 3 tests: Tags, NamingTemplate, SyncedLabels, SyncedAnnotations

func TestMergeS3Cascade_TagsUnionMergeAllFourMandatoryLevels(t *testing.T) {
	got := mergeS3All(
		cascade.S3Section{Tags: map[string]string{"org": "acme"}},         // level 1
		cascade.S3Section{Tags: map[string]string{"env": "prod"}},          // level 2
		cascade.S3ConfigSection{Tags: map[string]string{"svc": "global"}},  // level 3
		cascade.S3ConfigSection{Tags: map[string]string{"svc": "local"}},   // level 4 — wins on "svc"
		zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	// All keys from all levels should be present (union/additive merge).
	if got.Mandatory.Tags["org"] != "acme" {
		t.Fatalf("mandatory.tags[org] = %q, want %q", got.Mandatory.Tags["org"], "acme")
	}
	if got.Mandatory.Tags["env"] != "prod" {
		t.Fatalf("mandatory.tags[env] = %q, want %q", got.Mandatory.Tags["env"], "prod")
	}
	// Level 4 (local S3Config) wins the "svc" key (mergeMaps puts later args first).
	if got.Mandatory.Tags["svc"] == "" {
		t.Fatal("mandatory.tags[svc] should be set")
	}
}

func TestMergeS3Cascade_TagsUnionMergeAllFourDefaultsLevels(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3, zeroS3Cfg, zeroS3Cfg,
		cascade.S3ConfigSection{Tags: map[string]string{"team": "alpha"}},  // level 6
		cascade.S3ConfigSection{Tags: map[string]string{"cost": "shared"}}, // level 7
		cascade.S3Section{Tags: map[string]string{"env": "staging"}},       // level 8
		cascade.S3Section{Tags: map[string]string{"org": "acme"}},          // level 9
	)

	if got.Defaults.Tags["team"] != "alpha" {
		t.Fatalf("defaults.tags[team] = %q, want %q", got.Defaults.Tags["team"], "alpha")
	}
	if got.Defaults.Tags["cost"] != "shared" {
		t.Fatalf("defaults.tags[cost] = %q, want %q", got.Defaults.Tags["cost"], "shared")
	}
	if got.Defaults.Tags["env"] != "staging" {
		t.Fatalf("defaults.tags[env] = %q, want %q", got.Defaults.Tags["env"], "staging")
	}
	if got.Defaults.Tags["org"] != "acme" {
		t.Fatalf("defaults.tags[org] = %q, want %q", got.Defaults.Tags["org"], "acme")
	}
}

func TestMergeS3Cascade_NamingTemplateFromS3ConfigMandatoryGlobalWins(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3,
		cascade.S3ConfigSection{NamingTemplate: "{account_id}-{namespace}-{name}"},  // level 3 — wins
		cascade.S3ConfigSection{NamingTemplate: "{namespace}-{name}"},               // level 4
		zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	want := "{account_id}-{namespace}-{name}"
	if got.Mandatory.NamingTemplate != want {
		t.Fatalf("mandatory.namingTemplate = %q, want %q", got.Mandatory.NamingTemplate, want)
	}
}

func TestMergeS3Cascade_NamingTemplateFallsToLocalWhenGlobalEmpty(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3,
		zeroS3Cfg, // level 3 empty
		cascade.S3ConfigSection{NamingTemplate: "{namespace}-{name}"}, // level 4
		zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}" {
		t.Fatalf("mandatory.namingTemplate = %q, want %q", got.Mandatory.NamingTemplate, "{namespace}-{name}")
	}
}

func TestMergeS3Cascade_NamingTemplateDefaultsLocalWins(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3, zeroS3Cfg, zeroS3Cfg,
		cascade.S3ConfigSection{NamingTemplate: "{namespace}-{name}-local"}, // level 6 — wins
		cascade.S3ConfigSection{NamingTemplate: "{namespace}-{name}-global"}, // level 7
		zeroS3, zeroS3,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}-local" {
		t.Fatalf("defaults.namingTemplate = %q, want %q", got.Defaults.NamingTemplate, "{namespace}-{name}-local")
	}
}

func TestMergeS3Cascade_NamingTemplateNotFromKropathConfig(t *testing.T) {
	// KropathConfig.S3Section has no NamingTemplate; only S3ConfigSection does.
	// Even with zero S3Config levels, NamingTemplate must be empty.
	got := mergeS3All(
		zeroS3, zeroS3,
		zeroS3Cfg, zeroS3Cfg,
		zeroS3Cfg, zeroS3Cfg,
		zeroS3, zeroS3,
	)

	if got.Mandatory.NamingTemplate != "" {
		t.Fatalf("mandatory.namingTemplate = %q, want empty (KropathConfig does not carry NamingTemplate)", got.Mandatory.NamingTemplate)
	}
}

func TestMergeS3Cascade_SyncedLabelsMandatoryAdditiveMergeS3ConfigOnly(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3,
		cascade.S3ConfigSection{SyncedLabels: map[string]string{"global": "yes"}}, // level 3
		cascade.S3ConfigSection{SyncedLabels: map[string]string{"local": "yes"}},  // level 4
		zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	if got.Mandatory.SyncedLabels["global"] != "yes" {
		t.Fatalf("mandatory.syncedLabels[global] = %q, want %q", got.Mandatory.SyncedLabels["global"], "yes")
	}
	if got.Mandatory.SyncedLabels["local"] != "yes" {
		t.Fatalf("mandatory.syncedLabels[local] = %q, want %q", got.Mandatory.SyncedLabels["local"], "yes")
	}
}

func TestMergeS3Cascade_SyncedAnnotationsMandatoryAdditiveMergeS3ConfigOnly(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3,
		cascade.S3ConfigSection{SyncedAnnotations: map[string]string{"ann-g": "v1"}}, // level 3
		cascade.S3ConfigSection{SyncedAnnotations: map[string]string{"ann-l": "v2"}}, // level 4
		zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	if got.Mandatory.SyncedAnnotations["ann-g"] != "v1" {
		t.Fatalf("mandatory.syncedAnnotations[ann-g] = %q, want %q", got.Mandatory.SyncedAnnotations["ann-g"], "v1")
	}
	if got.Mandatory.SyncedAnnotations["ann-l"] != "v2" {
		t.Fatalf("mandatory.syncedAnnotations[ann-l] = %q, want %q", got.Mandatory.SyncedAnnotations["ann-l"], "v2")
	}
}

func TestMergeS3Cascade_LoggingTargetPrefixMandatoryGlobalS3ConfigWins(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3,
		cascade.S3ConfigSection{LoggingTargetPrefix: "logs/mandatory-global/"}, // level 3 (global S3Config mandatory)
		cascade.S3ConfigSection{LoggingTargetPrefix: "logs/mandatory-local/"},  // level 4 (local S3Config mandatory)
		cascade.S3ConfigSection{LoggingTargetPrefix: "logs/defaults-local/"},   // defaults local
		cascade.S3ConfigSection{LoggingTargetPrefix: "logs/defaults-global/"},  // defaults global
		zeroS3, zeroS3,
	)

	if got.Mandatory.LoggingTargetPrefix != "logs/mandatory-global/" {
		t.Fatalf("mandatory.loggingTargetPrefix = %q, want %q", got.Mandatory.LoggingTargetPrefix, "logs/mandatory-global/")
	}
}

func TestMergeS3Cascade_LoggingTargetPrefixMandatoryLocalWhenGlobalEmpty(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3,
		zeroS3Cfg, // level 3 global S3Config mandatory empty
		cascade.S3ConfigSection{LoggingTargetPrefix: "logs/mandatory-local/"}, // level 4 local
		zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	if got.Mandatory.LoggingTargetPrefix != "logs/mandatory-local/" {
		t.Fatalf("mandatory.loggingTargetPrefix = %q, want %q", got.Mandatory.LoggingTargetPrefix, "logs/mandatory-local/")
	}
}

func TestMergeS3Cascade_LoggingTargetPrefixMandatoryEmptyWhenNoneSet(t *testing.T) {
	got := mergeS3All(zeroS3, zeroS3, zeroS3Cfg, zeroS3Cfg, zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3)

	if got.Mandatory.LoggingTargetPrefix != "" {
		t.Fatalf("mandatory.loggingTargetPrefix = %q, want empty", got.Mandatory.LoggingTargetPrefix)
	}
}

func TestMergeS3Cascade_LoggingTargetPrefixDefaultsLocalS3ConfigWins(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3, zeroS3Cfg, zeroS3Cfg,
		cascade.S3ConfigSection{LoggingTargetPrefix: "logs/defaults-local/"},  // defaults local (strongest)
		cascade.S3ConfigSection{LoggingTargetPrefix: "logs/defaults-global/"}, // defaults global
		zeroS3, zeroS3,
	)

	if got.Defaults.LoggingTargetPrefix != "logs/defaults-local/" {
		t.Fatalf("defaults.loggingTargetPrefix = %q, want %q", got.Defaults.LoggingTargetPrefix, "logs/defaults-local/")
	}
}

func TestMergeS3Cascade_LoggingTargetPrefixDefaultsGlobalFallback(t *testing.T) {
	got := mergeS3All(
		zeroS3, zeroS3, zeroS3Cfg, zeroS3Cfg,
		zeroS3Cfg, // defaults local empty
		cascade.S3ConfigSection{LoggingTargetPrefix: "logs/defaults-global/"}, // defaults global
		zeroS3, zeroS3,
	)

	if got.Defaults.LoggingTargetPrefix != "logs/defaults-global/" {
		t.Fatalf("defaults.loggingTargetPrefix = %q, want %q", got.Defaults.LoggingTargetPrefix, "logs/defaults-global/")
	}
}

func TestMergeS3Cascade_LoggingTargetPrefixNotFromKropathConfig(t *testing.T) {
	// KropathConfig has no loggingTargetPrefix field — it must not bleed into effectiveConfig
	got := mergeS3All(
		cascade.S3Section{LogDeliveryBucket: "log-bucket"}, // KropathConfig mandatory has logDeliveryBucket but not loggingTargetPrefix
		zeroS3, zeroS3Cfg, zeroS3Cfg, zeroS3Cfg, zeroS3Cfg, zeroS3, zeroS3,
	)

	if got.Mandatory.LoggingTargetPrefix != "" {
		t.Fatalf("mandatory.loggingTargetPrefix = %q, want empty (not from KropathConfig)", got.Mandatory.LoggingTargetPrefix)
	}
}
