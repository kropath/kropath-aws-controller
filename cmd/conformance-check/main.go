// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

// conformance-check inspects an existing ACK/kro install and reports whether
// it satisfies the account/region placement preconditions kropath depends on
// (ADR-015 §5.8.4, KRO-1139, KRO-1141). It is a best-effort, read-only
// diagnostic — it never modifies cluster state and it is not a substitute
// for the runtime annotation checks kropath-aws-controller performs on every
// reconcile.
//
// Usage:
//
//	go run ./cmd/conformance-check --ack-namespace ack-system
//	go run ./cmd/conformance-check --ack-namespace ack-system --namespaces payments-prod,data-prod
//	go run ./cmd/conformance-check --ack-namespace ack-system --output json
//
// Exit codes: 0 = no blocking findings, 1 = at least one blocking finding
// (or, with --strict, at least one finding the checker could not verify),
// 2 = the checker could not run at all (e.g. cannot reach the API server).
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kropath/kropath-aws-controller/internal/carmcheck"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("conformance-check", flag.ContinueOnError)
	fs.SetOutput(stderr)

	ackNamespace := fs.String("ack-namespace", "", "Namespace hosting the ACK controller Deployments and the ack-role-account-map ConfigMap (required)")
	roleAccountMapName := fs.String("role-account-map", "", `Name of the ack-role-account-map ConfigMap (default "ack-role-account-map")`)
	namespacesFlag := fs.String("namespaces", "", "Comma-separated list of namespaces to check, overriding auto-discovery")
	output := fs.String("output", "text", `Output format: "text" or "json"`)
	strict := fs.Bool("strict", false, "Exit non-zero if any precondition could not be verified, not only on confirmed violations")
	timeout := fs.Duration("timeout", 30*time.Second, "Timeout for the API server calls this check makes")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *ackNamespace == "" {
		_, _ = fmt.Fprintln(stderr, "conformance-check: --ack-namespace is required")
		fs.Usage()
		return 2
	}
	if *output != "text" && *output != "json" {
		_, _ = fmt.Fprintf(stderr, "conformance-check: --output must be \"text\" or \"json\", got %q\n", *output)
		return 2
	}

	var namespaces []string
	if *namespacesFlag != "" {
		for _, ns := range strings.Split(*namespacesFlag, ",") {
			if ns = strings.TrimSpace(ns); ns != "" {
				namespaces = append(namespaces, ns)
			}
		}
	}

	sch := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(sch); err != nil {
		_, _ = fmt.Fprintf(stderr, "conformance-check: unable to build scheme: %v\n", err)
		return 2
	}

	restCfg, err := ctrl.GetConfig()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "conformance-check: unable to load kubeconfig: %v\n", err)
		return 2
	}

	c, err := client.New(restCfg, client.Options{Scheme: sch})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "conformance-check: unable to build client: %v\n", err)
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	report, err := carmcheck.Run(ctx, c, carmcheck.Options{
		ACKNamespace:       *ackNamespace,
		RoleAccountMapName: *roleAccountMapName,
		Namespaces:         namespaces,
	})
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "conformance-check: %v\n", err)
		return 2
	}

	var writeErr error
	if *output == "json" {
		writeErr = report.WriteJSON(stdout)
	} else {
		writeErr = report.WriteText(stdout)
	}
	if writeErr != nil {
		_, _ = fmt.Fprintf(stderr, "conformance-check: writing report: %v\n", writeErr)
		return 2
	}

	if report.HasBlocking() {
		return 1
	}
	if *strict && report.HasUnknown() {
		return 1
	}
	return 0
}
