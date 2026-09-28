// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package monitoring

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	stdlabels "github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"

	"github.com/kropath/kropath-controller/internal/metrics"
	"github.com/kropath/kropath-controller/internal/reconciler/kropathconfigstatus"
	"github.com/kropath/kropath-controller/internal/reconciler/policydocument"
	"github.com/kropath/kropath-controller/internal/reconciler/util"
)

// exemptLabels is design §10.2.1's one documented exemption: `package` on
// kropath_registry_* is bounded but code-derived from features.All and grows
// with every new family, so pinning it here would fail CI on every new one.
// The bidirectional metric-name check (design §10.2) already covers the
// metric itself.
var exemptLabels = map[string]bool{"package": true}

// regexMetachars is the set of characters that make a regex alternate
// something other than a plain literal. A `=~`/`!~` matcher whose value
// contains one of these outside of the `|` alternation separator cannot be
// safely decomposed into a closed set of literals.
const regexMetachars = `.*+?()[]{}^$\`

// metricNameLabel is the reserved label holding the metric name on a vector
// selector (e.g. `foo{bar="baz"}` matches `__name__="foo",bar="baz"`). Spelled
// as a local constant rather than importing stdlabels.MetricName, which is
// deprecated in this version of the module.
const metricNameLabel = "__name__"

// ClosedLabelValues returns the union of every reason/outcome/trigger/field/
// kind closed set the §10.2.1 label-value check validates a rule or dashboard
// panel's label matcher literals against. It imports the exported closed-set
// accessors from every reconciler package that owns a reason const (spec
// §2.1) plus internal/metrics' own outcome/trigger/field/kind sets.
func ClosedLabelValues() map[string]bool {
	set := make(map[string]bool)
	add := func(values []string) {
		for _, v := range values {
			set[v] = true
		}
	}

	add(util.PlacementReasons())
	add(util.ConfigProfileReasons())
	add(kropathconfigstatus.TierReasons())
	add(policydocument.ReadyReasons())
	for _, m := range metrics.OutcomeMetrics() {
		add(metrics.OutcomeValues(m))
	}

	return set
}

// Violation is one label matcher literal that resolved to no member of the
// closed-set union, or a regex matcher that could not be decomposed into
// plain-literal alternates at all.
type Violation struct {
	Source string // e.g. "config/monitoring/rules.yaml: alert KropathConfigUnreferenced"
	Metric string
	Label  string
	Value  string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s: metric %s label %s=%q is not a member of any closed label-value set",
		v.Source, v.Metric, v.Label, v.Value)
}

// CheckExpr parses a single PromQL expression and reports every label matcher
// literal that is not a member of closed. Regex matchers (`=~`/`!~`) are
// decomposed into their `|`-separated alternates; a matcher this function
// cannot decompose into plain-literal alternates is itself reported as a
// violation rather than skipped — spec §10.2.1: "fails on a matcher it cannot
// decompose into alternates rather than passing it."
func CheckExpr(source, expr string, closed map[string]bool) ([]Violation, error) {
	e, err := parser.NewParser(parser.Options{}).ParseExpr(expr)
	if err != nil {
		return nil, fmt.Errorf("%s: parse expr %q: %w", source, expr, err)
	}

	var violations []Violation
	parser.Inspect(e, func(node parser.Node, _ []parser.Node) error {
		vs, ok := node.(*parser.VectorSelector)
		if !ok {
			return nil
		}
		for _, m := range vs.LabelMatchers {
			if m.Name == metricNameLabel || exemptLabels[m.Name] {
				continue
			}
			literals, decomposable := decompose(m)
			if !decomposable {
				violations = append(violations, Violation{Source: source, Metric: vs.Name, Label: m.Name, Value: m.Value})
				continue
			}
			for _, lit := range literals {
				if !closed[lit] {
					violations = append(violations, Violation{Source: source, Metric: vs.Name, Label: m.Name, Value: lit})
				}
			}
		}
		return nil
	})
	return violations, nil
}

// decompose returns the literal values a matcher can resolve to. An `=`/`!=`
// matcher resolves to its one value. A `=~`/`!~` matcher resolves to its
// `|`-separated alternates, provided every alternate is a plain literal with
// no other regex metacharacter — anything else is reported ok=false.
func decompose(m *stdlabels.Matcher) (literals []string, ok bool) {
	switch m.Type {
	case stdlabels.MatchEqual, stdlabels.MatchNotEqual:
		return []string{m.Value}, true
	case stdlabels.MatchRegexp, stdlabels.MatchNotRegexp:
		parts := strings.Split(m.Value, "|")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if p == "" || strings.ContainsAny(p, regexMetachars) {
				return nil, false
			}
			out = append(out, p)
		}
		return out, true
	default:
		return nil, false
	}
}

// dashboard is the minimal subset of the Grafana dashboard JSON schema the
// label-value check reads: every panel's PromQL target expressions.
type dashboard struct {
	Panels []struct {
		Title   string `json:"title"`
		Targets []struct {
			Expr string `json:"expr"`
		} `json:"targets"`
	} `json:"panels"`
}

// CheckDashboard parses a dashboard JSON file and reports every label matcher
// literal in a panel target's expr that is not a member of closed.
func CheckDashboard(path string, closed map[string]bool) ([]Violation, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is a repo-relative config file, not user input
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var d dashboard
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	var violations []Violation
	for _, p := range d.Panels {
		for _, t := range p.Targets {
			source := fmt.Sprintf("%s: panel %q", path, p.Title)
			v, err := CheckExpr(source, t.Expr, closed)
			if err != nil {
				return nil, err
			}
			violations = append(violations, v...)
		}
	}
	return violations, nil
}

// CheckRulesFile parses a rules.yaml file and reports every label matcher
// literal in a rule's expr that is not a member of closed.
func CheckRulesFile(path string, closed map[string]bool) ([]Violation, error) {
	rf, err := LoadRulesFile(path)
	if err != nil {
		return nil, err
	}

	var violations []Violation
	for _, g := range rf.Groups {
		for _, r := range g.Rules {
			source := fmt.Sprintf("%s: alert %s", path, r.Alert)
			v, err := CheckExpr(source, r.Expr, closed)
			if err != nil {
				return nil, err
			}
			violations = append(violations, v...)
		}
	}
	return violations, nil
}

// CheckAll runs the full §10.2.1 label-value check against both artifacts and
// returns every violation found across both.
func CheckAll(rulesPath, dashboardPath string) ([]Violation, error) {
	closed := ClosedLabelValues()

	var all []Violation

	ruleViolations, err := CheckRulesFile(rulesPath, closed)
	if err != nil {
		return nil, err
	}
	all = append(all, ruleViolations...)

	dashViolations, err := CheckDashboard(dashboardPath, closed)
	if err != nil {
		return nil, err
	}
	all = append(all, dashViolations...)

	return all, nil
}
