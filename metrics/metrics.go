/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package metrics

import (
	"fmt"
	"time"

	"github.com/containerd/containerd/version"
	goMetrics "github.com/docker/go-metrics"
	"github.com/sirupsen/logrus"
)

type TimeMetric struct {
	start   map[string]time.Time
	elapsed map[string]time.Duration
}

var Timer *TimeMetric
var enable bool = false

func (timer *TimeMetric) StartTimer(name string) error {
	if !enable {
		return nil
	}
	if _, ok := timer.start[name]; ok {
		return fmt.Errorf("duplicated timer %s", name)
	}
	timer.start[name] = time.Now()
	return nil
}

func (timer *TimeMetric) FinishTimer(name string) error {
	if !enable {
		return nil
	}
	start, ok := timer.start[name]
	if !ok {
		return fmt.Errorf("%s timer does not start", name)
	}
	timer.elapsed[name] = time.Since(start)
	return nil
}

func (timer *TimeMetric) Report() {
	if !enable {
		return
	}
	defer func() {
		// clear all timer once report
		timer.start = make(map[string]time.Time)
		timer.elapsed = make(map[string]time.Duration)
	}()
	entry := logrus.WithField("timer", "dummy")

	for name := range timer.start {
		if _, ok := timer.elapsed[name]; !ok {
			entry.Errorf("timer-%s-not-finish", name)
			return
		}
	}

	for name, e := range timer.elapsed {
		entry = entry.WithField(name, e.String())
	}

	entry.Info("Containerd Time Metric")
}

func (timer *TimeMetric) Clean() {
	if !enable {
		return
	}
	timer.start = make(map[string]time.Time)
	timer.elapsed = make(map[string]time.Duration)
}

func init() {
	ns := goMetrics.NewNamespace("containerd", "", nil)
	c := ns.NewLabeledCounter("build_info", "containerd build information", "version", "revision")
	c.WithValues(version.Version, version.Revision).Inc()
	goMetrics.Register(ns)

	Timer = &TimeMetric{
		start:   make(map[string]time.Time),
		elapsed: make(map[string]time.Duration),
	}
}
