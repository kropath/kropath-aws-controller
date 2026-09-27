// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// M-14: the whole design §9 max-by aggregation contract rests on every
// replica exposing the same collector series, which requires this to be
// false.
func TestCollectorRunnable_NeedLeaderElection(t *testing.T) {
	r := NewCollectorRunnable(nil, nil, prometheus.NewRegistry())
	if r.NeedLeaderElection() {
		t.Error("NeedLeaderElection() = true, want false (M-14)")
	}
}

// withIsolatedFactories saves and restores the package-level factories slice
// so this test does not interact with factories registered elsewhere.
func withIsolatedFactories(t *testing.T) {
	t.Helper()
	factoriesMu.Lock()
	saved := factories
	factories = nil
	factoriesMu.Unlock()
	t.Cleanup(func() {
		factoriesMu.Lock()
		factories = saved
		factoriesMu.Unlock()
	})
}

// TestCollectorRunnable_RegistrationSeam proves the self-registration seam:
// a factory added via RegisterCollectorFactory before Start is called is
// invoked with Start's reader/apiReader and its collector is registered
// against the given registerer -- all without collectorRunnable or
// cmd/manager/main.go knowing which feature package the factory came from.
func TestCollectorRunnable_RegistrationSeam(t *testing.T) {
	withIsolatedFactories(t)

	testCollector := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mtest_registration_seam_total",
		Help: "test collector proving the self-registration seam.",
	})

	called := make(chan struct{})
	var gotReader, gotAPIReader client.Reader
	wantReader, wantAPIReader := fakeReader{name: "cache"}, fakeReader{name: "api"}
	RegisterCollectorFactory(func(reader, apiReader client.Reader) prometheus.Collector {
		gotReader, gotAPIReader = reader, apiReader
		close(called)
		return testCollector
	})

	reg := prometheus.NewRegistry()
	r := NewCollectorRunnable(wantReader, wantAPIReader, reg)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- r.Start(ctx) }()

	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("factory was never called")
	}

	if gotReader != wantReader || gotAPIReader != wantAPIReader {
		t.Errorf("factory got reader=%v apiReader=%v, want %v / %v", gotReader, gotAPIReader, wantReader, wantAPIReader)
	}

	// Start's Register call happens just after the factory returns, so poll
	// briefly rather than racing a single Gather against that return.
	deadline := time.Now().Add(2 * time.Second)
	var found bool
	for !found && time.Now().Before(deadline) {
		mfs, err := reg.Gather()
		if err != nil {
			t.Fatalf("Gather: %v", err)
		}
		for _, mf := range mfs {
			if mf.GetName() == "mtest_registration_seam_total" {
				found = true
			}
		}
		if !found {
			time.Sleep(10 * time.Millisecond)
		}
	}
	if !found {
		t.Error("expected the factory's collector to be registered against the given registerer")
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after ctx cancellation")
	}
}

// fakeReader is a distinguishable client.Reader stand-in -- it is never
// called; the test only asserts identity, not behaviour.
type fakeReader struct {
	client.Reader
	name string
}
