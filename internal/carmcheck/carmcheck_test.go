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

package carmcheck

import (
	"context"
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/reconciler/util"
)

func TestReportHasBlocking(t *testing.T) {
	r := &Report{Findings: []Finding{
		{Check: "a", Severity: SeverityOK},
		{Check: "b", Severity: SeverityBlocking},
	}}
	if !r.HasBlocking() {
		t.Error("HasBlocking() = false, want true")
	}
	if r.HasUnknown() {
		t.Error("HasUnknown() = true, want false")
	}
}

func TestReportHasUnknown(t *testing.T) {
	r := &Report{Findings: []Finding{
		{Check: "a", Severity: SeverityUnknown},
	}}
	if r.HasBlocking() {
		t.Error("HasBlocking() = true, want false")
	}
	if !r.HasUnknown() {
		t.Error("HasUnknown() = false, want true")
	}
}

func TestReportEmptyHasNoFindings(t *testing.T) {
	r := &Report{}
	if r.HasBlocking() || r.HasUnknown() {
		t.Error("empty report reported a finding")
	}
}

// TestRunEndToEndCleanInstall exercises Run against a fully-compliant
// install: one kropath-managed namespace, an ACK Deployment with CARM
// enabled and no watch-namespace restriction, and a role-account-map entry
// whose ARN account matches its key.
func TestRunEndToEndCleanInstall(t *testing.T) {
	ns := namespace("payments-prod",
		map[string]string{
			util.GlobalConfigNamespaceAnnotation: "payments-prod",
			"services.k8s.aws/owner-account-id":  "111122223333",
		}, nil)
	dep := testDeployment("ack-system", "ack-iam-controller", []string{"--enable-carm=true"})
	cm := testConfigMap("ack-system", "ack-role-account-map", map[string]string{
		"111122223333": "arn:aws:iam::111122223333:role/ACKRole",
	})

	c := testClient(t, ns, dep, cm).Build()
	report, err := Run(context.Background(), c, Options{ACKNamespace: "ack-system"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if report.HasBlocking() {
		t.Errorf("Run() reported a blocking finding on a clean install: %+v", report.Findings)
	}
}

// TestRunEndToEndBrokenInstall exercises every precondition violation at
// once against a single kropath-managed namespace.
func TestRunEndToEndBrokenInstall(t *testing.T) {
	ns := namespace("payments-prod",
		map[string]string{
			util.GlobalConfigNamespaceAnnotation: "payments-prod",
			"services.k8s.aws/owner-account-id":  "111122223333",
			"services.k8s.aws/team-id":           "platform-team",
		}, nil)
	dep := testDeployment("ack-system", "ack-iam-controller", []string{
		"--enable-carm=false",
		"--watch-namespace=other-namespace",
	})
	cm := testConfigMap("ack-system", "ack-role-account-map", map[string]string{
		"111122223333": "arn:aws:iam::999988887777:role/ACKRole",
	})

	c := testClient(t, ns, dep, cm).Build()
	report, err := Run(context.Background(), c, Options{ACKNamespace: "ack-system"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !report.HasBlocking() {
		t.Fatalf("Run() reported no blocking finding on a broken install: %+v", report.Findings)
	}

	wantChecks := map[string]bool{
		"enable-carm":           false,
		"watch-namespace-scope": false,
		"role-account-map":      false,
		"team-id":               false,
	}
	for _, f := range report.Findings {
		if f.Severity == SeverityBlocking {
			if _, ok := wantChecks[f.Check]; ok {
				wantChecks[f.Check] = true
			}
		}
	}
	for check, found := range wantChecks {
		if !found {
			t.Errorf("expected a blocking finding for check %q, got none. All findings: %+v", check, report.Findings)
		}
	}
}

func TestRunNamespaceOverrideSkipsAutoDiscovery(t *testing.T) {
	// This namespace does NOT carry GlobalConfigNamespaceAnnotation, so
	// auto-discovery would skip it — but an explicit override must still
	// pick it up.
	ns := namespace("explicit-ns", map[string]string{
		"services.k8s.aws/team-id": "some-team",
	}, nil)
	dep := testDeployment("ack-system", "ack-iam-controller", []string{"--enable-carm=true"})

	c := testClient(t, ns, dep).Build()
	report, err := Run(context.Background(), c, Options{
		ACKNamespace: "ack-system",
		Namespaces:   []string{"explicit-ns"},
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	found := false
	for _, f := range report.Findings {
		if f.Check == "team-id" && f.Namespace == "explicit-ns" && f.Severity == SeverityBlocking {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a blocking team-id finding for explicitly-overridden namespace, got: %+v", report.Findings)
	}
}

func TestRunNamespaceOverrideMissingNamespaceErrors(t *testing.T) {
	c := testClient(t).Build()
	_, err := Run(context.Background(), c, Options{
		ACKNamespace: "ack-system",
		Namespaces:   []string{"does-not-exist"},
	})
	if err == nil {
		t.Fatal("Run() error = nil, want error for nonexistent namespace override")
	}
}

func TestRunNoManagedNamespaces(t *testing.T) {
	// No namespace carries the annotation, so there is nothing
	// namespace-scoped to check, but the install-wide checks still run.
	dep := testDeployment("ack-system", "ack-iam-controller", []string{"--enable-carm=true"})
	c := testClient(t, dep).Build()

	report, err := Run(context.Background(), c, Options{ACKNamespace: "ack-system"})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if report.HasBlocking() {
		t.Errorf("Run() with no managed namespaces reported a blocking finding: %+v", report.Findings)
	}
	sawEnableCARM := false
	for _, f := range report.Findings {
		if f.Check == "enable-carm" {
			sawEnableCARM = true
		}
	}
	if !sawEnableCARM {
		t.Error("expected an enable-carm finding even with zero managed namespaces")
	}
}
