// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	"context"
	"testing"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TestReadyReasonsCoversEveryReadyConditionReason enumerates every literal
// argument passed to readyCondition(...) in reconciler.go and asserts
// ReadyReasons() contains exactly that set (spec §2.4.3, §10.2.1). Asserted
// by enumeration against the source of truth, not a hardcoded count, so a
// future eighth reason fails this test instead of silently failing the
// §10.2.1 label-value check.
func TestReadyReasonsCoversEveryReadyConditionReason(t *testing.T) {
	want := map[string]bool{
		ReasonDocumentResolved:         true, // reconcileDocument: raw passthrough + structured serialization
		ReasonInvalidDocumentJSON:      true, // reconcileDocument: spec.documentJSON invalid
		ReasonSidConflict:              true, // reconcileDocument: two merged statements share a Sid
		ReasonSourceNotReady:           true, // reconcileDocument: one or more statement refs unresolved
		ReasonSourceMissing:            true, // collectSourceStatements: source document not found
		ReasonSourcePending:            true, // collectSourceStatements: source document not yet resolved
		ReasonMergeFromRawNotSupported: true, // collectSourceStatements: source document uses spec.documentJSON
	}

	got := ReadyReasons()
	if len(got) != len(want) {
		t.Fatalf("ReadyReasons() has %d entries, want %d: %v", len(got), len(want), got)
	}
	for _, reason := range got {
		if !want[reason] {
			t.Errorf("ReadyReasons() contains unexpected reason %q", reason)
		}
		delete(want, reason)
	}
	for missing := range want {
		t.Errorf("ReadyReasons() is missing reason %q", missing)
	}
}

func resourceRefStatement(refs ...*v1alpha1.PolicyRef) v1alpha1.PolicyStatement {
	resources := make([]v1alpha1.PolicyResource, len(refs))
	for i, ref := range refs {
		resources[i] = v1alpha1.PolicyResource{Ref: ref}
	}
	return v1alpha1.PolicyStatement{Effect: "Allow", Actions: []string{"s3:GetObject"}, Resources: resources}
}

// recordingPendingResolve returns a resolve function that records every ref
// name it is called with and always reports it pending, so a test can assert
// whether every ref in a statement was attempted (M-10) rather than only the
// first.
func recordingPendingResolve(attempted *[]string) func(context.Context, client.Client, string, *v1alpha1.PolicyRef) (string, bool, error) {
	return func(_ context.Context, _ client.Client, _ string, ref *v1alpha1.PolicyRef) (string, bool, error) {
		*attempted = append(*attempted, ref.Name)
		return "", true, nil
	}
}

// TestResolveResourcesCollectsAllPending asserts resolveResources attempts
// every resource ref rather than stopping at the first pending one, and
// reports the total pending count (M-10, spec §2.4.2).
func TestResolveResourcesCollectsAllPending(t *testing.T) {
	var attempted []string
	resources := []v1alpha1.PolicyResource{
		{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "a"}},
		{Ref: &v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "b"}},
		{Ref: &v1alpha1.PolicyRef{Kind: "AWSLambdaFunction", Name: "c"}},
	}

	_, pending, err := resolveResources(context.Background(), nil, "default", resources, recordingPendingResolve(&attempted))
	if err != nil {
		t.Fatalf("resolveResources: %v", err)
	}
	if pending != 3 {
		t.Errorf("resolveResources pending = %d, want 3", pending)
	}
	if len(attempted) != 3 {
		t.Fatalf("resolve attempted %d refs, want 3 (all of them): %v", len(attempted), attempted)
	}
}

// TestResolvePrincipalsCollectsAllPending is resolveResources's counterpart
// for principals (M-10, spec §2.4.2).
func TestResolvePrincipalsCollectsAllPending(t *testing.T) {
	var attempted []string
	principals := []v1alpha1.PolicyPrincipal{
		{Type: "AWS", Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "a"}},
		{Type: "AWS", Ref: &v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "b"}},
		{Type: "AWS", Ref: &v1alpha1.PolicyRef{Kind: "AWSLambdaFunction", Name: "c"}},
	}

	_, pending, err := resolvePrincipals(context.Background(), nil, "default", principals, recordingPendingResolve(&attempted))
	if err != nil {
		t.Fatalf("resolvePrincipals: %v", err)
	}
	if pending != 3 {
		t.Errorf("resolvePrincipals pending = %d, want 3", pending)
	}
	if len(attempted) != 3 {
		t.Fatalf("resolve attempted %d refs, want 3 (all of them): %v", len(attempted), attempted)
	}
}

// TestResolveStatementReportsThreeNotOne is the spec §8 required case:
// resolveStatement with three unresolvable refs reports 3, not 1.
func TestResolveStatementReportsThreeNotOne(t *testing.T) {
	var attempted []string
	stmt := resourceRefStatement(
		&v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "a"},
		&v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "b"},
		&v1alpha1.PolicyRef{Kind: "AWSLambdaFunction", Name: "c"},
	)

	_, pending, err := resolveStatement(context.Background(), nil, "default", stmt, recordingPendingResolve(&attempted))
	if err != nil {
		t.Fatalf("resolveStatement: %v", err)
	}
	if pending != 3 {
		t.Errorf("resolveStatement pending = %d, want 3", pending)
	}
	if len(attempted) != 3 {
		t.Fatalf("resolve attempted %d refs, want 3 (all of them): %v", len(attempted), attempted)
	}
}

// TestReconcileDocumentCollectsPendingAcrossAllStatements is the
// reconcileDocument-level case of M-10 (spec §2.4.2): three statements, each
// with one unresolvable ref, must each get a resolve attempt -- the outer
// loop over doc.Spec.Statements previously returned on the first pending
// statement and never called resolveStatement on the rest.
func TestReconcileDocumentCollectsPendingAcrossAllStatements(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.PolicyDocument{
		ObjectMeta: metav1.ObjectMeta{Name: "ac09-doc", Namespace: "default", Generation: 1},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				resourceRefStatement(&v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "a"}),
				resourceRefStatement(&v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "b"}),
				resourceRefStatement(&v1alpha1.PolicyRef{Kind: "AWSLambdaFunction", Name: "c"}),
			},
		},
	}

	var attempted []string
	r := &Reconciler{
		Client:       fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build(),
		ResolveRefFn: recordingPendingResolve(&attempted),
	}

	result, err := r.reconcileDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("reconcileDocument: %v", err)
	}
	if result.RequeueAfter <= 0 {
		t.Errorf("expected a positive RequeueAfter for a pending document, got %v", result.RequeueAfter)
	}
	if len(attempted) != 3 {
		t.Fatalf("resolve attempted %d refs across all statements, want 3 (all of them): %v", len(attempted), attempted)
	}
	if doc.Status.ResolvedDocumentJSON != "" {
		t.Errorf("expected empty status.resolvedDocumentJSON while refs are pending, got %q", doc.Status.ResolvedDocumentJSON)
	}

	var readyCond *metav1.Condition
	for i := range doc.Status.Conditions {
		if doc.Status.Conditions[i].Type == v1alpha1.ConditionReady {
			readyCond = &doc.Status.Conditions[i]
		}
	}
	if readyCond == nil || readyCond.Reason != ReasonSourceNotReady {
		t.Errorf("expected Ready condition reason %q, got %+v", ReasonSourceNotReady, readyCond)
	}
}
