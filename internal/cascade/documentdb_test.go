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

// emptyDocDBKropath returns a zero DocumentDBKropathSection.
func emptyDocDBKropath() cascade.DocumentDBKropathSection {
	return cascade.DocumentDBKropathSection{}
}

// emptyDocDBCfg returns a zero DocumentDBConfigSection.
func emptyDocDBCfg() cascade.DocumentDBConfigSection {
	return cascade.DocumentDBConfigSection{}
}

func TestMergeDocumentDBCascade_GlobalKropathMandatoryWins(t *testing.T) {
	globalKropathMandatory := cascade.DocumentDBKropathSection{
		StorageEncrypted:       boolPtr(true),
		DeletionProtection:     boolPtr(true),
		AllowedInstanceClasses: []string{"db.r6g.large"},
		BackupRetentionPeriod:  7,
	}
	// All lower-priority sources say "false" or provide competing values.
	localKropathMandatory := cascade.DocumentDBKropathSection{
		StorageEncrypted: boolPtr(false),
	}
	globalCfgMandatory := cascade.DocumentDBConfigSection{
		StorageEncrypted:       boolPtr(false),
		AllowedInstanceClasses: []string{"db.t3.medium"},
		BackupRetentionPeriod:  3,
	}

	result := cascade.MergeDocumentDBCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalCfgMandatory,
		emptyDocDBCfg(), // localCfgMandatory
		emptyDocDBCfg(), // localCfgDefaults
		emptyDocDBCfg(), // globalCfgDefaults
		emptyDocDBKropath(),
		emptyDocDBKropath(),
	)

	if result.Mandatory.StorageEncrypted == nil || !*result.Mandatory.StorageEncrypted {
		t.Errorf("expected StorageEncrypted=true from globalKropathMandatory, got %v", result.Mandatory.StorageEncrypted)
	}
	if result.Mandatory.DeletionProtection == nil || !*result.Mandatory.DeletionProtection {
		t.Errorf("expected DeletionProtection=true from globalKropathMandatory, got %v", result.Mandatory.DeletionProtection)
	}
	if len(result.Mandatory.AllowedInstanceClasses) != 1 || result.Mandatory.AllowedInstanceClasses[0] != "db.r6g.large" {
		t.Errorf("expected AllowedInstanceClasses=[db.r6g.large] from globalKropathMandatory, got %v", result.Mandatory.AllowedInstanceClasses)
	}
	if result.Mandatory.BackupRetentionPeriod != 7 {
		t.Errorf("expected BackupRetentionPeriod=7 from globalKropathMandatory, got %d", result.Mandatory.BackupRetentionPeriod)
	}
}

func TestMergeDocumentDBCascade_LocalCfgMandatoryFillsKropathGap(t *testing.T) {
	// globalKropathMandatory does NOT enforce kmsKeyArn or dbInstanceClass.
	// globalCfgMandatory provides kmsKeyArn.
	// localCfgMandatory provides dbInstanceClass.
	globalCfgMandatory := cascade.DocumentDBConfigSection{
		KmsKeyArn: "arn:aws:kms:us-east-1:123456789:key/global",
	}
	localCfgMandatory := cascade.DocumentDBConfigSection{
		DbInstanceClass: "db.r6g.large",
	}

	result := cascade.MergeDocumentDBCascade(
		emptyDocDBKropath(),
		emptyDocDBKropath(),
		globalCfgMandatory,
		localCfgMandatory,
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBKropath(),
		emptyDocDBKropath(),
	)

	if result.Mandatory.KmsKeyArn != "arn:aws:kms:us-east-1:123456789:key/global" {
		t.Errorf("expected KmsKeyArn from globalCfgMandatory, got %q", result.Mandatory.KmsKeyArn)
	}
	if result.Mandatory.DbInstanceClass != "db.r6g.large" {
		t.Errorf("expected DbInstanceClass from localCfgMandatory, got %q", result.Mandatory.DbInstanceClass)
	}
}

func TestMergeDocumentDBCascade_LocalCfgDefaultsWinsOverGlobalKropathDefaults(t *testing.T) {
	// Defaults priority: L6 localCfgDefaults > L7 globalCfgDefaults > L8 localKropathDefaults > L9 globalKropathDefaults.
	localCfgDefaults := cascade.DocumentDBConfigSection{
		EngineVersion: "5.0.0",
	}
	globalCfgDefaults := cascade.DocumentDBConfigSection{
		EngineVersion: "4.0.0",
	}
	localKropathDefaults := cascade.DocumentDBKropathSection{
		BackupRetentionPeriod: 5,
	}
	globalKropathDefaults := cascade.DocumentDBKropathSection{
		BackupRetentionPeriod: 3,
	}

	result := cascade.MergeDocumentDBCascade(
		emptyDocDBKropath(),
		emptyDocDBKropath(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		localCfgDefaults,
		globalCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)

	if result.Defaults.EngineVersion != "5.0.0" {
		t.Errorf("expected EngineVersion=5.0.0 from localCfgDefaults, got %q", result.Defaults.EngineVersion)
	}
	if result.Defaults.BackupRetentionPeriod != 5 {
		t.Errorf("expected BackupRetentionPeriod=5 from localKropathDefaults, got %d", result.Defaults.BackupRetentionPeriod)
	}
}

func TestMergeDocumentDBCascade_GlobalKropathDefaultsAsFallback(t *testing.T) {
	// globalKropathDefaults is the weakest defaults source; should still show up when nothing else provides a value.
	globalKropathDefaults := cascade.DocumentDBKropathSection{
		StorageEncrypted:      boolPtr(true),
		BackupRetentionPeriod: 7,
	}

	result := cascade.MergeDocumentDBCascade(
		emptyDocDBKropath(),
		emptyDocDBKropath(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBKropath(),
		globalKropathDefaults,
	)

	if result.Defaults.StorageEncrypted == nil || !*result.Defaults.StorageEncrypted {
		t.Errorf("expected StorageEncrypted=true from globalKropathDefaults, got %v", result.Defaults.StorageEncrypted)
	}
	if result.Defaults.BackupRetentionPeriod != 7 {
		t.Errorf("expected BackupRetentionPeriod=7 from globalKropathDefaults, got %d", result.Defaults.BackupRetentionPeriod)
	}
}

func TestMergeDocumentDBCascade_TagsAdditiveUnionMandatory(t *testing.T) {
	// Tags are additive union: all four mandatory sources contribute.
	// Key conflict: higher-priority source wins (globalKropathMandatory > localKropathMandatory > globalCfgMandatory > localCfgMandatory).
	globalKropathMandatory := cascade.DocumentDBKropathSection{
		Tags: map[string]string{"env": "prod", "owner": "platform"},
	}
	localKropathMandatory := cascade.DocumentDBKropathSection{
		Tags: map[string]string{"env": "staging", "team": "backend"},
	}
	globalCfgMandatory := cascade.DocumentDBConfigSection{
		Tags: map[string]string{"env": "dev", "service": "api"},
	}
	localCfgMandatory := cascade.DocumentDBConfigSection{
		Tags: map[string]string{"env": "local", "app": "myapp"},
	}

	result := cascade.MergeDocumentDBCascade(
		globalKropathMandatory,
		localKropathMandatory,
		globalCfgMandatory,
		localCfgMandatory,
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBKropath(),
		emptyDocDBKropath(),
	)

	// "env" should be "prod" from globalKropathMandatory (highest priority).
	if result.Mandatory.Tags["env"] != "prod" {
		t.Errorf("expected Tags[env]=prod, got %q", result.Mandatory.Tags["env"])
	}
	// "team" from localKropathMandatory.
	if result.Mandatory.Tags["team"] != "backend" {
		t.Errorf("expected Tags[team]=backend, got %q", result.Mandatory.Tags["team"])
	}
	// "service" from globalCfgMandatory.
	if result.Mandatory.Tags["service"] != "api" {
		t.Errorf("expected Tags[service]=api, got %q", result.Mandatory.Tags["service"])
	}
	// "app" from localCfgMandatory.
	if result.Mandatory.Tags["app"] != "myapp" {
		t.Errorf("expected Tags[app]=myapp, got %q", result.Mandatory.Tags["app"])
	}
	// "owner" from globalKropathMandatory.
	if result.Mandatory.Tags["owner"] != "platform" {
		t.Errorf("expected Tags[owner]=platform, got %q", result.Mandatory.Tags["owner"])
	}
}

func TestMergeDocumentDBCascade_TagsAdditiveUnionDefaults(t *testing.T) {
	// In defaults tier, localCfgDefaults wins on conflict.
	localCfgDefaults := cascade.DocumentDBConfigSection{
		Tags: map[string]string{"env": "local", "app": "myapp"},
	}
	globalCfgDefaults := cascade.DocumentDBConfigSection{
		Tags: map[string]string{"env": "dev", "service": "api"},
	}
	localKropathDefaults := cascade.DocumentDBKropathSection{
		Tags: map[string]string{"env": "staging", "team": "backend"},
	}
	globalKropathDefaults := cascade.DocumentDBKropathSection{
		Tags: map[string]string{"env": "prod", "owner": "platform"},
	}

	result := cascade.MergeDocumentDBCascade(
		emptyDocDBKropath(),
		emptyDocDBKropath(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		localCfgDefaults,
		globalCfgDefaults,
		localKropathDefaults,
		globalKropathDefaults,
	)

	// "env" should be "local" from localCfgDefaults (highest priority in defaults).
	if result.Defaults.Tags["env"] != "local" {
		t.Errorf("expected Tags[env]=local from localCfgDefaults, got %q", result.Defaults.Tags["env"])
	}
	if result.Defaults.Tags["app"] != "myapp" {
		t.Errorf("expected Tags[app]=myapp, got %q", result.Defaults.Tags["app"])
	}
	if result.Defaults.Tags["service"] != "api" {
		t.Errorf("expected Tags[service]=api from globalCfgDefaults, got %q", result.Defaults.Tags["service"])
	}
	if result.Defaults.Tags["team"] != "backend" {
		t.Errorf("expected Tags[team]=backend from localKropathDefaults, got %q", result.Defaults.Tags["team"])
	}
	if result.Defaults.Tags["owner"] != "platform" {
		t.Errorf("expected Tags[owner]=platform from globalKropathDefaults, got %q", result.Defaults.Tags["owner"])
	}
}

func TestMergeDocumentDBCascade_SyncedLabelsFromDocDBConfigOnly(t *testing.T) {
	// SyncedLabels/Annotations come only from DocumentDBConfig levels (not KropathConfig).
	globalCfgMandatory := cascade.DocumentDBConfigSection{
		SyncedLabels:      map[string]string{"global-label": "value"},
		SyncedAnnotations: map[string]string{"global-annotation": "note"},
	}
	localCfgMandatory := cascade.DocumentDBConfigSection{
		SyncedLabels: map[string]string{"local-label": "value", "global-label": "overridden"},
	}

	result := cascade.MergeDocumentDBCascade(
		emptyDocDBKropath(),
		emptyDocDBKropath(),
		globalCfgMandatory,
		localCfgMandatory,
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBKropath(),
		emptyDocDBKropath(),
	)

	// globalCfgMandatory wins on conflict ("global-label" has both but L3 > L4).
	if result.Mandatory.SyncedLabels["global-label"] != "value" {
		t.Errorf("expected SyncedLabels[global-label]=value from globalCfgMandatory, got %q", result.Mandatory.SyncedLabels["global-label"])
	}
	if result.Mandatory.SyncedLabels["local-label"] != "value" {
		t.Errorf("expected SyncedLabels[local-label]=value from localCfgMandatory, got %q", result.Mandatory.SyncedLabels["local-label"])
	}
	if result.Mandatory.SyncedAnnotations["global-annotation"] != "note" {
		t.Errorf("expected SyncedAnnotations[global-annotation]=note, got %q", result.Mandatory.SyncedAnnotations["global-annotation"])
	}
}

func TestMergeDocumentDBCascade_AllSourcesEmpty(t *testing.T) {
	result := cascade.MergeDocumentDBCascade(
		emptyDocDBKropath(),
		emptyDocDBKropath(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBKropath(),
		emptyDocDBKropath(),
	)

	if result.Mandatory.StorageEncrypted != nil {
		t.Errorf("expected nil StorageEncrypted, got %v", result.Mandatory.StorageEncrypted)
	}
	if result.Mandatory.DbInstanceClass != "" {
		t.Errorf("expected empty DbInstanceClass, got %q", result.Mandatory.DbInstanceClass)
	}
	if result.Mandatory.BackupRetentionPeriod != 0 {
		t.Errorf("expected 0 BackupRetentionPeriod, got %d", result.Mandatory.BackupRetentionPeriod)
	}
	if len(result.Mandatory.Tags) != 0 {
		t.Errorf("expected empty Tags, got %v", result.Mandatory.Tags)
	}
}

func TestMergeDocumentDBCascade_EnableCloudwatchLogsExportsMandatory(t *testing.T) {
	// L3 (globalCfgMandatory) wins over L4 (localCfgMandatory) for []string fields.
	globalCfgMandatory := cascade.DocumentDBConfigSection{
		EnableCloudwatchLogsExports: []string{"audit", "profiler"},
	}
	localCfgMandatory := cascade.DocumentDBConfigSection{
		EnableCloudwatchLogsExports: []string{"audit"},
	}

	result := cascade.MergeDocumentDBCascade(
		emptyDocDBKropath(),
		emptyDocDBKropath(),
		globalCfgMandatory,
		localCfgMandatory,
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBKropath(),
		emptyDocDBKropath(),
	)

	if len(result.Mandatory.EnableCloudwatchLogsExports) != 2 {
		t.Errorf("expected 2 EnableCloudwatchLogsExports from globalCfgMandatory, got %v", result.Mandatory.EnableCloudwatchLogsExports)
	}
}

// --- ValidateDocumentDBInstanceClass tests ---

func TestValidateDocumentDBInstanceClass_NoConstraint(t *testing.T) {
	// Both fields empty — no constraint, always valid.
	mandatory := cascade.EffectiveDocumentDBSection{}
	valid, msg := cascade.ValidateDocumentDBInstanceClass(mandatory)
	if !valid || msg != "" {
		t.Errorf("expected valid=true, msg=''; got valid=%v, msg=%q", valid, msg)
	}
}

func TestValidateDocumentDBInstanceClass_NoAllowedList(t *testing.T) {
	// dbInstanceClass set but no allowedInstanceClasses — constraint does not apply.
	mandatory := cascade.EffectiveDocumentDBSection{
		DbInstanceClass: "db.r6g.large",
	}
	valid, msg := cascade.ValidateDocumentDBInstanceClass(mandatory)
	if !valid || msg != "" {
		t.Errorf("expected valid=true, msg=''; got valid=%v, msg=%q", valid, msg)
	}
}

func TestValidateDocumentDBInstanceClass_NoDbInstanceClass(t *testing.T) {
	// allowedInstanceClasses set but dbInstanceClass empty — constraint does not apply.
	mandatory := cascade.EffectiveDocumentDBSection{
		AllowedInstanceClasses: []string{"db.r6g.large", "db.r6g.xlarge"},
	}
	valid, msg := cascade.ValidateDocumentDBInstanceClass(mandatory)
	if !valid || msg != "" {
		t.Errorf("expected valid=true, msg=''; got valid=%v, msg=%q", valid, msg)
	}
}

func TestValidateDocumentDBInstanceClass_ValidMember(t *testing.T) {
	// dbInstanceClass is in allowedInstanceClasses — valid.
	mandatory := cascade.EffectiveDocumentDBSection{
		DbInstanceClass:        "db.r6g.large",
		AllowedInstanceClasses: []string{"db.r6g.large", "db.r6g.xlarge"},
	}
	valid, msg := cascade.ValidateDocumentDBInstanceClass(mandatory)
	if !valid || msg != "" {
		t.Errorf("expected valid=true, msg=''; got valid=%v, msg=%q", valid, msg)
	}
}

func TestValidateDocumentDBInstanceClass_InvalidNotInList(t *testing.T) {
	// dbInstanceClass not in allowedInstanceClasses — invalid.
	mandatory := cascade.EffectiveDocumentDBSection{
		DbInstanceClass:        "db.t3.medium",
		AllowedInstanceClasses: []string{"db.r6g.large", "db.r6g.xlarge"},
	}
	valid, msg := cascade.ValidateDocumentDBInstanceClass(mandatory)
	if valid {
		t.Error("expected valid=false, got true")
	}
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestValidateDocumentDBInstanceClass_SingleElementListMatch(t *testing.T) {
	// Exact match against a single-element list.
	mandatory := cascade.EffectiveDocumentDBSection{
		DbInstanceClass:        "db.r6g.xlarge",
		AllowedInstanceClasses: []string{"db.r6g.xlarge"},
	}
	valid, msg := cascade.ValidateDocumentDBInstanceClass(mandatory)
	if !valid || msg != "" {
		t.Errorf("expected valid=true, msg=''; got valid=%v, msg=%q", valid, msg)
	}
}

func TestValidateDocumentDBInstanceClass_SingleElementListNoMatch(t *testing.T) {
	// No match against a single-element list.
	mandatory := cascade.EffectiveDocumentDBSection{
		DbInstanceClass:        "db.r6g.large",
		AllowedInstanceClasses: []string{"db.r6g.xlarge"},
	}
	valid, msg := cascade.ValidateDocumentDBInstanceClass(mandatory)
	if valid {
		t.Errorf("expected valid=false, got true; msg=%q", msg)
	}
}

func TestMergeDocumentDBCascade_AllowedInstanceClassesKropathVsCfg(t *testing.T) {
	// KropathConfig level 1 wins over DocumentDBConfig level 3 for AllowedInstanceClasses.
	globalKropathMandatory := cascade.DocumentDBKropathSection{
		AllowedInstanceClasses: []string{"db.r6g.large"},
	}
	globalCfgMandatory := cascade.DocumentDBConfigSection{
		AllowedInstanceClasses: []string{"db.t3.medium"},
	}

	result := cascade.MergeDocumentDBCascade(
		globalKropathMandatory,
		emptyDocDBKropath(),
		globalCfgMandatory,
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBKropath(),
		emptyDocDBKropath(),
	)

	if len(result.Mandatory.AllowedInstanceClasses) != 1 || result.Mandatory.AllowedInstanceClasses[0] != "db.r6g.large" {
		t.Errorf("expected AllowedInstanceClasses=[db.r6g.large] from globalKropathMandatory, got %v", result.Mandatory.AllowedInstanceClasses)
	}
}

func TestMergeDocumentDBCascade_BackupRetentionPeriodKropathDefaults(t *testing.T) {
	// Defaults: L8 localKropathDefaults > L9 globalKropathDefaults for BackupRetentionPeriod.
	localKropathDefaults := cascade.DocumentDBKropathSection{
		BackupRetentionPeriod: 14,
	}
	globalKropathDefaults := cascade.DocumentDBKropathSection{
		BackupRetentionPeriod: 7,
	}

	result := cascade.MergeDocumentDBCascade(
		emptyDocDBKropath(),
		emptyDocDBKropath(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		emptyDocDBCfg(),
		localKropathDefaults,
		globalKropathDefaults,
	)

	if result.Defaults.BackupRetentionPeriod != 14 {
		t.Errorf("expected BackupRetentionPeriod=14 from localKropathDefaults, got %d", result.Defaults.BackupRetentionPeriod)
	}
}
