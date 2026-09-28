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
	"github.com/go-logr/logr"
	"github.com/kropath/kropath-controller/internal/metrics"
)

// RecordMapFuncListError logs and counts a watch map function's List failure
// (spec §2.7). Every cascade reconciler's three map functions call this
// instead of logging inline, so a List failure -- which drops the whole batch
// of enqueue requests for that watch event -- is visible as
// kropath_cascade_mapfunc_errors_total{family,trigger} instead of only a log
// line. trigger is one of "kropathconfig", "familyconfig", "namespace".
func RecordMapFuncListError(log logr.Logger, family, trigger string, err error) {
	log.Error(err, "unable to list configs", "family", family, "trigger", trigger)
	metrics.CascadeMapFuncError(family, trigger)
}
