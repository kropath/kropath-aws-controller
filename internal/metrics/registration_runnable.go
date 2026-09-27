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

package metrics

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CollectorFactory constructs a scrape-time collector using the manager's
// cached reader (most kinds) or its uncached apiReader (kinds that must not
// block on a cold informer -- design §6, e.g. an absent <Family>Config CRD).
type CollectorFactory func(reader, apiReader client.Reader) prometheus.Collector

var (
	factoriesMu sync.Mutex
	factories   []CollectorFactory
)

// RegisterCollectorFactory adds a factory that collectorRunnable calls
// exactly once, after the manager's cache has synced, to build a feature
// package's scrape-time collector(s) (spec §3.6). Call this from the feature
// package that owns the collector -- e.g. from a package-level init() -- so
// that adding a new collector never requires editing collectorRunnable or
// cmd/manager/main.go (this is the self-registration seam I-2/I-3/I-4 build
// on).
func RegisterCollectorFactory(f CollectorFactory) {
	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	factories = append(factories, f)
}

// collectorRunnable registers every collector added via
// RegisterCollectorFactory against registerer once started, then blocks
// until ctx is done (spec §3.6, M-14).
type collectorRunnable struct {
	reader     client.Reader
	apiReader  client.Reader
	registerer prometheus.Registerer
}

// NewCollectorRunnable constructs the registration runnable. reader is
// typically the manager's cache (mgr.GetCache()); apiReader is typically the
// manager's uncached reader (mgr.GetAPIReader()); registerer is typically
// ctrlmetrics.Registry. Add the result via mgr.Add beside the CRD watcher
// (cmd/manager/main.go).
func NewCollectorRunnable(reader, apiReader client.Reader, registerer prometheus.Registerer) *collectorRunnable {
	return &collectorRunnable{reader: reader, apiReader: apiReader, registerer: registerer}
}

// Start registers every collector factory added via RegisterCollectorFactory,
// then blocks until ctx is done.
func (c *collectorRunnable) Start(ctx context.Context) error {
	factoriesMu.Lock()
	fs := make([]CollectorFactory, len(factories))
	copy(fs, factories)
	factoriesMu.Unlock()

	for _, f := range fs {
		if err := c.registerer.Register(f(c.reader, c.apiReader)); err != nil {
			return err
		}
	}

	<-ctx.Done()
	return nil
}

// NeedLeaderElection reports false so this runnable starts in the manager's
// Others group -- after cache sync, on every replica, not only the leader
// (M-14). Every replica must expose the same collector series for the design
// §9 max-by aggregation contract to hold; this is the repo's first
// non-leader mgr.Add runnable (the CRD watcher returns true and runs on the
// leader alone -- precedent for the registration mechanism only, never for
// this group).
func (c *collectorRunnable) NeedLeaderElection() bool {
	return false
}
