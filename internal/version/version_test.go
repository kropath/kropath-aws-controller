// Copyright 2026 kropath Authors.
// SPDX-License-Identifier: Apache-2.0

package version_test

import (
	"testing"

	"github.com/kropath/kropath-aws-controller/internal/version"
)

func TestDefaultValuesAreNonEmpty(t *testing.T) {
	if version.Version == "" {
		t.Error("version.Version must not be empty")
	}
	if version.GitCommit == "" {
		t.Error("version.GitCommit must not be empty")
	}
	if version.BuildDate == "" {
		t.Error("version.BuildDate must not be empty")
	}
}
