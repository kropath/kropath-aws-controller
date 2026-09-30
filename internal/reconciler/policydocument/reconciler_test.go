// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package policydocument

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/kropath/kropath-aws-controller/api/v1alpha1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type recordingClient struct {
	client.Client
	statusUpdates int
}

func (c *recordingClient) Status() client.StatusWriter {
	return &recordingStatusWriter{
		StatusWriter: c.Client.Status(),
		onUpdate: func() {
			c.statusUpdates++
		},
	}
}

type recordingStatusWriter struct {
	client.StatusWriter
	onUpdate func()
}

func (w *recordingStatusWriter) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error {
	w.onUpdate()
	return w.StatusWriter.Update(ctx, obj, opts...)
}

func TestReconcilePassesThroughDocumentJSON(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "raw", Namespace: "default", Generation: 3},
		Spec: v1alpha1.PolicyDocumentSpec{
			DocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if doc.Status.ResolvedDocumentJSON != doc.Spec.DocumentJSON {
		t.Fatalf("resolved json mismatch: got %s", doc.Status.ResolvedDocumentJSON)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("ready condition not true: %#v", doc.Status.Conditions)
	}
}

func TestReconcileClearsResolvedDocumentOnInvalidDocumentJSON(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "raw", Namespace: "default", Generation: 4},
		Spec: v1alpha1.PolicyDocumentSpec{
			DocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:PutObject","Resource":"*"}]}`,
			StatementCount:       1,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("seed reconcile: %v", err)
	}

	doc.Spec.DocumentJSON = `{"Version":"2012-10-17",`
	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile invalid json: %v", err)
	}

	if doc.Status.ResolvedDocumentJSON != "" {
		t.Fatalf("expected resolved json cleared, got %s", doc.Status.ResolvedDocumentJSON)
	}
	if doc.Status.StatementCount != 0 {
		t.Fatalf("expected statement count cleared, got %d", doc.Status.StatementCount)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionFalse) {
		t.Fatalf("ready condition not false: %#v", doc.Status.Conditions)
	}
}

func TestReconcileSkipsStatusUpdateWhenStatusIsUnchanged(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "raw", Namespace: "default", Generation: 3},
		Spec: v1alpha1.PolicyDocumentSpec{
			DocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
	}

	calc := &Reconciler{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()}
	if _, err := calc.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("seed reconcile: %v", err)
	}

	stored := &v1alpha1.PolicyDocument{
		TypeMeta:   doc.TypeMeta,
		ObjectMeta: doc.ObjectMeta,
		Spec:       doc.Spec,
		Status:     snapshotStatus(doc.Status),
	}

	baseClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(stored).Build()
	r := &Reconciler{Client: &recordingClient{Client: baseClient}}

	if _, err := r.Reconcile(context.Background(), ctrlRequest("default", "raw")); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if rc, ok := r.Client.(*recordingClient); !ok {
		t.Fatalf("unexpected client type %T", r.Client)
	} else if rc.statusUpdates != 0 {
		t.Fatalf("expected no status updates, got %d", rc.statusUpdates)
	}
}

func TestReconcileResolvesPredictedArnRef(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	role := &unstructured.Unstructured{}
	role.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	role.SetNamespace("default")
	role.SetName("lambda-exec")
	if err := unstructured.SetNestedField(role.Object, "arn:aws:iam::123456789012:role/lambda-exec", "status", "predictedArn"); err != nil {
		t.Fatalf("seed role status: %v", err)
	}

	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "policy", Namespace: "default", Generation: 5},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Sid:     "AllowLambdaInvoke",
					Effect:  "Allow",
					Actions: []string{"lambda:InvokeFunction"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "lambda-exec"}},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc, role).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if !jsonContains(doc.Status.ResolvedDocumentJSON, "arn:aws:iam::123456789012:role/lambda-exec") {
		t.Fatalf("resolved json missing arn: %s", doc.Status.ResolvedDocumentJSON)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("ready condition not true: %#v", doc.Status.Conditions)
	}
}

func TestReconcileMarksSourceNotReadyWhenRefPending(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "policy", Namespace: "default", Generation: 2},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"lambda:InvokeFunction"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "lambda-exec"}},
					},
				},
			},
		},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
			StatementCount:       1,
			SourceCount:          1,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	r := &Reconciler{Client: c, RequeueAfter: 5 * time.Second}

	result, err := r.reconcileDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if result.RequeueAfter != 5*time.Second {
		t.Fatalf("expected requeue after 5s, got %s", result.RequeueAfter)
	}

	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady, metav1.ConditionTrue) {
		t.Fatalf("source not ready not true: %#v", doc.Status.Conditions)
	}
	if conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("ready should be false: %#v", doc.Status.Conditions)
	}
	if doc.Status.ResolvedDocumentJSON != "" {
		t.Fatalf("expected resolved json cleared, got %s", doc.Status.ResolvedDocumentJSON)
	}
	if doc.Status.StatementCount != 0 {
		t.Fatalf("expected statement count cleared, got %d", doc.Status.StatementCount)
	}
}

func TestReconcileResolvesPrincipalRefToPredictedArn(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	bucket := &unstructured.Unstructured{}
	bucket.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSS3Bucket"})
	bucket.SetNamespace("default")
	bucket.SetName("my-bucket")
	if err := unstructured.SetNestedField(bucket.Object, "arn:aws:s3:::my-bucket", "status", "predictedArn"); err != nil {
		t.Fatalf("seed bucket status: %v", err)
	}

	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "bucket-policy", Namespace: "default", Generation: 6},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"s3:GetObject"},
					Principals: []v1alpha1.PolicyPrincipal{
						{Type: "AWS", Ref: &v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "my-bucket"}},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc, bucket).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if !jsonContains(doc.Status.ResolvedDocumentJSON, "arn:aws:s3:::my-bucket") {
		t.Fatalf("resolved json missing bucket arn: %s", doc.Status.ResolvedDocumentJSON)
	}
	if !jsonContains(doc.Status.ResolvedDocumentJSON, `"Principal":{"AWS":"arn:aws:s3:::my-bucket"}`) {
		t.Fatalf("resolved json missing principal: %s", doc.Status.ResolvedDocumentJSON)
	}
}

func TestReconcileSucceedsAfterRefBecomesReady(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	role := &unstructured.Unstructured{}
	role.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSIAMRole"})
	role.SetNamespace("default")
	role.SetName("lambda-exec")

	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "policy", Namespace: "default", Generation: 7},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"lambda:InvokeFunction"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "lambda-exec"}},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc, role).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("first reconcile: %v", err)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady, metav1.ConditionTrue) {
		t.Fatalf("expected pending state first: %#v", doc.Status.Conditions)
	}

	if err := unstructured.SetNestedField(role.Object, "arn:aws:iam::123456789012:role/lambda-exec", "status", "predictedArn"); err != nil {
		t.Fatalf("seed role status: %v", err)
	}
	if err := c.Update(context.Background(), role); err != nil {
		t.Fatalf("update role: %v", err)
	}

	doc.Status = v1alpha1.PolicyDocumentStatus{}
	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("second reconcile: %v", err)
	}

	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("expected ready after ref population: %#v", doc.Status.Conditions)
	}
	if !jsonContains(doc.Status.ResolvedDocumentJSON, "arn:aws:iam::123456789012:role/lambda-exec") {
		t.Fatalf("resolved json missing arn after ready: %s", doc.Status.ResolvedDocumentJSON)
	}
	if conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady, metav1.ConditionTrue) {
		t.Fatalf("source not ready should clear: %#v", doc.Status.Conditions)
	}
}

func TestResolveRefUsesStatusArn(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "aws.kropath.run", Version: "v1alpha1", Kind: "AWSIAMPolicy"})
	obj.SetNamespace("default")
	obj.SetName("policy")
	if err := unstructured.SetNestedField(obj.Object, "arn:aws:iam::123456789012:policy/example", "status", "arn"); err != nil {
		t.Fatalf("seed arn: %v", err)
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).Build()
	arn, pending, err := resolveRef(context.Background(), c, "default", &v1alpha1.PolicyRef{Kind: "AWSIAMPolicy", Name: "policy", Field: "arn"})
	if err != nil {
		t.Fatalf("resolve ref: %v", err)
	}
	if pending {
		t.Fatalf("expected resolved ref")
	}
	if arn != "arn:aws:iam::123456789012:policy/example" {
		t.Fatalf("unexpected arn %q", arn)
	}
}

// noMatchKindClient simulates a ref kind whose CRD is not yet registered on
// the API server, which client-go reports as a *meta.NoKindMatchError rather
// than a NotFound error.
type noMatchKindClient struct {
	client.Client
}

func (noMatchKindClient) Get(_ context.Context, _ client.ObjectKey, _ client.Object, _ ...client.GetOption) error {
	return &apimeta.NoKindMatchError{
		GroupKind:        schema.GroupKind{Group: "aws.kropath.run", Kind: "AWSLambdaFunction"},
		SearchedVersions: []string{"v1alpha1"},
	}
}

func TestResolveRefTreatsMissingCRDAsPending(t *testing.T) {
	arn, pending, err := resolveRef(context.Background(), noMatchKindClient{}, "default", &v1alpha1.PolicyRef{Kind: "AWSLambdaFunction", Name: "fn"})
	if err != nil {
		t.Fatalf("expected no error when the ref's CRD is not yet registered, got: %v", err)
	}
	if !pending {
		t.Fatalf("expected pending=true when the ref's CRD is not yet registered")
	}
	if arn != "" {
		t.Fatalf("expected empty arn, got %q", arn)
	}
}

func TestRequestsForSourceDocumentReturnsParentsReferencingSource(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	source := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "base-policy", Namespace: "default"},
	}
	related := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "app-policy", Namespace: "default"},
		Spec: v1alpha1.PolicyDocumentSpec{
			Sources: []v1alpha1.PolicyDocumentSource{{Name: "base-policy"}},
		},
	}
	unrelated := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "standalone", Namespace: "default"},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(source, related, unrelated).Build()
	r := &Reconciler{Client: c}

	requests, err := r.requestsForSourceDocument(context.Background(), "default", "base-policy")
	if err != nil {
		t.Fatalf("requests for source document: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one parent request, got %d: %#v", len(requests), requests)
	}
	if requests[0].Namespace != "default" || requests[0].Name != "app-policy" {
		t.Fatalf("unexpected request: %#v", requests[0])
	}

	requests, err = r.requestsForSourceDocument(context.Background(), "default", "missing")
	if err != nil {
		t.Fatalf("requests for missing source document: %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("expected no parent requests for missing source, got %#v", requests)
	}
}

func TestRequestsForReferencedResourceReturnsMatchingDocuments(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	match := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "match", Namespace: "default"},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"s3:GetObject"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSIAMRole", Name: "lambda-exec"}},
					},
				},
			},
		},
	}
	unrelated := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: "default"},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Effect:  "Allow",
					Actions: []string{"s3:GetObject"},
					Resources: []v1alpha1.PolicyResource{
						{Ref: &v1alpha1.PolicyRef{Kind: "AWSS3Bucket", Name: "my-bucket"}},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(match, unrelated).Build()
	r := &Reconciler{Client: c}

	requests, err := r.requestsForReferencedResource(context.Background(), "default", "lambda-exec", "AWSIAMRole")
	if err != nil {
		t.Fatalf("requests for referenced resource: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected one matching request, got %d: %#v", len(requests), requests)
	}
	if requests[0].Name != "match" {
		t.Fatalf("unexpected request: %#v", requests[0])
	}

	requests, err = r.requestsForReferencedResource(context.Background(), "default", "lambda-exec", "AWSS3Bucket")
	if err != nil {
		t.Fatalf("requests for mismatched kind: %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("expected no requests for mismatched kind, got %#v", requests)
	}
}

func TestReconcileMergesSourceStatementsBeforeOwnStatements(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	source := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "base", Namespace: "default", Generation: 1},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Sid":"BaseAllow","Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::base-bucket"}]}`,
		},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Sid:     "BaseAllow",
					Effect:  "Allow",
					Actions: []string{"s3:GetObject"},
					Resources: []v1alpha1.PolicyResource{
						{ARN: "arn:aws:s3:::base-bucket"},
					},
				},
			},
		},
	}
	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "child", Namespace: "default", Generation: 2},
		Spec: v1alpha1.PolicyDocumentSpec{
			Sources: []v1alpha1.PolicyDocumentSource{{Name: "base"}},
			Statements: []v1alpha1.PolicyStatement{
				{
					Sid:     "ChildAllow",
					Effect:  "Allow",
					Actions: []string{"s3:PutObject"},
					Resources: []v1alpha1.PolicyResource{
						{ARN: "arn:aws:s3:::child-bucket"},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(source, doc).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile child: %v", err)
	}

	parsed := mustParsePolicyDocument(t, doc.Status.ResolvedDocumentJSON)
	if got, want := len(parsed.Statement), 2; got != want {
		t.Fatalf("expected %d statements, got %d", want, got)
	}
	if got, want := parsed.Statement[0].Sid, "BaseAllow"; got != want {
		t.Fatalf("expected first statement Sid %q, got %q", want, got)
	}
	if got, want := parsed.Statement[1].Sid, "ChildAllow"; got != want {
		t.Fatalf("expected second statement Sid %q, got %q", want, got)
	}
}

func TestReconcileRejectsMergeFromRawSource(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	source := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "base", Namespace: "default", Generation: 1},
		Spec: v1alpha1.PolicyDocumentSpec{
			DocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
	}
	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "child", Namespace: "default", Generation: 2},
		Spec: v1alpha1.PolicyDocumentSpec{
			Sources: []v1alpha1.PolicyDocumentSource{{Name: "base"}},
		},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Sid":"Keep","Effect":"Allow","Action":"s3:ListBucket","Resource":"*"}]}`,
			StatementCount:       1,
			SourceCount:          1,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(source, doc).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if got, want := doc.Status.ResolvedDocumentJSON, `{"Version":"2012-10-17","Statement":[{"Sid":"Keep","Effect":"Allow","Action":"s3:ListBucket","Resource":"*"}]}`; got != want {
		t.Fatalf("resolved document changed unexpectedly: got %s", got)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionFalse) {
		t.Fatalf("ready should be false: %#v", doc.Status.Conditions)
	}
	if got, want := conditionReason(doc.Status.Conditions, v1alpha1.ConditionReady), "MergeFromRawNotSupported"; got != want {
		t.Fatalf("unexpected ready reason: got %q want %q", got, want)
	}
}

func TestReconcileDetectsAndClearsSidConflicts(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	source := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "base", Namespace: "default", Generation: 1},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Sid":"AllowS3Read","Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		},
	}
	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "child", Namespace: "default", Generation: 2},
		Spec: v1alpha1.PolicyDocumentSpec{
			Sources: []v1alpha1.PolicyDocumentSource{{Name: "base"}},
			Statements: []v1alpha1.PolicyStatement{
				{
					Sid:     "AllowS3Read",
					Effect:  "Allow",
					Actions: []string{"s3:PutObject"},
					Resources: []v1alpha1.PolicyResource{
						{ARN: "arn:aws:s3:::child-bucket"},
					},
				},
			},
		},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Sid":"Keep","Effect":"Allow","Action":"s3:ListBucket","Resource":"*"}]}`,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(source, doc).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile with conflict: %v", err)
	}

	if got, want := doc.Status.ResolvedDocumentJSON, `{"Version":"2012-10-17","Statement":[{"Sid":"Keep","Effect":"Allow","Action":"s3:ListBucket","Resource":"*"}]}`; got != want {
		t.Fatalf("resolved document changed unexpectedly: got %s", got)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSidConflict, metav1.ConditionTrue) {
		t.Fatalf("sid conflict should be true: %#v", doc.Status.Conditions)
	}
	if got, want := conditionReason(doc.Status.Conditions, v1alpha1.ConditionSidConflict), "SidConflict"; got != want {
		t.Fatalf("unexpected sid conflict reason: got %q want %q", got, want)
	}
	if got := conditionMessage(doc.Status.Conditions, v1alpha1.ConditionSidConflict); !strings.Contains(got, "AllowS3Read") {
		t.Fatalf("sid conflict message missing sid: %q", got)
	}

	doc.Spec.Statements[0].Sid = "AllowS3Write"
	if _, err := r.reconcileDocument(context.Background(), doc); err != nil {
		t.Fatalf("reconcile after conflict resolution: %v", err)
	}

	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSidConflict, metav1.ConditionFalse) {
		t.Fatalf("sid conflict should be false after rename: %#v", doc.Status.Conditions)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionReady, metav1.ConditionTrue) {
		t.Fatalf("ready should be true after rename: %#v", doc.Status.Conditions)
	}
	parsed := mustParsePolicyDocument(t, doc.Status.ResolvedDocumentJSON)
	if got, want := len(parsed.Statement), 2; got != want {
		t.Fatalf("expected %d statements, got %d", want, got)
	}
	if got, want := parsed.Statement[0].Sid, "AllowS3Read"; got != want {
		t.Fatalf("expected first statement Sid %q, got %q", want, got)
	}
	if got, want := parsed.Statement[1].Sid, "AllowS3Write"; got != want {
		t.Fatalf("expected second statement Sid %q, got %q", want, got)
	}
}

func TestReconcileMarksSourceMissingAndRequeues(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	doc := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "child", Namespace: "default", Generation: 2},
		Spec: v1alpha1.PolicyDocumentSpec{
			Sources: []v1alpha1.PolicyDocumentSource{{Name: "missing"}},
		},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Sid":"Keep","Effect":"Allow","Action":"s3:ListBucket","Resource":"*"}]}`,
			StatementCount:       1,
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(doc).Build()
	r := &Reconciler{Client: c}

	result, err := r.reconcileDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if result.RequeueAfter != defaultRequeueAfter {
		t.Fatalf("expected requeue after %s, got %s", defaultRequeueAfter, result.RequeueAfter)
	}
	if !conditionHasStatus(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady, metav1.ConditionTrue) {
		t.Fatalf("source not ready should be true: %#v", doc.Status.Conditions)
	}
	if got, want := conditionReason(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady), "SourceMissing"; got != want {
		t.Fatalf("unexpected source not ready reason: got %q want %q", got, want)
	}
	if got := conditionMessage(doc.Status.Conditions, v1alpha1.ConditionSourceNotReady); !strings.Contains(got, "missing") {
		t.Fatalf("source not ready message missing source name: %q", got)
	}
	if got, want := doc.Status.ResolvedDocumentJSON, ""; got != want {
		t.Fatalf("expected resolved document cleared, got %q", got)
	}
}

func TestReconcileMergesTransitiveSourcesThroughResolvedDocuments(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	base := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "base", Namespace: "default", Generation: 1},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Sid":"BaseAllow","Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::base-bucket"}]}`,
		},
		Spec: v1alpha1.PolicyDocumentSpec{
			Statements: []v1alpha1.PolicyStatement{
				{
					Sid:     "BaseAllow",
					Effect:  "Allow",
					Actions: []string{"s3:GetObject"},
					Resources: []v1alpha1.PolicyResource{
						{ARN: "arn:aws:s3:::base-bucket"},
					},
				},
			},
		},
	}
	mid := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "mid", Namespace: "default", Generation: 2},
		Status: v1alpha1.PolicyDocumentStatus{
			ResolvedDocumentJSON: `{"Version":"2012-10-17","Statement":[{"Sid":"BaseAllow","Effect":"Allow","Action":"s3:GetObject","Resource":"arn:aws:s3:::base-bucket"},{"Sid":"MidAllow","Effect":"Allow","Action":"s3:PutObject","Resource":"arn:aws:s3:::mid-bucket"}]}`,
		},
		Spec: v1alpha1.PolicyDocumentSpec{
			Sources: []v1alpha1.PolicyDocumentSource{{Name: "base"}},
			Statements: []v1alpha1.PolicyStatement{
				{
					Sid:     "MidAllow",
					Effect:  "Allow",
					Actions: []string{"s3:PutObject"},
					Resources: []v1alpha1.PolicyResource{
						{ARN: "arn:aws:s3:::mid-bucket"},
					},
				},
			},
		},
	}
	leaf := &v1alpha1.PolicyDocument{
		TypeMeta:   metav1.TypeMeta{APIVersion: "aws.kropath.run/v1alpha1", Kind: v1alpha1.PolicyDocumentKind},
		ObjectMeta: metav1.ObjectMeta{Name: "leaf", Namespace: "default", Generation: 3},
		Spec: v1alpha1.PolicyDocumentSpec{
			Sources: []v1alpha1.PolicyDocumentSource{{Name: "mid"}},
			Statements: []v1alpha1.PolicyStatement{
				{
					Sid:     "LeafAllow",
					Effect:  "Allow",
					Actions: []string{"s3:DeleteObject"},
					Resources: []v1alpha1.PolicyResource{
						{ARN: "arn:aws:s3:::leaf-bucket"},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(base, mid, leaf).Build()
	r := &Reconciler{Client: c}

	if _, err := r.reconcileDocument(context.Background(), leaf); err != nil {
		t.Fatalf("reconcile leaf: %v", err)
	}

	parsed := mustParsePolicyDocument(t, leaf.Status.ResolvedDocumentJSON)
	if got, want := len(parsed.Statement), 3; got != want {
		t.Fatalf("expected %d statements, got %d", want, got)
	}
	if got, want := parsed.Statement[0].Sid, "BaseAllow"; got != want {
		t.Fatalf("expected first statement Sid %q, got %q", want, got)
	}
	if got, want := parsed.Statement[1].Sid, "MidAllow"; got != want {
		t.Fatalf("expected second statement Sid %q, got %q", want, got)
	}
	if got, want := parsed.Statement[2].Sid, "LeafAllow"; got != want {
		t.Fatalf("expected third statement Sid %q, got %q", want, got)
	}
}

func jsonContains(raw, needle string) bool {
	return strings.Contains(raw, needle)
}

func mustParsePolicyDocument(t *testing.T, raw string) policyDocumentJSON {
	t.Helper()

	var doc policyDocumentJSON
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("parse policy document: %v", err)
	}
	return doc
}

func conditionHasStatus(conditions []metav1.Condition, conditionType string, status metav1.ConditionStatus) bool {
	for _, cond := range conditions {
		if cond.Type == conditionType && cond.Status == status {
			return true
		}
	}
	return false
}

func conditionReason(conditions []metav1.Condition, conditionType string) string {
	for _, cond := range conditions {
		if cond.Type == conditionType {
			return cond.Reason
		}
	}
	return ""
}

func conditionMessage(conditions []metav1.Condition, conditionType string) string {
	for _, cond := range conditions {
		if cond.Type == conditionType {
			return cond.Message
		}
	}
	return ""
}

func ctrlRequest(namespace, name string) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: namespace, Name: name}}
}
