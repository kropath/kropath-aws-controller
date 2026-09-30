// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package features_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/features"
)

var testReconcilers = []features.Reconciler{
	{Name: "AlphaConfig", Package: "alphaconfig", Description: "Alpha reconciler.", Kinds: []string{"AlphaConfig", "KropathConfig"}, SinceVersion: "v0.0.1", Stability: "stable"},
	{Name: "BetaConfig", Package: "betaconfig", Description: "Beta reconciler.", Kinds: []string{"BetaConfig", "KropathConfig"}, SinceVersion: "v0.0.1", Stability: "stable"},
}

func newTestHandler() http.Handler {
	return features.Handler("v0.1.0", "abc1234", "2026-08-13T00:00:00Z", "go1.26.0", testReconcilers, nil)
}

// TestHandlerSuccess: full GET returns 200, application/json, all features.
func TestHandlerSuccess(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/features", nil)
	w := httptest.NewRecorder()
	newTestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	var resp features.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody: %s", err, w.Body.String())
	}
	if resp.Version != "v0.1.0" {
		t.Errorf("version: got %q, want v0.1.0", resp.Version)
	}
	if resp.GitCommit != "abc1234" {
		t.Errorf("gitCommit: got %q, want abc1234", resp.GitCommit)
	}
	if resp.GoVersion == "" {
		t.Error("goVersion must not be empty")
	}
	if len(resp.Features) != len(testReconcilers) {
		t.Errorf("features count: got %d, want %d", len(resp.Features), len(testReconcilers))
	}
}

// TestHandlerSingleFeatureLookup: ?name=<pkg> returns exactly that reconciler.
func TestHandlerSingleFeatureLookup(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/features?name=alphaconfig", nil)
	w := httptest.NewRecorder()
	newTestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	var resp features.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody: %s", err, w.Body.String())
	}
	if len(resp.Features) != 1 {
		t.Fatalf("want 1 feature, got %d", len(resp.Features))
	}
	if resp.Features[0].Package != "alphaconfig" {
		t.Errorf("feature package: got %q, want alphaconfig", resp.Features[0].Package)
	}
	if resp.Features[0].Name != "AlphaConfig" {
		t.Errorf("feature name: got %q, want AlphaConfig", resp.Features[0].Name)
	}
}

// TestHandlerUnknownFeature: ?name=<unknown> returns 404 with a JSON error body.
func TestHandlerUnknownFeature(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/features?name=doesnotexist", nil)
	w := httptest.NewRecorder()
	newTestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody: %s", err, w.Body.String())
	}
	if errResp.Error == "" {
		t.Error("error field must not be empty in 404 response")
	}
}

// TestHandlerWrongMethod: POST returns 405 with a JSON body and Allow header.
func TestHandlerWrongMethod(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/features", nil)
			w := httptest.NewRecorder()
			newTestHandler().ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Fatalf("want 405, got %d", w.Code)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type: got %q, want application/json", ct)
			}
			if allow := w.Header().Get("Allow"); allow == "" {
				t.Error("Allow header must be set on 405")
			}
			var errResp struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
				t.Fatalf("body is not valid JSON: %v\nbody: %s", err, w.Body.String())
			}
		})
	}
}

// TestHandlerEmptyRegistry: nil registry returns 200 with an empty JSON array, not null.
func TestHandlerEmptyRegistry(t *testing.T) {
	h := features.Handler("dev", "none", "unknown", "go1.26.0", nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/features", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp features.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody: %s", err, w.Body.String())
	}
	if resp.Features == nil {
		t.Error("features field must not be null for an empty registry — want []")
	}
	if len(resp.Features) != 0 {
		t.Errorf("want 0 features, got %d", len(resp.Features))
	}
}

// stubStateReader implements features.StateReader for tests.
type stubStateReader struct {
	states map[string]string
	missing map[string][]string
}

func (s *stubStateReader) ReconcilerState(pkg string) (string, []string) {
	return s.states[pkg], s.missing[pkg]
}

// TestHandlerStateReader_ActiveAndPending: coordinator populates state and missingKinds per-request.
func TestHandlerStateReader_ActiveAndPending(t *testing.T) {
	sr := &stubStateReader{
		states:  map[string]string{"alphaconfig": "active", "betaconfig": "pending"},
		missing: map[string][]string{"betaconfig": {"BetaConfig"}},
	}
	h := features.Handler("v0.1.0", "abc1234", "2026-08-13T00:00:00Z", "go1.26.0", testReconcilers, sr)
	req := httptest.NewRequest(http.MethodGet, "/features", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp features.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	for _, rec := range resp.Features {
		switch rec.Package {
		case "alphaconfig":
			if rec.State != "active" {
				t.Errorf("alphaconfig state: got %q, want active", rec.State)
			}
			if len(rec.MissingKinds) != 0 {
				t.Errorf("alphaconfig missingKinds: got %v, want empty", rec.MissingKinds)
			}
		case "betaconfig":
			if rec.State != "pending" {
				t.Errorf("betaconfig state: got %q, want pending", rec.State)
			}
			if len(rec.MissingKinds) != 1 || rec.MissingKinds[0] != "BetaConfig" {
				t.Errorf("betaconfig missingKinds: got %v, want [BetaConfig]", rec.MissingKinds)
			}
		}
	}
}

// TestHandlerStateReader_NameFilter: ?name= query also populates state.
func TestHandlerStateReader_NameFilter(t *testing.T) {
	sr := &stubStateReader{
		states:  map[string]string{"alphaconfig": "pending"},
		missing: map[string][]string{"alphaconfig": {"AlphaConfig"}},
	}
	h := features.Handler("v0.1.0", "abc1234", "2026-08-13T00:00:00Z", "go1.26.0", testReconcilers, sr)
	req := httptest.NewRequest(http.MethodGet, "/features?name=alphaconfig", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	var resp features.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if len(resp.Features) != 1 {
		t.Fatalf("want 1 feature, got %d", len(resp.Features))
	}
	if resp.Features[0].State != "pending" {
		t.Errorf("state: got %q, want pending", resp.Features[0].State)
	}
	if len(resp.Features[0].MissingKinds) != 1 {
		t.Errorf("missingKinds: got %v, want [AlphaConfig]", resp.Features[0].MissingKinds)
	}
}

// TestHandlerHeadMethod: HEAD returns 200 with the same headers as GET, no body.
func TestHandlerHeadMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodHead, "/features", nil)
	w := httptest.NewRecorder()
	newTestHandler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}
