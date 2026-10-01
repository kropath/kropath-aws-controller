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
	zeroS3AdvKropath    = cascade.S3AdvancedKropathSection{}
	zeroS3AdvCfg        = cascade.S3AdvancedConfigSection{}
)

// mergeS3Advanced is a test helper that calls MergeS3AdvancedCascade with
// named slots for readability.
func mergeS3Advanced(
	globalKropathMandatory, localKropathMandatory cascade.S3AdvancedKropathSection,
	globalS3AdvCfgMandatory, localS3AdvCfgMandatory cascade.S3AdvancedConfigSection,
	localS3AdvCfgDefaults, globalS3AdvCfgDefaults cascade.S3AdvancedConfigSection,
	localKropathDefaults, globalKropathDefaults cascade.S3AdvancedKropathSection,
) cascade.EffectiveS3AdvancedConfig {
	return cascade.MergeS3AdvancedCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalS3AdvCfgMandatory,
		localS3AdvCfgMandatory,
		localS3AdvCfgDefaults,
		globalS3AdvCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)
}

// TestMergeS3AdvancedCascade_ZeroValuePassthrough verifies that an all-zero
// input yields an all-zero effective config (no spurious enforcement).
func TestMergeS3AdvancedCascade_ZeroValuePassthrough(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Mandatory.TableBucketEncryption.SseAlgorithm != "" {
		t.Fatalf("mandatory.tableBucketEncryption.sseAlgorithm = %q, want empty", got.Mandatory.TableBucketEncryption.SseAlgorithm)
	}
	if got.Mandatory.AccessPointVpcOnly {
		t.Fatal("mandatory.accessPointVpcOnly should be false for zero input")
	}
	if got.Defaults.NamingTemplate != "" {
		t.Fatalf("defaults.namingTemplate = %q, want empty", got.Defaults.NamingTemplate)
	}
}

// TestMergeS3AdvancedCascade_MandatoryTableBucketEncryption covers AC-2:
// mandatory.tableBucketEncryption fields pass through from S3AdvancedConfig.
func TestMergeS3AdvancedCascade_MandatoryTableBucketEncryption(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		cascade.S3AdvancedConfigSection{
			TableBucketEncryption: cascade.TableBucketEncryptionSection{
				SseAlgorithm: "aws:kms",
				KmsKeyARN:    "arn:aws:kms:us-east-1:123:key/pci",
			},
		},
		zeroS3AdvCfg,
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Mandatory.TableBucketEncryption.SseAlgorithm != "aws:kms" {
		t.Fatalf("mandatory.tableBucketEncryption.sseAlgorithm = %q, want %q", got.Mandatory.TableBucketEncryption.SseAlgorithm, "aws:kms")
	}
	if got.Mandatory.TableBucketEncryption.KmsKeyARN != "arn:aws:kms:us-east-1:123:key/pci" {
		t.Fatalf("mandatory.tableBucketEncryption.kmsKeyARN = %q, want pci key", got.Mandatory.TableBucketEncryption.KmsKeyARN)
	}
}

// TestMergeS3AdvancedCascade_DefaultsTableBucketEncryption covers AC-1:
// defaults.tableBucketEncryption fields pass through from S3AdvancedConfig.
func TestMergeS3AdvancedCascade_DefaultsTableBucketEncryption(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{
			TableBucketEncryption: cascade.TableBucketEncryptionSection{
				SseAlgorithm: "aws:kms",
				KmsKeyARN:    "arn:aws:kms:us-east-1:123:key/abc",
			},
		},
		zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Defaults.TableBucketEncryption.SseAlgorithm != "aws:kms" {
		t.Fatalf("defaults.tableBucketEncryption.sseAlgorithm = %q, want %q", got.Defaults.TableBucketEncryption.SseAlgorithm, "aws:kms")
	}
	if got.Defaults.TableBucketEncryption.KmsKeyARN != "arn:aws:kms:us-east-1:123:key/abc" {
		t.Fatalf("defaults.tableBucketEncryption.kmsKeyARN = %q, want abc key", got.Defaults.TableBucketEncryption.KmsKeyARN)
	}
}

// TestMergeS3AdvancedCascade_DefaultsVectorEncryption covers AC-5:
// defaults.vectorBucketEncryption and defaults.vectorIndexEncryption pass through.
func TestMergeS3AdvancedCascade_DefaultsVectorEncryption(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{
			VectorBucketEncryption: cascade.VectorBucketEncryptionSection{
				SseType:   "aws:kms",
				KmsKeyARN: "arn:aws:kms:us-east-1:123:key/vec-bucket",
			},
			VectorIndexEncryption: cascade.VectorIndexEncryptionSection{
				SseType:   "aws:kms",
				KmsKeyARN: "arn:aws:kms:us-east-1:123:key/vec-index",
			},
		},
		zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Defaults.VectorBucketEncryption.SseType != "aws:kms" {
		t.Fatalf("defaults.vectorBucketEncryption.sseType = %q, want %q", got.Defaults.VectorBucketEncryption.SseType, "aws:kms")
	}
	if got.Defaults.VectorBucketEncryption.KmsKeyARN != "arn:aws:kms:us-east-1:123:key/vec-bucket" {
		t.Fatal("defaults.vectorBucketEncryption.kmsKeyARN mismatch")
	}
	if got.Defaults.VectorIndexEncryption.SseType != "aws:kms" {
		t.Fatalf("defaults.vectorIndexEncryption.sseType = %q, want %q", got.Defaults.VectorIndexEncryption.SseType, "aws:kms")
	}
	if got.Defaults.VectorIndexEncryption.KmsKeyARN != "arn:aws:kms:us-east-1:123:key/vec-index" {
		t.Fatal("defaults.vectorIndexEncryption.kmsKeyARN mismatch")
	}
}

// TestMergeS3AdvancedCascade_DefaultsFilesEncryption covers AC-6:
// defaults.filesEncryption.kmsKeyID passes through.
func TestMergeS3AdvancedCascade_DefaultsFilesEncryption(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{
			FilesEncryption: cascade.FilesEncryptionSection{
				KmsKeyID: "arn:aws:kms:us-east-1:123:key/files",
			},
		},
		zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Defaults.FilesEncryption.KmsKeyID != "arn:aws:kms:us-east-1:123:key/files" {
		t.Fatalf("defaults.filesEncryption.kmsKeyID = %q, want files key", got.Defaults.FilesEncryption.KmsKeyID)
	}
}

// TestMergeS3AdvancedCascade_DefaultsAccessPointPublicAccessBlock covers AC-3:
// all four boolean fields in defaults.accessPointPublicAccessBlock pass through.
func TestMergeS3AdvancedCascade_DefaultsAccessPointPublicAccessBlock(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{
			AccessPointPublicAccessBlock: cascade.AccessPointPublicAccessBlockSection{
				BlockPublicACLs:       true,
				BlockPublicPolicy:     true,
				IgnorePublicACLs:      true,
				RestrictPublicBuckets: true,
			},
		},
		zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	block := got.Defaults.AccessPointPublicAccessBlock
	if !block.BlockPublicACLs {
		t.Fatal("defaults.accessPointPublicAccessBlock.blockPublicACLs should be true")
	}
	if !block.BlockPublicPolicy {
		t.Fatal("defaults.accessPointPublicAccessBlock.blockPublicPolicy should be true")
	}
	if !block.IgnorePublicACLs {
		t.Fatal("defaults.accessPointPublicAccessBlock.ignorePublicACLs should be true")
	}
	if !block.RestrictPublicBuckets {
		t.Fatal("defaults.accessPointPublicAccessBlock.restrictPublicBuckets should be true")
	}
}

// TestMergeS3AdvancedCascade_KropathConfigMandatoryWins covers AC-4:
// KropathConfig.mandatory wins over S3AdvancedConfig.defaults for
// accessPointPublicAccessBlock fields.
func TestMergeS3AdvancedCascade_KropathConfigMandatoryWins(t *testing.T) {
	got := mergeS3Advanced(
		cascade.S3AdvancedKropathSection{
			AccessPointPublicAccessBlock: cascade.AccessPointPublicAccessBlockSection{
				BlockPublicACLs: true,
			},
		},
		zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{
			AccessPointPublicAccessBlock: cascade.AccessPointPublicAccessBlockSection{
				BlockPublicACLs: true,
			},
		},
		zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	// The mandatory tier should pick up the KropathConfig.mandatory value
	if !got.Mandatory.AccessPointPublicAccessBlock.BlockPublicACLs {
		t.Fatal("mandatory.accessPointPublicAccessBlock.blockPublicACLs should be true from KropathConfig.mandatory")
	}
}

// TestMergeS3AdvancedCascade_MandatoryLevel1WinsOverLevel4 verifies that
// KropathConfig.mandatory (level 1) wins over S3AdvancedConfig.mandatory (level 4)
// when both set accessPointPublicAccessBlock.
func TestMergeS3AdvancedCascade_MandatoryLevel1WinsOverLevel4(t *testing.T) {
	got := mergeS3Advanced(
		cascade.S3AdvancedKropathSection{
			AccessPointPublicAccessBlock: cascade.AccessPointPublicAccessBlockSection{
				BlockPublicACLs: true,
			},
			AccessPointVpcOnly: true,
		},
		zeroS3AdvKropath,
		zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{
			// local S3AdvancedConfig also sets these — level 4 should not override level 1
			AccessPointPublicAccessBlock: cascade.AccessPointPublicAccessBlockSection{
				BlockPublicACLs: true,
			},
			AccessPointVpcOnly: true,
		},
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if !got.Mandatory.AccessPointPublicAccessBlock.BlockPublicACLs {
		t.Fatal("mandatory.blockPublicACLs should be true")
	}
	if !got.Mandatory.AccessPointVpcOnly {
		t.Fatal("mandatory.accessPointVpcOnly should be true")
	}
}

// TestMergeS3AdvancedCascade_MandatoryAccessPointVpcOnly verifies that
// accessPointVpcOnly follows the full 8-level cascade.
func TestMergeS3AdvancedCascade_MandatoryAccessPointVpcOnly(t *testing.T) {
	// Only level 4 (localS3AdvCfgMandatory) sets vpcOnly.
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{AccessPointVpcOnly: true},
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if !got.Mandatory.AccessPointVpcOnly {
		t.Fatal("mandatory.accessPointVpcOnly should be true from level-4 S3AdvancedConfig.mandatory")
	}
	if got.Defaults.AccessPointVpcOnly {
		t.Fatal("defaults.accessPointVpcOnly should be false; mandatory input should not bleed through")
	}
}

// TestMergeS3AdvancedCascade_DefaultsCascadeOrder verifies D-6→D-9 priority:
// localS3AdvCfgDefaults (level 6) wins over globalS3AdvCfgDefaults (level 7)
// which wins over localKropathDefaults (level 8) which wins over
// globalKropathDefaults (level 9) for accessPointVpcOnly.
func TestMergeS3AdvancedCascade_DefaultsCascadeOrder(t *testing.T) {
	// Only level 8 (localKropathDefaults) sets vpcOnly — levels 6/7 are zero.
	gotL8 := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedKropathSection{AccessPointVpcOnly: true},
		zeroS3AdvKropath,
	)

	if !gotL8.Defaults.AccessPointVpcOnly {
		t.Fatal("defaults.accessPointVpcOnly should be true from level-8 localKropathDefaults")
	}

	// Level 6 (localS3AdvCfgDefaults) should beat level 8.
	gotL6 := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{AccessPointVpcOnly: true},
		zeroS3AdvCfg,
		cascade.S3AdvancedKropathSection{AccessPointVpcOnly: false},
		zeroS3AdvKropath,
	)
	if !gotL6.Defaults.AccessPointVpcOnly {
		t.Fatal("defaults.accessPointVpcOnly should be true from level-6; level-8 false should not override")
	}
}

// TestMergeS3AdvancedCascade_CommonFields covers AC-7:
// defaults.namingTemplate passes through and mandatory.tags uses additive merge.
func TestMergeS3AdvancedCascade_CommonFields(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		cascade.S3AdvancedConfigSection{
			Tags: map[string]string{"team": "platform"},
		},
		zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{
			NamingTemplate: "{namespace}-{name}",
		},
		zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Mandatory.Tags["team"] != "platform" {
		t.Fatalf("mandatory.tags[team] = %q, want %q", got.Mandatory.Tags["team"], "platform")
	}
	if got.Defaults.NamingTemplate != "{namespace}-{name}" {
		t.Fatalf("defaults.namingTemplate = %q, want %q", got.Defaults.NamingTemplate, "{namespace}-{name}")
	}
}

// TestMergeS3AdvancedCascade_PCIProfile covers AC-11: a PCI-style profile
// with multiple mandatory fields set simultaneously.
func TestMergeS3AdvancedCascade_PCIProfile(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		cascade.S3AdvancedConfigSection{
			TableBucketEncryption: cascade.TableBucketEncryptionSection{
				SseAlgorithm: "aws:kms",
				KmsKeyARN:    "arn:aws:kms:us-east-1:123:key/pci",
			},
			VectorBucketEncryption: cascade.VectorBucketEncryptionSection{
				SseType:   "aws:kms",
				KmsKeyARN: "arn:aws:kms:us-east-1:123:key/pci",
			},
			VectorIndexEncryption: cascade.VectorIndexEncryptionSection{
				SseType:   "aws:kms",
				KmsKeyARN: "arn:aws:kms:us-east-1:123:key/pci",
			},
			AccessPointVpcOnly: true,
			AccessPointPublicAccessBlock: cascade.AccessPointPublicAccessBlockSection{
				BlockPublicACLs:       true,
				BlockPublicPolicy:     true,
				IgnorePublicACLs:      true,
				RestrictPublicBuckets: true,
			},
		},
		zeroS3AdvCfg,
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Mandatory.TableBucketEncryption.SseAlgorithm != "aws:kms" {
		t.Fatal("PCI profile: mandatory.tableBucketEncryption.sseAlgorithm should be aws:kms")
	}
	if got.Mandatory.VectorBucketEncryption.SseType != "aws:kms" {
		t.Fatal("PCI profile: mandatory.vectorBucketEncryption.sseType should be aws:kms")
	}
	if got.Mandatory.VectorIndexEncryption.SseType != "aws:kms" {
		t.Fatal("PCI profile: mandatory.vectorIndexEncryption.sseType should be aws:kms")
	}
	if !got.Mandatory.AccessPointVpcOnly {
		t.Fatal("PCI profile: mandatory.accessPointVpcOnly should be true")
	}
	block := got.Mandatory.AccessPointPublicAccessBlock
	if !block.BlockPublicACLs || !block.BlockPublicPolicy || !block.IgnorePublicACLs || !block.RestrictPublicBuckets {
		t.Fatal("PCI profile: all four mandatory.accessPointPublicAccessBlock booleans should be true")
	}
}

// TestMergeS3AdvancedCascade_TagsAdditiveUnionMandatory verifies that tags from
// multiple mandatory levels are additively union-merged.
func TestMergeS3AdvancedCascade_TagsAdditiveUnionMandatory(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		cascade.S3AdvancedConfigSection{Tags: map[string]string{"env": "prod"}},
		cascade.S3AdvancedConfigSection{Tags: map[string]string{"team": "platform"}},
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Mandatory.Tags["env"] != "prod" {
		t.Fatalf("mandatory.tags[env] = %q, want %q", got.Mandatory.Tags["env"], "prod")
	}
	if got.Mandatory.Tags["team"] != "platform" {
		t.Fatalf("mandatory.tags[team] = %q, want %q", got.Mandatory.Tags["team"], "platform")
	}
}

// TestMergeS3AdvancedCascade_TagsAdditiveUnionDefaults verifies that tags from
// multiple defaults levels are additively union-merged.
func TestMergeS3AdvancedCascade_TagsAdditiveUnionDefaults(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{Tags: map[string]string{"env": "dev"}},
		cascade.S3AdvancedConfigSection{Tags: map[string]string{"owner": "ops"}},
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Defaults.Tags["env"] != "dev" {
		t.Fatalf("defaults.tags[env] = %q, want %q", got.Defaults.Tags["env"], "dev")
	}
	if got.Defaults.Tags["owner"] != "ops" {
		t.Fatalf("defaults.tags[owner] = %q, want %q", got.Defaults.Tags["owner"], "ops")
	}
}

// TestMergeS3AdvancedCascade_EncryptionNotLeakingAcrossTiers verifies that
// mandatory encryption values do not appear in the defaults tier.
func TestMergeS3AdvancedCascade_EncryptionNotLeakingAcrossTiers(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		cascade.S3AdvancedConfigSection{
			TableBucketEncryption: cascade.TableBucketEncryptionSection{SseAlgorithm: "aws:kms"},
		},
		zeroS3AdvCfg,
		zeroS3AdvCfg, zeroS3AdvCfg,
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Defaults.TableBucketEncryption.SseAlgorithm != "" {
		t.Fatalf("defaults.tableBucketEncryption.sseAlgorithm = %q; mandatory value must not bleed into defaults tier", got.Defaults.TableBucketEncryption.SseAlgorithm)
	}
}

// TestMergeS3AdvancedCascade_NamingTemplateS3AdvCfgOnly verifies that
// namingTemplate is governed only at S3AdvancedConfig levels and that
// local (level 6) wins over global (level 7) in the defaults tier.
func TestMergeS3AdvancedCascade_NamingTemplateS3AdvCfgOnly(t *testing.T) {
	got := mergeS3Advanced(
		zeroS3AdvKropath, zeroS3AdvKropath,
		zeroS3AdvCfg, zeroS3AdvCfg,
		cascade.S3AdvancedConfigSection{NamingTemplate: "local-{namespace}-{name}"},
		cascade.S3AdvancedConfigSection{NamingTemplate: "global-{namespace}-{name}"},
		zeroS3AdvKropath, zeroS3AdvKropath,
	)

	if got.Defaults.NamingTemplate != "local-{namespace}-{name}" {
		t.Fatalf("defaults.namingTemplate = %q, want local wins over global", got.Defaults.NamingTemplate)
	}
}
