// Copyright 2026 The kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	goruntime "runtime"
	"time"

	"github.com/kropath/kropath-controller/api/v1alpha1"
	"github.com/kropath/kropath-controller/internal/features"
	"github.com/kropath/kropath-controller/internal/metrics"
	"github.com/kropath/kropath-controller/internal/registry"
	"github.com/kropath/kropath-controller/internal/version"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

var (
	metricsAddr string
	probeAddr   string
)

func leaderElectionNamespace() string {
	if ns := os.Getenv("LEADER_ELECTION_NAMESPACE"); ns != "" {
		return ns
	}
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns
	}
	return "default"
}

func main() {
	// "features" subcommand: print the same JSON as GET /features, then exit.
	// No manager, no kubeconfig, no cluster connection — safe for CI and laptops.
	if len(os.Args) > 1 && os.Args[1] == "features" {
		resp := features.Response{
			Version:   version.Version,
			GitCommit: version.GitCommit,
			BuildDate: version.BuildDate,
			GoVersion: goruntime.Version(),
			Features:  features.All,
		}
		out, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
		return
	}

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctrl.Log.Info("starting manager",
		"version", version.Version,
		"gitCommit", version.GitCommit,
		"buildDate", version.BuildDate,
		"metrics", metricsAddr,
		"probes", probeAddr,
		"reconcilers", len(features.All),
	)

	sch := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(sch); err != nil {
		ctrl.Log.Error(err, "unable to add client-go scheme")
		os.Exit(1)
	}
	if err := v1alpha1.AddToScheme(sch); err != nil {
		ctrl.Log.Error(err, "unable to add api scheme")
		os.Exit(1)
	}

	restCfg := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(restCfg, ctrl.Options{
		Scheme:                  sch,
		Metrics:                 server.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress:  probeAddr,
		LeaderElection:          true,
		LeaderElectionNamespace: leaderElectionNamespace(),
		LeaderElectionID:        "kropath-controller.aws.kropath.run",
	})
	if err != nil {
		ctrl.Log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Build the reconciler registry and gate each entry on CRD availability.
	// Entries whose Required GVKs are all served become active immediately.
	// Entries with missing Required GVKs stay pending; the CRD watcher (KRO-849)
	// promotes them when their CRD later appears.
	dc, err := discovery.NewDiscoveryClientForConfig(restCfg)
	if err != nil {
		ctrl.Log.Error(err, "unable to create discovery client")
		os.Exit(1)
	}
	servedGVKs, err := registry.GatherServedGVKs(dc)
	if err != nil {
		ctrl.Log.Error(err, "unable to discover served GVKs")
		os.Exit(1)
	}

	coord := &registry.Coordinator{}
	for _, e := range registry.All() {
		coord.Add(e)
	}
	bctx := registry.BuildCtx{
		Manager: mgr,
		Log:     ctrl.Log,
	}
	if err := coord.RunGate(bctx, servedGVKs); err != nil {
		ctrl.Log.Error(err, "startup gate failed")
		os.Exit(1)
	}

	// /features endpoint: returns version + live reconciler list with per-reconciler state.
	if err := mgr.AddMetricsServerExtraHandler("/features", features.Handler(
		version.Version, version.GitCommit, version.BuildDate, goruntime.Version(), features.All, coord,
	)); err != nil {
		ctrl.Log.Error(err, "unable to register /features handler")
		os.Exit(1)
	}

	// CRD watcher: activates pending reconcilers when their CRDs arrive at runtime.
	if err := mgr.Add(registry.NewWatcher(coord, bctx)); err != nil {
		ctrl.Log.Error(err, "unable to add CRD watcher")
		os.Exit(1)
	}

	// Scrape-time metrics collectors (design §6, M-14): registers every
	// collector added via metrics.RegisterCollectorFactory once the manager's
	// cache has synced, on every replica -- see internal/metrics for the
	// self-registration seam feature packages use to add their own
	// collector(s) without editing this line.
	if err := mgr.Add(metrics.NewCollectorRunnable(mgr.GetCache(), mgr.GetAPIReader(), ctrlmetrics.Registry)); err != nil {
		ctrl.Log.Error(err, "unable to add metrics collector runnable")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		ctrl.Log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", func(req *http.Request) error {
		ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
		defer cancel()
		if !mgr.GetCache().WaitForCacheSync(ctx) {
			return fmt.Errorf("informer cache not synced")
		}
		return nil
	}); err != nil {
		ctrl.Log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		ctrl.Log.Error(err, "problem running manager")
		os.Exit(1)
	}
}
