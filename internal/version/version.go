// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

// Package version holds build-time metadata stamped via -ldflags.
// The Makefile and the release workflow both set these at link time.
package version

// These variables are overridden at link time by the Makefile.
// -ldflags "-X github.com/kropath/kropath-aws-controller/internal/version.Version=v0.1.0 ..."
var (
	Version   = "dev"     // semantic version, e.g. "v0.1.2"
	GitCommit = "none"    // 7-char git SHA, e.g. "abc1234"
	BuildDate = "unknown" // ISO-8601 UTC, e.g. "2026-08-13T00:00:00Z"
)
