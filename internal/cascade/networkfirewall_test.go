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

// zeroNFWKropath is a zero-value NetworkFirewallKropathSection (absent source).
var zeroNFWKropath = cascade.NetworkFirewallKropathSection{}

// zeroNFWCfg is a zero-value NetworkFirewallConfigSection (absent source).
var zeroNFWCfg = cascade.NetworkFirewallConfigSection{}

// mergeNFWAll calls MergeNetworkFirewallCascade with all eight inputs.
func mergeNFWAll(
	globalKropathMandatory,
	localKropathMandatory cascade.NetworkFirewallKropathSection,
	globalNFWCfgMandatory,
	localNFWCfgMandatory,
	localNFWCfgDefaults,
	globalNFWCfgDefaults cascade.NetworkFirewallConfigSection,
	localKropathDefaults,
	globalKropathDefaults cascade.NetworkFirewallKropathSection,
) cascade.EffectiveNetworkFirewallConfig {
	return cascade.MergeNetworkFirewallCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalNFWCfgMandatory,
		localNFWCfgMandatory,
		localNFWCfgDefaults,
		globalNFWCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeNetworkFirewallCascade_AC1 — globalKropathConfig.mandatory.networkfirewall.encryptionType
// at level 1 propagates to effCfg.mandatory.encryptionType (level 1 wins).
func TestMergeNetworkFirewallCascade_AC1(t *testing.T) {
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{EncryptionType: "CUSTOMER_KMS"}, // level 1
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Mandatory.EncryptionType != "CUSTOMER_KMS" {
		t.Errorf("AC-1: mandatory.encryptionType = %q, want CUSTOMER_KMS (level 1 wins)", got.Mandatory.EncryptionType)
	}
	if got.Defaults.EncryptionType != "" {
		t.Errorf("AC-1: defaults.encryptionType = %q, must not bleed from mandatory", got.Defaults.EncryptionType)
	}
}

// TestMergeNetworkFirewallCascade_AC2 — globalNFWConfig.mandatory.encryptionType at level 3 wins
// when levels 1-2 are empty.
func TestMergeNetworkFirewallCascade_AC2(t *testing.T) {
	got := mergeNFWAll(
		zeroNFWKropath,
		zeroNFWKropath,
		cascade.NetworkFirewallConfigSection{EncryptionType: "CUSTOMER_KMS"}, // level 3
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Mandatory.EncryptionType != "CUSTOMER_KMS" {
		t.Errorf("AC-2: mandatory.encryptionType = %q, want CUSTOMER_KMS (level 3 wins)", got.Mandatory.EncryptionType)
	}
}

// TestMergeNetworkFirewallCascade_AC3 — globalKropathConfig.mandatory.networkfirewall.deleteProtection
// (bool, firstTrue semantics) — true at level 1 propagates to mandatory.
func TestMergeNetworkFirewallCascade_AC3(t *testing.T) {
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{DeleteProtection: true}, // level 1
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if !got.Mandatory.DeleteProtection {
		t.Errorf("AC-3: mandatory.deleteProtection = false, want true (level 1 wins)")
	}
	if got.Defaults.DeleteProtection {
		t.Errorf("AC-3: defaults.deleteProtection = true, must not bleed from mandatory")
	}
}

// TestMergeNetworkFirewallCascade_AC3b — boolean false sentinel: false at level 1 does not
// block a true at level 3 from propagating (firstTrue semantics).
func TestMergeNetworkFirewallCascade_AC3b(t *testing.T) {
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{DeleteProtection: false}, // level 1 — not enforced
		zeroNFWKropath,
		cascade.NetworkFirewallConfigSection{DeleteProtection: true}, // level 3 — enforced
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if !got.Mandatory.DeleteProtection {
		t.Errorf("AC-3b: mandatory.deleteProtection = false, want true (false is not enforced, level 3 wins)")
	}
}

// TestMergeNetworkFirewallCascade_AC4 — globalKropathConfig.mandatory.networkfirewall.statefulRuleOrder
// at level 1 propagates to effCfg.mandatory.statefulRuleOrder.
func TestMergeNetworkFirewallCascade_AC4(t *testing.T) {
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{StatefulRuleOrder: "STRICT_ORDER"}, // level 1
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Mandatory.StatefulRuleOrder != "STRICT_ORDER" {
		t.Errorf("AC-4: mandatory.statefulRuleOrder = %q, want STRICT_ORDER", got.Mandatory.StatefulRuleOrder)
	}
}

// TestMergeNetworkFirewallCascade_AC5 — globalKropathConfig.mandatory.networkfirewall.statefulDefaultActions
// at level 1 propagates (firstNonEmptyStrings semantics).
func TestMergeNetworkFirewallCascade_AC5(t *testing.T) {
	actions := []string{"aws:drop_strict"}
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{StatefulDefaultActions: actions}, // level 1
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if len(got.Mandatory.StatefulDefaultActions) != 1 || got.Mandatory.StatefulDefaultActions[0] != "aws:drop_strict" {
		t.Errorf("AC-5: mandatory.statefulDefaultActions = %v, want [aws:drop_strict]", got.Mandatory.StatefulDefaultActions)
	}
}

// TestMergeNetworkFirewallCascade_AC5b — nil/empty slice sentinel: empty at level 1 does not
// block level 3 from propagating.
func TestMergeNetworkFirewallCascade_AC5b(t *testing.T) {
	actions := []string{"aws:drop_strict"}
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{StatefulDefaultActions: nil}, // level 1 — empty, not enforced
		zeroNFWKropath,
		cascade.NetworkFirewallConfigSection{StatefulDefaultActions: actions}, // level 3
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if len(got.Mandatory.StatefulDefaultActions) != 1 || got.Mandatory.StatefulDefaultActions[0] != "aws:drop_strict" {
		t.Errorf("AC-5b: mandatory.statefulDefaultActions = %v, want [aws:drop_strict] (level 3 wins when level 1 is empty)", got.Mandatory.StatefulDefaultActions)
	}
}

// TestMergeNetworkFirewallCascade_AC6 — globalKropathConfig.mandatory.networkfirewall.streamExceptionPolicy
// at level 1 propagates to effCfg.mandatory.streamExceptionPolicy.
func TestMergeNetworkFirewallCascade_AC6(t *testing.T) {
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{StreamExceptionPolicy: "DROP"}, // level 1
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Mandatory.StreamExceptionPolicy != "DROP" {
		t.Errorf("AC-6: mandatory.streamExceptionPolicy = %q, want DROP", got.Mandatory.StreamExceptionPolicy)
	}
}

// TestMergeNetworkFirewallCascade_AC7 — globalNFWConfig.defaults.encryptionType at level 7 wins
// when mandatory layers are empty and defaults level 6 is empty.
func TestMergeNetworkFirewallCascade_AC7(t *testing.T) {
	got := mergeNFWAll(
		zeroNFWKropath,
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,                                                              // level 6 empty
		cascade.NetworkFirewallConfigSection{EncryptionType: "CUSTOMER_KMS"},    // level 7
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Defaults.EncryptionType != "CUSTOMER_KMS" {
		t.Errorf("AC-7: defaults.encryptionType = %q, want CUSTOMER_KMS (level 7 wins)", got.Defaults.EncryptionType)
	}
	if got.Mandatory.EncryptionType != "" {
		t.Errorf("AC-7: mandatory.encryptionType = %q, must not bleed from defaults", got.Mandatory.EncryptionType)
	}
}

// TestMergeNetworkFirewallCascade_AC8 — globalNFWConfig.mandatory.namingTemplate at level 3
// propagates; KropathConfig levels carry no namingTemplate.
func TestMergeNetworkFirewallCascade_AC8(t *testing.T) {
	got := mergeNFWAll(
		zeroNFWKropath,
		zeroNFWKropath,
		cascade.NetworkFirewallConfigSection{NamingTemplate: "{namespace}-{name}-fw"}, // level 3
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Mandatory.NamingTemplate != "{namespace}-{name}-fw" {
		t.Errorf("AC-8: mandatory.namingTemplate = %q, want {namespace}-{name}-fw", got.Mandatory.NamingTemplate)
	}
}

// TestMergeNetworkFirewallCascade_AC9 — globalNFWConfig.defaults.namingTemplate at level 7
// propagates to effCfg.defaults.namingTemplate.
func TestMergeNetworkFirewallCascade_AC9(t *testing.T) {
	got := mergeNFWAll(
		zeroNFWKropath,
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		cascade.NetworkFirewallConfigSection{NamingTemplate: "{namespace}-{name}-fw"}, // level 7
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Defaults.NamingTemplate != "{namespace}-{name}-fw" {
		t.Errorf("AC-9: defaults.namingTemplate = %q, want {namespace}-{name}-fw", got.Defaults.NamingTemplate)
	}
}

// TestMergeNetworkFirewallCascade_AC10 — Tags union merge: KropathConfig.mandatory.tags and
// NetworkFirewallConfig.mandatory.tags merge additively; KropathConfig wins on key conflict.
func TestMergeNetworkFirewallCascade_AC10(t *testing.T) {
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{Tags: map[string]string{"env": "prod", "owner": "platform"}}, // level 1
		zeroNFWKropath,
		cascade.NetworkFirewallConfigSection{Tags: map[string]string{"env": "staging", "team": "security"}}, // level 3
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Mandatory.Tags["env"] != "prod" {
		t.Errorf("AC-10: mandatory.tags[env] = %q, want prod (level 1 wins on conflict)", got.Mandatory.Tags["env"])
	}
	if got.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("AC-10: mandatory.tags[owner] = %q, want platform", got.Mandatory.Tags["owner"])
	}
	if got.Mandatory.Tags["team"] != "security" {
		t.Errorf("AC-10: mandatory.tags[team] = %q, want security (level 3 key added)", got.Mandatory.Tags["team"])
	}
}

// TestMergeNetworkFirewallCascade_AC11 — all mandatory inputs zero → all mandatory fields zero.
// Confirms zero-value propagation does not pollute the output.
func TestMergeNetworkFirewallCascade_AC11(t *testing.T) {
	got := mergeNFWAll(
		zeroNFWKropath,
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Mandatory.EncryptionType != "" {
		t.Errorf("AC-11: mandatory.encryptionType = %q, want empty", got.Mandatory.EncryptionType)
	}
	if got.Mandatory.DeleteProtection {
		t.Errorf("AC-11: mandatory.deleteProtection = true, want false")
	}
	if got.Mandatory.FirewallPolicyChangeProtection {
		t.Errorf("AC-11: mandatory.firewallPolicyChangeProtection = true, want false")
	}
	if got.Mandatory.SubnetChangeProtection {
		t.Errorf("AC-11: mandatory.subnetChangeProtection = true, want false")
	}
	if got.Mandatory.StatefulRuleOrder != "" {
		t.Errorf("AC-11: mandatory.statefulRuleOrder = %q, want empty", got.Mandatory.StatefulRuleOrder)
	}
	if len(got.Mandatory.StatefulDefaultActions) != 0 {
		t.Errorf("AC-11: mandatory.statefulDefaultActions = %v, want empty", got.Mandatory.StatefulDefaultActions)
	}
	if got.Mandatory.StreamExceptionPolicy != "" {
		t.Errorf("AC-11: mandatory.streamExceptionPolicy = %q, want empty", got.Mandatory.StreamExceptionPolicy)
	}
	if got.Mandatory.NamingTemplate != "" {
		t.Errorf("AC-11: mandatory.namingTemplate = %q, want empty", got.Mandatory.NamingTemplate)
	}
	if len(got.Mandatory.Tags) != 0 {
		t.Errorf("AC-11: mandatory.tags = %v, want empty", got.Mandatory.Tags)
	}
}

// TestMergeNetworkFirewallCascade_level1WinsOverLevel3 — level 1 (globalKropathMandatory)
// takes priority over level 3 (globalNFWConfigMandatory) for encryptionType.
func TestMergeNetworkFirewallCascade_level1WinsOverLevel3(t *testing.T) {
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{EncryptionType: "CUSTOMER_KMS"}, // level 1
		zeroNFWKropath,
		cascade.NetworkFirewallConfigSection{EncryptionType: "AWS_OWNED_KMS"}, // level 3
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Mandatory.EncryptionType != "CUSTOMER_KMS" {
		t.Errorf("level1WinsOverLevel3: mandatory.encryptionType = %q, want CUSTOMER_KMS (level 1 wins)", got.Mandatory.EncryptionType)
	}
}

// TestMergeNetworkFirewallCascade_defaultsLevel6WinsOverLevel7 — level 6 (localNFWConfigDefaults)
// wins over level 7 (globalNFWConfigDefaults) for encryptionType.
func TestMergeNetworkFirewallCascade_defaultsLevel6WinsOverLevel7(t *testing.T) {
	got := mergeNFWAll(
		zeroNFWKropath,
		zeroNFWKropath,
		zeroNFWCfg,
		zeroNFWCfg,
		cascade.NetworkFirewallConfigSection{EncryptionType: "CUSTOMER_KMS"}, // level 6
		cascade.NetworkFirewallConfigSection{EncryptionType: "AWS_OWNED_KMS"}, // level 7
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if got.Defaults.EncryptionType != "CUSTOMER_KMS" {
		t.Errorf("defaults: level 6 wins: encryptionType = %q, want CUSTOMER_KMS", got.Defaults.EncryptionType)
	}
}

// TestMergeNetworkFirewallCascade_syncedLabelsNFWLevelsOnly — SyncedLabels union merge
// from NetworkFirewallConfig levels 3/4 only (not KropathConfig levels).
func TestMergeNetworkFirewallCascade_syncedLabelsNFWLevelsOnly(t *testing.T) {
	got := mergeNFWAll(
		zeroNFWKropath,
		zeroNFWKropath,
		cascade.NetworkFirewallConfigSection{SyncedLabels: map[string]string{"app": "firewall", "tier": "network"}}, // level 3
		cascade.NetworkFirewallConfigSection{SyncedLabels: map[string]string{"app": "local-firewall", "env": "prod"}}, // level 4
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	// Level 3 wins on conflict; level 4 unique keys are included
	if got.Mandatory.SyncedLabels["app"] != "firewall" {
		t.Errorf("syncedLabels: syncedLabels[app] = %q, want firewall (level 3 wins)", got.Mandatory.SyncedLabels["app"])
	}
	if got.Mandatory.SyncedLabels["tier"] != "network" {
		t.Errorf("syncedLabels: syncedLabels[tier] = %q, want network", got.Mandatory.SyncedLabels["tier"])
	}
	if got.Mandatory.SyncedLabels["env"] != "prod" {
		t.Errorf("syncedLabels: syncedLabels[env] = %q, want prod (level 4 key)", got.Mandatory.SyncedLabels["env"])
	}
}

// TestMergeNetworkFirewallCascade_booleanProtectionAllFields — all three protection booleans
// propagate independently via firstTrue.
func TestMergeNetworkFirewallCascade_booleanProtectionAllFields(t *testing.T) {
	got := mergeNFWAll(
		cascade.NetworkFirewallKropathSection{
			DeleteProtection:               false,
			FirewallPolicyChangeProtection: true,
			SubnetChangeProtection:         false,
		}, // level 1
		zeroNFWKropath,
		cascade.NetworkFirewallConfigSection{
			DeleteProtection:               true,
			FirewallPolicyChangeProtection: false,
			SubnetChangeProtection:         true,
		}, // level 3
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWCfg,
		zeroNFWKropath,
		zeroNFWKropath,
	)

	if !got.Mandatory.DeleteProtection {
		t.Errorf("boolAll: deleteProtection = false, want true (level 3 wins, level 1 is false=not enforced)")
	}
	if !got.Mandatory.FirewallPolicyChangeProtection {
		t.Errorf("boolAll: firewallPolicyChangeProtection = false, want true (level 1 wins)")
	}
	if !got.Mandatory.SubnetChangeProtection {
		t.Errorf("boolAll: subnetChangeProtection = false, want true (level 3 wins, level 1 is false=not enforced)")
	}
}
