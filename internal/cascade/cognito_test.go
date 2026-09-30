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

// emptyCognitoKropathSection returns a zero-value CognitoKropathSection (no settings at this level).
func emptyCognitoKropathSection() cascade.CognitoKropathSection {
	return cascade.CognitoKropathSection{}
}

// emptyCognitoConfigSection returns a zero-value CognitoConfigSection.
func emptyCognitoConfigSection() cascade.CognitoConfigSection {
	return cascade.CognitoConfigSection{}
}

// TestMergeCognitoCascade_AC7_GlobalKropathMandatoryMfa verifies that
// KropathConfig.mandatory.cognito.mfaConfiguration="ON" in the global kro-system
// namespace flows through to effectiveConfig.mandatory.mfaConfiguration.
//
// Spec AC-7: global KropathConfig.mandatory.cognito.mfaConfiguration="ON" →
// effectiveConfig.mandatory.mfaConfiguration="ON".
func TestMergeCognitoCascade_AC7_GlobalKropathMandatoryMfa(t *testing.T) {
	globalKropathMandatory := cascade.CognitoKropathSection{
		MfaConfiguration: "ON",
	}

	got := cascade.MergeCognitoCascade(
		globalKropathMandatory,
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	if got.Mandatory.MfaConfiguration != "ON" {
		t.Errorf("AC-7: Mandatory.MfaConfiguration = %q, want %q", got.Mandatory.MfaConfiguration, "ON")
	}
}

// TestMergeCognitoCascade_AC7_LocalKropathMandatoryMfaOverridesGlobal verifies that
// a local namespace KropathConfig.mandatory.cognito.mfaConfiguration="OPTIONAL" cannot
// override the global KropathConfig setting of "ON" (level 1 wins over level 2).
func TestMergeCognitoCascade_AC7_GlobalWinsOverLocalKropathMandatory(t *testing.T) {
	globalKropathMandatory := cascade.CognitoKropathSection{
		MfaConfiguration: "ON",
	}
	localKropathMandatory := cascade.CognitoKropathSection{
		MfaConfiguration: "OPTIONAL",
	}

	got := cascade.MergeCognitoCascade(
		globalKropathMandatory,
		localKropathMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	if got.Mandatory.MfaConfiguration != "ON" {
		t.Errorf("AC-7 (level 1 > 2): Mandatory.MfaConfiguration = %q, want %q", got.Mandatory.MfaConfiguration, "ON")
	}
}

// TestMergeCognitoCascade_AC8_KropathPasswordPolicyMinLengthWins verifies that
// KropathConfig.mandatory.cognito.passwordPolicy.minimumLength=14 wins over
// CognitoConfig.mandatory.passwordPolicy.minimumLength=12 (level 1 beats level 4).
//
// Spec AC-8: global KropathConfig.mandatory.cognito.passwordPolicy.minimumLength=14
// wins over CognitoConfig.mandatory.passwordPolicy.minimumLength=12.
func TestMergeCognitoCascade_AC8_KropathPasswordPolicyMinLengthWins(t *testing.T) {
	globalKropathMandatory := cascade.CognitoKropathSection{
		PasswordPolicy: cascade.CognitoPasswordPolicySection{
			MinimumLength: 14,
		},
	}
	localCognitoCfgMandatory := cascade.CognitoConfigSection{
		PasswordPolicy: cascade.CognitoPasswordPolicySection{
			MinimumLength: 12,
		},
	}

	got := cascade.MergeCognitoCascade(
		globalKropathMandatory,
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	if got.Mandatory.PasswordPolicy.MinimumLength != 14 {
		t.Errorf("AC-8: Mandatory.PasswordPolicy.MinimumLength = %d, want 14", got.Mandatory.PasswordPolicy.MinimumLength)
	}
}

// TestMergeCognitoCascade_AC8_PasswordPolicyBoolPerSubfieldCascade verifies that
// *bool sub-fields in passwordPolicy cascade independently per-sub-field:
// level 1 sets RequireLowercase, level 4 sets RequireNumbers. Both should appear.
func TestMergeCognitoCascade_AC8_PasswordPolicyBoolPerSubfieldCascade(t *testing.T) {
	trueVal := true
	falseVal := false

	globalKropathMandatory := cascade.CognitoKropathSection{
		PasswordPolicy: cascade.CognitoPasswordPolicySection{
			RequireLowercase: &trueVal,
		},
	}
	localCognitoCfgMandatory := cascade.CognitoConfigSection{
		PasswordPolicy: cascade.CognitoPasswordPolicySection{
			RequireNumbers: &falseVal,
		},
	}

	got := cascade.MergeCognitoCascade(
		globalKropathMandatory,
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	if got.Mandatory.PasswordPolicy.RequireLowercase == nil || *got.Mandatory.PasswordPolicy.RequireLowercase != true {
		t.Errorf("AC-8 (bool cascade): RequireLowercase = %v, want true", got.Mandatory.PasswordPolicy.RequireLowercase)
	}
	if got.Mandatory.PasswordPolicy.RequireNumbers == nil || *got.Mandatory.PasswordPolicy.RequireNumbers != false {
		t.Errorf("AC-8 (bool cascade): RequireNumbers = %v, want false (explicitly set)", got.Mandatory.PasswordPolicy.RequireNumbers)
	}
}

// TestMergeCognitoCascade_AC8_PasswordPolicyKropathBoolOverridesCognitoConfig verifies that
// a *bool set at level 1 (globalKropathMandatory) wins over level 4 (localCognitoCfgMandatory).
func TestMergeCognitoCascade_AC8_KropathBoolWinsOverCognitoConfig(t *testing.T) {
	trueVal := true
	falseVal := false

	globalKropathMandatory := cascade.CognitoKropathSection{
		PasswordPolicy: cascade.CognitoPasswordPolicySection{
			RequireUppercase: &trueVal,
		},
	}
	localCognitoCfgMandatory := cascade.CognitoConfigSection{
		PasswordPolicy: cascade.CognitoPasswordPolicySection{
			RequireUppercase: &falseVal,
		},
	}

	got := cascade.MergeCognitoCascade(
		globalKropathMandatory,
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	if got.Mandatory.PasswordPolicy.RequireUppercase == nil || *got.Mandatory.PasswordPolicy.RequireUppercase != true {
		t.Errorf("AC-8: level 1 RequireUppercase=true should win over level 4 false; got %v", got.Mandatory.PasswordPolicy.RequireUppercase)
	}
}

// TestMergeCognitoCascade_AC9_DefaultsNamingTemplateFromCognitoConfig verifies that
// CognitoConfig.defaults.namingTemplate cascades to effectiveConfig.defaults.namingTemplate.
//
// Spec AC-9: CognitoConfig.defaults.namingTemplate="{namespace}-{name}" →
// effectiveConfig.defaults.namingTemplate="{namespace}-{name}".
func TestMergeCognitoCascade_AC9_DefaultsNamingTemplateFromCognitoConfig(t *testing.T) {
	localCognitoCfgDefaults := cascade.CognitoConfigSection{
		NamingTemplate: "{namespace}-{name}",
	}

	got := cascade.MergeCognitoCascade(
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgDefaults,
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("AC-9: Defaults.NamingTemplate = %q, want %q", got.Defaults.NamingTemplate, "{namespace}-{name}")
	}
}

// TestMergeCognitoCascade_AC9_NamingTemplateNotPropagatedFromKropathConfig verifies that
// KropathConfig does NOT carry namingTemplate — if both KropathConfig-level tags exist and
// CognitoConfig sets namingTemplate, only CognitoConfig's setting appears.
func TestMergeCognitoCascade_AC9_NamingTemplateOnlyFromCognitoConfig(t *testing.T) {
	// KropathConfig settings exist (mfaConfiguration) but no namingTemplate.
	globalKropathMandatory := cascade.CognitoKropathSection{
		MfaConfiguration: "ON",
	}
	globalCognitoCfgMandatory := cascade.CognitoConfigSection{
		NamingTemplate: "global-{name}",
	}
	localCognitoCfgMandatory := cascade.CognitoConfigSection{
		NamingTemplate: "local-{name}",
	}

	got := cascade.MergeCognitoCascade(
		globalKropathMandatory,
		emptyCognitoKropathSection(),
		globalCognitoCfgMandatory,
		localCognitoCfgMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	// Level 3 (globalCognitoCfgMandatory) wins over level 4 (local).
	if got.Mandatory.NamingTemplate != "global-{name}" {
		t.Errorf("AC-9: Mandatory.NamingTemplate = %q, want %q", got.Mandatory.NamingTemplate, "global-{name}")
	}
}

// TestMergeCognitoCascade_AC10_TagUnionMerge verifies that tags are additively merged
// across all mandatory sources, with KropathConfig tags winning on conflict.
//
// Spec AC-10: KropathConfig.mandatory.tags and CognitoConfig.mandatory.tags merge;
// the KropathConfig tag value wins when the same key appears in both.
func TestMergeCognitoCascade_AC10_TagUnionMerge(t *testing.T) {
	globalKropathMandatory := cascade.CognitoKropathSection{
		Tags: map[string]string{
			"env":  "production",
			"team": "platform",
		},
	}
	localCognitoCfgMandatory := cascade.CognitoConfigSection{
		Tags: map[string]string{
			"env":     "staging",
			"service": "cognito",
		},
	}

	got := cascade.MergeCognitoCascade(
		globalKropathMandatory,
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	// "env" key: level 1 (globalKropathMandatory) wins over level 4 (localCognitoCfgMandatory).
	if got.Mandatory.Tags["env"] != "production" {
		t.Errorf("AC-10: Tags[env] = %q, want %q (level 1 wins)", got.Mandatory.Tags["env"], "production")
	}
	// "team" key: only in level 1.
	if got.Mandatory.Tags["team"] != "platform" {
		t.Errorf("AC-10: Tags[team] = %q, want %q", got.Mandatory.Tags["team"], "platform")
	}
	// "service" key: only in level 4.
	if got.Mandatory.Tags["service"] != "cognito" {
		t.Errorf("AC-10: Tags[service] = %q, want %q", got.Mandatory.Tags["service"], "cognito")
	}
}

// TestMergeCognitoCascade_AC10_DefaultsTagUnionMerge verifies tag union in the defaults tier.
func TestMergeCognitoCascade_AC10_DefaultsTagUnionMerge(t *testing.T) {
	localCognitoCfgDefaults := cascade.CognitoConfigSection{
		Tags: map[string]string{
			"owner": "team-a",
			"cost":  "high",
		},
	}
	globalKropathDefaults := cascade.CognitoKropathSection{
		Tags: map[string]string{
			"cost":   "low",
			"region": "us-east-1",
		},
	}

	got := cascade.MergeCognitoCascade(
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgDefaults,
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		globalKropathDefaults,
	)

	// level 6 (localCognitoCfgDefaults) wins over level 9 (globalKropathDefaults) in defaults tier.
	if got.Defaults.Tags["cost"] != "high" {
		t.Errorf("AC-10 defaults: Tags[cost] = %q, want %q (level 6 wins)", got.Defaults.Tags["cost"], "high")
	}
	if got.Defaults.Tags["owner"] != "team-a" {
		t.Errorf("AC-10 defaults: Tags[owner] = %q, want %q", got.Defaults.Tags["owner"], "team-a")
	}
	if got.Defaults.Tags["region"] != "us-east-1" {
		t.Errorf("AC-10 defaults: Tags[region] = %q, want %q", got.Defaults.Tags["region"], "us-east-1")
	}
}

// TestMergeCognitoCascade_AC11_SyncedLabelsCascade verifies that
// CognitoConfig.mandatory.syncedLabels cascades to effectiveConfig.mandatory.syncedLabels.
//
// Spec AC-11: CognitoConfig.mandatory.syncedLabels={"app":"cognito"} →
// effectiveConfig.mandatory.syncedLabels={"app":"cognito"}.
func TestMergeCognitoCascade_AC11_SyncedLabelsCascade(t *testing.T) {
	localCognitoCfgMandatory := cascade.CognitoConfigSection{
		SyncedLabels: map[string]string{
			"app": "cognito",
		},
	}

	got := cascade.MergeCognitoCascade(
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	if got.Mandatory.SyncedLabels["app"] != "cognito" {
		t.Errorf("AC-11: Mandatory.SyncedLabels[app] = %q, want %q", got.Mandatory.SyncedLabels["app"], "cognito")
	}
}

// TestMergeCognitoCascade_AC11_SyncedLabelsGlobalWinsOnConflict verifies that
// the global CognitoConfig (level 3) wins over local (level 4) on key conflict.
func TestMergeCognitoCascade_AC11_SyncedLabelsGlobalCognitoWins(t *testing.T) {
	globalCognitoCfgMandatory := cascade.CognitoConfigSection{
		SyncedLabels: map[string]string{
			"app":  "cognito-global",
			"tier": "mandatory",
		},
	}
	localCognitoCfgMandatory := cascade.CognitoConfigSection{
		SyncedLabels: map[string]string{
			"app":  "cognito-local",
			"svc":  "auth",
		},
	}

	got := cascade.MergeCognitoCascade(
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
		globalCognitoCfgMandatory,
		localCognitoCfgMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	// Level 3 wins on "app" key conflict.
	if got.Mandatory.SyncedLabels["app"] != "cognito-global" {
		t.Errorf("AC-11: SyncedLabels[app] = %q, want %q (level 3 wins)", got.Mandatory.SyncedLabels["app"], "cognito-global")
	}
	// "tier" only in level 3.
	if got.Mandatory.SyncedLabels["tier"] != "mandatory" {
		t.Errorf("AC-11: SyncedLabels[tier] = %q, want %q", got.Mandatory.SyncedLabels["tier"], "mandatory")
	}
	// "svc" only in level 4.
	if got.Mandatory.SyncedLabels["svc"] != "auth" {
		t.Errorf("AC-11: SyncedLabels[svc] = %q, want %q", got.Mandatory.SyncedLabels["svc"], "auth")
	}
}

// TestMergeCognitoCascade_AC11_SyncedAnnotationsDefaultsCascade verifies that
// CognitoConfig.defaults.syncedAnnotations flows to effectiveConfig.defaults.syncedAnnotations.
func TestMergeCognitoCascade_AC11_SyncedAnnotationsDefaultsCascade(t *testing.T) {
	localCognitoCfgDefaults := cascade.CognitoConfigSection{
		SyncedAnnotations: map[string]string{
			"iam.amazonaws.com/role": "arn:aws:iam::123456789012:role/cognito-role",
		},
	}

	got := cascade.MergeCognitoCascade(
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgDefaults,
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	const wantAnnotation = "arn:aws:iam::123456789012:role/cognito-role"
	if got.Defaults.SyncedAnnotations["iam.amazonaws.com/role"] != wantAnnotation {
		t.Errorf("AC-11 annotations: got %q, want %q", got.Defaults.SyncedAnnotations["iam.amazonaws.com/role"], wantAnnotation)
	}
}

// TestMergeCognitoCascade_ZeroValues verifies that all-zero inputs produce a zero-value result.
func TestMergeCognitoCascade_ZeroValues(t *testing.T) {
	got := cascade.MergeCognitoCascade(
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
	)

	if got.Mandatory.MfaConfiguration != "" {
		t.Errorf("zero: Mandatory.MfaConfiguration = %q, want %q", got.Mandatory.MfaConfiguration, "")
	}
	if got.Mandatory.PasswordPolicy.MinimumLength != 0 {
		t.Errorf("zero: Mandatory.PasswordPolicy.MinimumLength = %d, want 0", got.Mandatory.PasswordPolicy.MinimumLength)
	}
	if got.Mandatory.PasswordPolicy.RequireLowercase != nil {
		t.Errorf("zero: Mandatory.PasswordPolicy.RequireLowercase = %v, want nil", got.Mandatory.PasswordPolicy.RequireLowercase)
	}
	if got.Defaults.NamingTemplate != "" {
		t.Errorf("zero: Defaults.NamingTemplate = %q, want %q", got.Defaults.NamingTemplate, "")
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("zero: Mandatory.Tags = %v, want empty", got.Mandatory.Tags)
	}
}

// TestMergeCognitoCascade_DefaultsTierString verifies string field cascade in defaults tier
// (level 6 wins over 7, 8, 9).
func TestMergeCognitoCascade_DefaultsTierStringCascade(t *testing.T) {
	localCognitoCfgDefaults := cascade.CognitoConfigSection{
		MfaConfiguration: "OPTIONAL",
	}
	globalCognitoCfgDefaults := cascade.CognitoConfigSection{
		MfaConfiguration: "OFF",
	}
	localKropathDefaults := cascade.CognitoKropathSection{
		MfaConfiguration: "ON",
	}
	globalKropathDefaults := cascade.CognitoKropathSection{
		MfaConfiguration: "ON",
	}

	got := cascade.MergeCognitoCascade(
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgDefaults,
		globalCognitoCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)

	// Level 6 (localCognitoCfgDefaults) wins.
	if got.Defaults.MfaConfiguration != "OPTIONAL" {
		t.Errorf("defaults tier: MfaConfiguration = %q, want %q (level 6 wins)", got.Defaults.MfaConfiguration, "OPTIONAL")
	}
}

// TestMergeCognitoCascade_TemporaryPasswordValidityDaysInt64 verifies that
// TemporaryPasswordValidityDays uses firstNonZeroInt64 (0 = not set).
func TestMergeCognitoCascade_TemporaryPasswordValidityDaysInt64(t *testing.T) {
	localCognitoCfgMandatory := cascade.CognitoConfigSection{
		PasswordPolicy: cascade.CognitoPasswordPolicySection{
			TemporaryPasswordValidityDays: 7,
		},
	}
	localKropathDefaults := cascade.CognitoKropathSection{
		PasswordPolicy: cascade.CognitoPasswordPolicySection{
			TemporaryPasswordValidityDays: 30,
		},
	}

	got := cascade.MergeCognitoCascade(
		emptyCognitoKropathSection(),
		emptyCognitoKropathSection(),
		emptyCognitoConfigSection(),
		localCognitoCfgMandatory,
		emptyCognitoConfigSection(),
		emptyCognitoConfigSection(),
		localKropathDefaults,
		emptyCognitoKropathSection(),
	)

	if got.Mandatory.PasswordPolicy.TemporaryPasswordValidityDays != 7 {
		t.Errorf("int64: Mandatory.PasswordPolicy.TemporaryPasswordValidityDays = %d, want 7", got.Mandatory.PasswordPolicy.TemporaryPasswordValidityDays)
	}
	if got.Defaults.PasswordPolicy.TemporaryPasswordValidityDays != 30 {
		t.Errorf("int64: Defaults.PasswordPolicy.TemporaryPasswordValidityDays = %d, want 30", got.Defaults.PasswordPolicy.TemporaryPasswordValidityDays)
	}
}
