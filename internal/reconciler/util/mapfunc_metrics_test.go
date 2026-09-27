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

package util

import (
	"errors"
	"testing"

	"github.com/go-logr/logr"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

// cascadeMapfuncErrorsTotalValue reads kropath_cascade_mapfunc_errors_total
// for the given labels straight from the shared registry -- internal/metrics
// registers its own counter and does not export it, so this is the only
// black-box way to assert the increment from this package.
func cascadeMapfuncErrorsTotalValue(t *testing.T, family, trigger string) float64 {
	t.Helper()
	families, err := ctrlmetrics.Registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}
	for _, mf := range families {
		if mf.GetName() != "kropath_cascade_mapfunc_errors_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			var gotFamily, gotTrigger string
			for _, lp := range m.GetLabel() {
				switch lp.GetName() {
				case "family":
					gotFamily = lp.GetValue()
				case "trigger":
					gotTrigger = lp.GetValue()
				}
			}
			if gotFamily == family && gotTrigger == trigger {
				return m.GetCounter().GetValue()
			}
		}
	}
	return 0
}

func TestRecordMapFuncListError(t *testing.T) {
	before := cascadeMapfuncErrorsTotalValue(t, "s3config", "kropathconfig")

	RecordMapFuncListError(logr.Discard(), "s3config", "kropathconfig", errors.New("boom"))

	after := cascadeMapfuncErrorsTotalValue(t, "s3config", "kropathconfig")
	if delta := after - before; delta != 1 {
		t.Errorf("kropath_cascade_mapfunc_errors_total{family=\"s3config\",trigger=\"kropathconfig\"} delta: want 1, got %v", delta)
	}
}
