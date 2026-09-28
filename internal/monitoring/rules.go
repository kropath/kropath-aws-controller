// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

// Package monitoring holds the shared types for config/monitoring/rules.yaml —
// the canonical Prometheus rule groups (spec §4, design §10). cmd/gen-monitoring
// reads a RulesFile to generate config/monitoring/prometheusrule.yaml, and the
// internal/metrics label-value check reads it to validate every label matcher
// literal against the closed value sets (spec §10.2.1).
package monitoring

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RulesFile is the top-level shape of config/monitoring/rules.yaml — plain
// Prometheus rule groups, the form `promtool check rules` validates directly.
type RulesFile struct {
	Groups []RuleGroup `yaml:"groups"`
}

// RuleGroup is one named group of alerting rules.
type RuleGroup struct {
	Name  string `yaml:"name"`
	Rules []Rule `yaml:"rules"`
}

// Rule is a single Prometheus alerting rule. Labels and Annotations are fixed
// structs rather than maps so that re-marshaling is deterministic — a
// map[string]string's key order is not guaranteed stable across `go run`
// invocations, which would make the monitoring-verify drift gate flaky.
type Rule struct {
	Alert       string          `yaml:"alert"`
	Expr        string          `yaml:"expr"`
	For         string          `yaml:"for,omitempty"`
	Labels      RuleLabels      `yaml:"labels"`
	Annotations RuleAnnotations `yaml:"annotations"`
}

// RuleLabels are the two labels every rule in rules.yaml carries (design §10.4).
type RuleLabels struct {
	Severity  string `yaml:"severity"`
	Component string `yaml:"component"`
}

// RuleAnnotations are the operator-facing summary/description on a rule.
type RuleAnnotations struct {
	Summary     string `yaml:"summary"`
	Description string `yaml:"description,omitempty"`
}

// LoadRulesFile reads and parses a rules.yaml file.
func LoadRulesFile(path string) (*RulesFile, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is a repo-relative config file, not user input
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var rf RulesFile
	if err := yaml.Unmarshal(data, &rf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &rf, nil
}
