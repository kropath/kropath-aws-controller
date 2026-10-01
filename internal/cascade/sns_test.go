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

// zeroSNSKropath is a zero-value SNSKropathSection (absent KropathConfig source).
var zeroSNSKropath = cascade.SNSKropathSection{}

// zeroSNSCfg is a zero-value SNSConfigSection (absent SNSConfig source).
var zeroSNSCfg = cascade.SNSConfigSection{}

// mergeSNSAll calls MergeSNSCascade with all eight inputs.
func mergeSNSAll(
	globalKropathMandatory,
	localKropathMandatory cascade.SNSKropathSection,
	globalSNSCfgMandatory,
	localSNSCfgMandatory,
	localSNSCfgDefaults,
	globalSNSCfgDefaults cascade.SNSConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.SNSKropathSection,
) cascade.EffectiveSNSConfig {
	return cascade.MergeSNSCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalSNSCfgMandatory,
		localSNSCfgMandatory,
		localSNSCfgDefaults,
		globalSNSCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeSNSCascade_AC1 — globalKropathConfig.mandatory.sns.kmsMasterKeyId at level 1
// propagates to effCfg.mandatory.kmsMasterKeyId (level 1 wins).
func TestMergeSNSCascade_AC1(t *testing.T) {
	const key = "arn:aws:kms:ap-southeast-2:123:key/org-key"
	got := mergeSNSAll(
		cascade.SNSKropathSection{KmsMasterKeyId: key}, // level 1
		zeroSNSKropath,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.KmsMasterKeyId != key {
		t.Errorf("AC-1: mandatory.kmsMasterKeyId = %q, want %q", got.Mandatory.KmsMasterKeyId, key)
	}
	if got.Defaults.KmsMasterKeyId != "" {
		t.Errorf("AC-1: defaults.kmsMasterKeyId = %q, must not bleed from mandatory", got.Defaults.KmsMasterKeyId)
	}
}

// TestMergeSNSCascade_AC2 — globalSNSConfig.mandatory.kmsMasterKeyId at level 3 propagates
// when levels 1-2 are empty (level 3 wins when 1-2 absent).
func TestMergeSNSCascade_AC2(t *testing.T) {
	const key = "alias/aws/sns"
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		cascade.SNSConfigSection{KmsMasterKeyId: key}, // level 3
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.KmsMasterKeyId != key {
		t.Errorf("AC-2: mandatory.kmsMasterKeyId = %q, want %q (level 3 wins when 1-2 absent)", got.Mandatory.KmsMasterKeyId, key)
	}
}

// TestMergeSNSCascade_AC3 — localSNSConfig.defaults.kmsMasterKeyId at level 6 propagates
// when all mandatory tiers are empty.
func TestMergeSNSCascade_AC3(t *testing.T) {
	const key = "alias/aws/sns"
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		zeroSNSCfg,
		zeroSNSCfg,
		cascade.SNSConfigSection{KmsMasterKeyId: key}, // level 6
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.KmsMasterKeyId != "" {
		t.Errorf("AC-3: mandatory.kmsMasterKeyId = %q, want empty", got.Mandatory.KmsMasterKeyId)
	}
	if got.Defaults.KmsMasterKeyId != key {
		t.Errorf("AC-3: defaults.kmsMasterKeyId = %q, want %q (level 6)", got.Defaults.KmsMasterKeyId, key)
	}
}

// TestMergeSNSCascade_AC4 — globalKropathConfig.mandatory.sns.signatureVersion="2" at
// level 1 propagates to effCfg.mandatory.signatureVersion.
func TestMergeSNSCascade_AC4(t *testing.T) {
	got := mergeSNSAll(
		cascade.SNSKropathSection{SignatureVersion: "2"}, // level 1
		zeroSNSKropath,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.SignatureVersion != "2" {
		t.Errorf("AC-4: mandatory.signatureVersion = %q, want 2 (level 1)", got.Mandatory.SignatureVersion)
	}
}

// TestMergeSNSCascade_AC5 — globalSNSConfig.defaults.signatureVersion="2" (level 7) wins
// over globalKropathConfig.defaults.sns.signatureVersion="1" (level 9).
func TestMergeSNSCascade_AC5(t *testing.T) {
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		cascade.SNSConfigSection{SignatureVersion: "2"},  // level 7
		zeroSNSKropath,
		cascade.SNSKropathSection{SignatureVersion: "1"}, // level 9
	)

	if got.Defaults.SignatureVersion != "2" {
		t.Errorf("AC-5: defaults.signatureVersion = %q, want 2 (level 7 wins over level 9)", got.Defaults.SignatureVersion)
	}
}

// TestMergeSNSCascade_AC6 — globalKropathConfig.mandatory.sns.tracingConfig="Active" at
// level 1 propagates to effCfg.mandatory.tracingConfig.
func TestMergeSNSCascade_AC6(t *testing.T) {
	got := mergeSNSAll(
		cascade.SNSKropathSection{TracingConfig: "Active"}, // level 1
		zeroSNSKropath,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.TracingConfig != "Active" {
		t.Errorf("AC-6: mandatory.tracingConfig = %q, want Active (level 1)", got.Mandatory.TracingConfig)
	}
}

// TestMergeSNSCascade_AC7 — globalSNSConfig.mandatory.dataProtectionPolicy at level 3
// propagates (KropathConfig.sns has no dataProtectionPolicy counterpart at levels 1-2/8-9).
func TestMergeSNSCascade_AC7(t *testing.T) {
	const policy = `{"Name":"org-policy","Statements":[]}`
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		cascade.SNSConfigSection{DataProtectionPolicy: policy}, // level 3
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.DataProtectionPolicy != policy {
		t.Errorf("AC-7: mandatory.dataProtectionPolicy = %q, want %q (level 3)", got.Mandatory.DataProtectionPolicy, policy)
	}
}

// TestMergeSNSCascade_AC8 — globalSNSConfig.mandatory.deliveryFeedback.http at level 3
// propagates all three HTTP feedback fields to effCfg.mandatory.deliveryFeedback.http.
func TestMergeSNSCascade_AC8(t *testing.T) {
	feedback := cascade.SNSConfigSection{
		DeliveryFeedback: cascade.DeliveryFeedback{
			HTTP: &cascade.DeliveryFeedbackProtocol{
				SuccessFeedbackRoleArn:    "arn:aws:iam::123:role/http-log",
				FailureFeedbackRoleArn:    "arn:aws:iam::123:role/http-log-fail",
				SuccessFeedbackSampleRate: "100",
			},
		},
	}
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		feedback, // level 3
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.DeliveryFeedback == nil {
		t.Fatal("AC-8: mandatory.deliveryFeedback is nil, want non-nil")
	}
	if got.Mandatory.DeliveryFeedback.HTTP == nil {
		t.Fatal("AC-8: mandatory.deliveryFeedback.http is nil, want non-nil")
	}
	http := got.Mandatory.DeliveryFeedback.HTTP
	if http.SuccessFeedbackRoleArn != "arn:aws:iam::123:role/http-log" {
		t.Errorf("AC-8: http.successFeedbackRoleArn = %q, want arn:aws:iam::123:role/http-log", http.SuccessFeedbackRoleArn)
	}
	if http.FailureFeedbackRoleArn != "arn:aws:iam::123:role/http-log-fail" {
		t.Errorf("AC-8: http.failureFeedbackRoleArn = %q, want arn:aws:iam::123:role/http-log-fail", http.FailureFeedbackRoleArn)
	}
	if http.SuccessFeedbackSampleRate != "100" {
		t.Errorf("AC-8: http.successFeedbackSampleRate = %q, want 100", http.SuccessFeedbackSampleRate)
	}
	if got.Mandatory.DeliveryFeedback.SQS != nil {
		t.Errorf("AC-8: mandatory.deliveryFeedback.sqs must be nil (not set)")
	}
}

// TestMergeSNSCascade_AC9 — globalSNSConfig.defaults.deliveryFeedback.sqs.successFeedbackRoleArn
// at level 7 propagates to effCfg.defaults.deliveryFeedback.sqs.
func TestMergeSNSCascade_AC9(t *testing.T) {
	const roleArn = "arn:aws:iam::123:role/sqs-log-default"
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		cascade.SNSConfigSection{ // level 7
			DeliveryFeedback: cascade.DeliveryFeedback{
				SQS: &cascade.DeliveryFeedbackProtocol{
					SuccessFeedbackRoleArn: roleArn,
				},
			},
		},
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Defaults.DeliveryFeedback == nil {
		t.Fatal("AC-9: defaults.deliveryFeedback is nil, want non-nil")
	}
	if got.Defaults.DeliveryFeedback.SQS == nil {
		t.Fatal("AC-9: defaults.deliveryFeedback.sqs is nil, want non-nil")
	}
	if got.Defaults.DeliveryFeedback.SQS.SuccessFeedbackRoleArn != roleArn {
		t.Errorf("AC-9: defaults.deliveryFeedback.sqs.successFeedbackRoleArn = %q, want %q (level 7)", got.Defaults.DeliveryFeedback.SQS.SuccessFeedbackRoleArn, roleArn)
	}
}

// TestMergeSNSCascade_AC10 — globalSNSConfig.defaults.namingTemplate at level 7 propagates
// (KropathConfig.sns has no namingTemplate counterpart at levels 1-2/8-9).
func TestMergeSNSCascade_AC10(t *testing.T) {
	const tmpl = "{namespace}-{name}"
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		cascade.SNSConfigSection{NamingTemplate: tmpl}, // level 7
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Defaults.NamingTemplate != tmpl {
		t.Errorf("AC-10: defaults.namingTemplate = %q, want %q (level 7)", got.Defaults.NamingTemplate, tmpl)
	}
	// KropathConfig.sns has no namingTemplate — level 9 must not carry it.
	got2 := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		cascade.SNSKropathSection{KmsMasterKeyId: "anything"}, // level 9 — no namingTemplate
	)
	if got2.Defaults.NamingTemplate != "" {
		t.Errorf("AC-10: KropathConfig.sns must not carry namingTemplate; got %q", got2.Defaults.NamingTemplate)
	}
}

// TestMergeSNSCascade_AC11 — tags from KropathConfig.mandatory.tags and
// SNSConfig.mandatory.tags are union-merged into effCfg.mandatory.tags.
func TestMergeSNSCascade_AC11(t *testing.T) {
	got := mergeSNSAll(
		cascade.SNSKropathSection{Tags: map[string]string{"cost-centre": "infra"}}, // level 1
		zeroSNSKropath,
		cascade.SNSConfigSection{Tags: map[string]string{"topic-type": "messaging"}}, // level 3
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("AC-11: mandatory.tags[cost-centre] = %q, want infra", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["topic-type"] != "messaging" {
		t.Errorf("AC-11: mandatory.tags[topic-type] = %q, want messaging (additive union)", got.Mandatory.Tags["topic-type"])
	}
}

// TestMergeSNSCascade_AC12 — globalSNSConfig.mandatory.syncedLabels at level 3 propagates
// to effCfg.mandatory.syncedLabels.
func TestMergeSNSCascade_AC12(t *testing.T) {
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		cascade.SNSConfigSection{SyncedLabels: map[string]string{"data-class": "internal"}}, // level 3
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "internal" {
		t.Errorf("AC-12: mandatory.syncedLabels[data-class] = %q, want internal (level 3)", got.Mandatory.SyncedLabels["data-class"])
	}
}

// TestMergeSNSCascade_AllAbsent — zero inputs produce zero output;
// deliveryFeedback pointers must be nil (not empty structs).
func TestMergeSNSCascade_AllAbsent(t *testing.T) {
	got := mergeSNSAll(
		zeroSNSKropath, zeroSNSKropath,
		zeroSNSCfg, zeroSNSCfg,
		zeroSNSCfg, zeroSNSCfg,
		zeroSNSKropath, zeroSNSKropath,
	)

	if got.Mandatory.KmsMasterKeyId != "" {
		t.Errorf("AllAbsent: mandatory.kmsMasterKeyId = %q, want empty", got.Mandatory.KmsMasterKeyId)
	}
	if got.Mandatory.DeliveryFeedback != nil {
		t.Errorf("AllAbsent: mandatory.deliveryFeedback should be nil, got non-nil")
	}
	if got.Defaults.DeliveryFeedback != nil {
		t.Errorf("AllAbsent: defaults.deliveryFeedback should be nil, got non-nil")
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("AllAbsent: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
}

// TestMergeSNSCascade_MandatoryPriorityOrder verifies that lower level numbers win for
// mandatory scalar fields (L1 > L2 > L3 > L4) for kmsMasterKeyId.
func TestMergeSNSCascade_MandatoryPriorityOrder(t *testing.T) {
	cases := []struct {
		name                   string
		globalKropathMandatory cascade.SNSKropathSection
		localKropathMandatory  cascade.SNSKropathSection
		globalSNSCfgMandatory  cascade.SNSConfigSection
		localSNSCfgMandatory   cascade.SNSConfigSection
		wantKmsKey             string
	}{
		{
			name:                   "level1-wins",
			globalKropathMandatory: cascade.SNSKropathSection{KmsMasterKeyId: "level1"},
			localKropathMandatory:  cascade.SNSKropathSection{KmsMasterKeyId: "level2"},
			globalSNSCfgMandatory:  cascade.SNSConfigSection{KmsMasterKeyId: "level3"},
			localSNSCfgMandatory:   cascade.SNSConfigSection{KmsMasterKeyId: "level4"},
			wantKmsKey:             "level1",
		},
		{
			name:                  "level2-wins-when-1-absent",
			localKropathMandatory: cascade.SNSKropathSection{KmsMasterKeyId: "level2"},
			globalSNSCfgMandatory: cascade.SNSConfigSection{KmsMasterKeyId: "level3"},
			localSNSCfgMandatory:  cascade.SNSConfigSection{KmsMasterKeyId: "level4"},
			wantKmsKey:            "level2",
		},
		{
			name:                  "level3-wins-when-1-2-absent",
			globalSNSCfgMandatory: cascade.SNSConfigSection{KmsMasterKeyId: "level3"},
			localSNSCfgMandatory:  cascade.SNSConfigSection{KmsMasterKeyId: "level4"},
			wantKmsKey:            "level3",
		},
		{
			name:                 "level4-fallback",
			localSNSCfgMandatory: cascade.SNSConfigSection{KmsMasterKeyId: "level4"},
			wantKmsKey:           "level4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeSNSAll(
				tc.globalKropathMandatory,
				tc.localKropathMandatory,
				tc.globalSNSCfgMandatory,
				tc.localSNSCfgMandatory,
				zeroSNSCfg, zeroSNSCfg,
				zeroSNSKropath, zeroSNSKropath,
			)
			if got.Mandatory.KmsMasterKeyId != tc.wantKmsKey {
				t.Errorf("mandatory.kmsMasterKeyId = %q, want %q", got.Mandatory.KmsMasterKeyId, tc.wantKmsKey)
			}
		})
	}
}

// TestMergeSNSCascade_DefaultsPriorityOrder verifies that lower level numbers win for
// defaults scalar fields (L6 > L7 > L8 > L9).
func TestMergeSNSCascade_DefaultsPriorityOrder(t *testing.T) {
	cases := []struct {
		name                  string
		localSNSCfgDefaults   cascade.SNSConfigSection
		globalSNSCfgDefaults  cascade.SNSConfigSection
		localKropathDefaults  cascade.SNSKropathSection
		globalKropathDefaults cascade.SNSKropathSection
		wantKmsKey            string
	}{
		{
			name:                  "level6-wins",
			localSNSCfgDefaults:   cascade.SNSConfigSection{KmsMasterKeyId: "level6"},
			globalSNSCfgDefaults:  cascade.SNSConfigSection{KmsMasterKeyId: "level7"},
			localKropathDefaults:  cascade.SNSKropathSection{KmsMasterKeyId: "level8"},
			globalKropathDefaults: cascade.SNSKropathSection{KmsMasterKeyId: "level9"},
			wantKmsKey:            "level6",
		},
		{
			name:                  "level7-wins-when-6-absent",
			globalSNSCfgDefaults:  cascade.SNSConfigSection{KmsMasterKeyId: "level7"},
			localKropathDefaults:  cascade.SNSKropathSection{KmsMasterKeyId: "level8"},
			globalKropathDefaults: cascade.SNSKropathSection{KmsMasterKeyId: "level9"},
			wantKmsKey:            "level7",
		},
		{
			name:                  "level8-wins-when-6-7-absent",
			localKropathDefaults:  cascade.SNSKropathSection{KmsMasterKeyId: "level8"},
			globalKropathDefaults: cascade.SNSKropathSection{KmsMasterKeyId: "level9"},
			wantKmsKey:            "level8",
		},
		{
			name:                  "level9-fallback",
			globalKropathDefaults: cascade.SNSKropathSection{KmsMasterKeyId: "level9"},
			wantKmsKey:            "level9",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mergeSNSAll(
				zeroSNSKropath, zeroSNSKropath,
				zeroSNSCfg, zeroSNSCfg,
				tc.localSNSCfgDefaults,
				tc.globalSNSCfgDefaults,
				tc.localKropathDefaults,
				tc.globalKropathDefaults,
			)
			if got.Defaults.KmsMasterKeyId != tc.wantKmsKey {
				t.Errorf("defaults.kmsMasterKeyId = %q, want %q", got.Defaults.KmsMasterKeyId, tc.wantKmsKey)
			}
		})
	}
}

// TestMergeSNSCascade_MandatoryIsolatedFromDefaults — mandatory fields must not bleed
// into defaults and vice versa.
func TestMergeSNSCascade_MandatoryIsolatedFromDefaults(t *testing.T) {
	got := mergeSNSAll(
		cascade.SNSKropathSection{KmsMasterKeyId: "kms-mandatory", TracingConfig: "Active"},
		zeroSNSKropath,
		cascade.SNSConfigSection{DataProtectionPolicy: `{"Name":"org"}`},
		zeroSNSCfg,
		cascade.SNSConfigSection{KmsMasterKeyId: "kms-default", NamingTemplate: "{namespace}-{name}"}, // defaults level 6
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.KmsMasterKeyId != "kms-mandatory" {
		t.Errorf("mandatory.kmsMasterKeyId = %q, want kms-mandatory", got.Mandatory.KmsMasterKeyId)
	}
	if got.Mandatory.TracingConfig != "Active" {
		t.Errorf("mandatory.tracingConfig = %q, want Active", got.Mandatory.TracingConfig)
	}
	if got.Mandatory.DataProtectionPolicy != `{"Name":"org"}` {
		t.Errorf("mandatory.dataProtectionPolicy = %q, want org policy", got.Mandatory.DataProtectionPolicy)
	}
	if got.Defaults.KmsMasterKeyId != "kms-default" {
		t.Errorf("defaults.kmsMasterKeyId = %q, want kms-default", got.Defaults.KmsMasterKeyId)
	}
	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Errorf("defaults.namingTemplate = %q, want {namespace}-{name}", got.Defaults.NamingTemplate)
	}
	if got.Defaults.TracingConfig != "" {
		t.Errorf("defaults.tracingConfig = %q, must not bleed from mandatory", got.Defaults.TracingConfig)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("mandatory.namingTemplate = %q, must not bleed from defaults", got.Mandatory.NamingTemplate)
	}
}

// TestMergeSNSCascade_TagUnionMerge — tags from KropathConfig and SNSConfig levels are
// union-merged; higher-priority (lower-level number) source wins on key conflict.
func TestMergeSNSCascade_TagUnionMerge(t *testing.T) {
	got := mergeSNSAll(
		cascade.SNSKropathSection{Tags: map[string]string{"owner": "platform", "cost-centre": "infra"}}, // level 1 (wins)
		zeroSNSKropath,
		cascade.SNSConfigSection{Tags: map[string]string{"cost-centre": "payments", "env": "prod"}}, // level 3
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.Tags["cost-centre"] != "infra" {
		t.Errorf("tag union: mandatory.tags[cost-centre] = %q, want infra (level 1 wins)", got.Mandatory.Tags["cost-centre"])
	}
	if got.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("tag union: mandatory.tags[owner] = %q, want platform", got.Mandatory.Tags["owner"])
	}
	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("tag union: mandatory.tags[env] = %q, want prod (additive)", got.Mandatory.Tags["env"])
	}
}

// TestMergeSNSCascade_SyncedLabelsUnionMerge — syncedLabels from SNSConfig levels only are
// union-merged; level 3 wins over level 4 on key conflict.
func TestMergeSNSCascade_SyncedLabelsUnionMerge(t *testing.T) {
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		cascade.SNSConfigSection{SyncedLabels: map[string]string{"team": "platform", "data-class": "public"}},   // level 3 (wins)
		cascade.SNSConfigSection{SyncedLabels: map[string]string{"data-class": "internal", "region": "ap-se"}}, // level 4
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.SyncedLabels["data-class"] != "public" {
		t.Errorf("syncedLabels union: mandatory.syncedLabels[data-class] = %q, want public (level 3 wins)", got.Mandatory.SyncedLabels["data-class"])
	}
	if got.Mandatory.SyncedLabels["team"] != "platform" {
		t.Errorf("syncedLabels union: mandatory.syncedLabels[team] = %q, want platform", got.Mandatory.SyncedLabels["team"])
	}
	if got.Mandatory.SyncedLabels["region"] != "ap-se" {
		t.Errorf("syncedLabels union: mandatory.syncedLabels[region] = %q, want ap-se (additive from level 4)", got.Mandatory.SyncedLabels["region"])
	}
}

// TestMergeSNSCascade_DeliveryFeedbackPerKeyMerge — per-key merge semantics: each leaf
// field within each deliveryFeedback protocol is resolved independently. Level 3 (global
// SNSConfig mandatory) wins per leaf; fields only in level 4 are additive.
func TestMergeSNSCascade_DeliveryFeedbackPerKeyMerge(t *testing.T) {
	got := mergeSNSAll(
		zeroSNSKropath,
		zeroSNSKropath,
		cascade.SNSConfigSection{ // level 3: sets http.successArn, lambda.successArn
			DeliveryFeedback: cascade.DeliveryFeedback{
				HTTP: &cascade.DeliveryFeedbackProtocol{
					SuccessFeedbackRoleArn: "arn:l3:http-success",
				},
				Lambda: &cascade.DeliveryFeedbackProtocol{
					SuccessFeedbackRoleArn: "arn:l3:lambda-success",
				},
			},
		},
		cascade.SNSConfigSection{ // level 4: http.successArn loses to L3; http.failureArn is additive
			DeliveryFeedback: cascade.DeliveryFeedback{
				HTTP: &cascade.DeliveryFeedbackProtocol{
					SuccessFeedbackRoleArn: "arn:l4:http-success",
					FailureFeedbackRoleArn: "arn:l4:http-failure",
				},
			},
		},
		zeroSNSCfg,
		zeroSNSCfg,
		zeroSNSKropath,
		zeroSNSKropath,
	)

	if got.Mandatory.DeliveryFeedback == nil {
		t.Fatal("per-key: mandatory.deliveryFeedback is nil")
	}
	http := got.Mandatory.DeliveryFeedback.HTTP
	if http == nil {
		t.Fatal("per-key: mandatory.deliveryFeedback.http is nil")
	}
	if http.SuccessFeedbackRoleArn != "arn:l3:http-success" {
		t.Errorf("per-key: http.successFeedbackRoleArn = %q, want arn:l3:http-success (level 3 wins)", http.SuccessFeedbackRoleArn)
	}
	if http.FailureFeedbackRoleArn != "arn:l4:http-failure" {
		t.Errorf("per-key: http.failureFeedbackRoleArn = %q, want arn:l4:http-failure (additive from level 4)", http.FailureFeedbackRoleArn)
	}
	lambda := got.Mandatory.DeliveryFeedback.Lambda
	if lambda == nil {
		t.Fatal("per-key: mandatory.deliveryFeedback.lambda is nil")
	}
	if lambda.SuccessFeedbackRoleArn != "arn:l3:lambda-success" {
		t.Errorf("per-key: lambda.successFeedbackRoleArn = %q, want arn:l3:lambda-success", lambda.SuccessFeedbackRoleArn)
	}
	if got.Mandatory.DeliveryFeedback.SQS != nil {
		t.Errorf("per-key: mandatory.deliveryFeedback.sqs must be nil (not set)")
	}
}

// TestMergeSNSCascade_DeliveryFeedbackNilWhenAllAbsent — when no deliveryFeedback is
// set in any source, the merged pointer must be nil (not an empty struct).
func TestMergeSNSCascade_DeliveryFeedbackNilWhenAllAbsent(t *testing.T) {
	got := mergeSNSAll(
		zeroSNSKropath, zeroSNSKropath,
		zeroSNSCfg, zeroSNSCfg,
		zeroSNSCfg, zeroSNSCfg,
		zeroSNSKropath, zeroSNSKropath,
	)
	if got.Mandatory.DeliveryFeedback != nil {
		t.Errorf("DeliveryFeedbackNilWhenAllAbsent: mandatory.deliveryFeedback must be nil, got non-nil")
	}
	if got.Defaults.DeliveryFeedback != nil {
		t.Errorf("DeliveryFeedbackNilWhenAllAbsent: defaults.deliveryFeedback must be nil, got non-nil")
	}
}
