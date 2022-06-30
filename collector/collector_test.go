// Copyright 2019-2022 The sakuracloud_exporter Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
)

var logbuf *bytes.Buffer
var testLogger log.Logger
var testErrors *prometheus.CounterVec

func collectDescs(collector prometheus.Collector) []*prometheus.Desc {
	initLoggerAndErrors()

	ch := make(chan *prometheus.Desc)
	go func() {
		collector.Describe(ch)
		close(ch)
	}()

	var res []*prometheus.Desc
	for desc := range ch {
		res = append(res, desc)
	}
	return res
}

type collectResult struct {
	logged []string
	errors *dto.Metric
	// collected []*dto.Metric
	collected []*collectedMetric
}

type collectedMetric struct {
	desc   *prometheus.Desc
	metric *dto.Metric
}

func (c *collectedMetric) String() string {
	labels := labelToString(c.metric.Label)
	return fmt.Sprintf("Labels:%s Value: %f Desc: %s", labels, *c.metric.Gauge.Value, c.desc.String())
}

func labelToString(label []*dto.LabelPair) string {
	var labels string
	for _, l := range label {
		labels += fmt.Sprintf(" %s=%s,", *l.Name, *l.Value)
	}
	return labels
}

func requireMetricsEqual(t *testing.T, m1, m2 []*collectedMetric) {
	sort.Slice(m1, func(i, j int) bool {
		s1 := m1[i].desc.String()
		s2 := m1[j].desc.String()
		if s1 != s2 {
			return s1 < s2
		}
		return labelToString(m1[i].metric.Label) < labelToString(m1[j].metric.Label)
	})
	sort.Slice(m2, func(i, j int) bool {
		s1 := m2[i].desc.String()
		s2 := m2[j].desc.String()
		if s1 != s2 {
			return s1 < s2
		}
		return labelToString(m2[i].metric.Label) < labelToString(m2[j].metric.Label)
	})
	require.Equal(t, m1, m2)
}

func initLoggerAndErrors() {
	logbuf = &bytes.Buffer{}
	testLogger = log.NewLogfmtLogger(log.NewSyncWriter(logbuf))
	testErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "sakuracloud_exporter_errors_total",
	}, []string{"collector"})
}

func collectMetrics(collector prometheus.Collector, errLabel string) (*collectResult, error) {
	ch := make(chan prometheus.Metric)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	var metrics []*collectedMetric
	for metric := range ch {
		v := &dto.Metric{}
		err := metric.Write(v)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, &collectedMetric{
			desc:   metric.Desc(),
			metric: v,
		})
	}

	errs := &dto.Metric{}
	if err := testErrors.WithLabelValues(errLabel).Write(errs); err != nil {
		return nil, err
	}

	logs := strings.Split(logbuf.String(), "\n")
	sort.Strings(logs)
	var trimed []string
	for _, l := range logs {
		if l != "" {
			trimed = append(trimed, l)
		}
	}

	return &collectResult{
		logged:    trimed,
		errors:    errs,
		collected: metrics,
	}, nil
}

func createGaugeMetric(value float64, labels map[string]string) *dto.Metric {
	metric := &dto.Metric{
		Gauge: &dto.Gauge{
			Value: &value,
		},
	}

	var keys []string
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i := range keys {
		key := keys[i]
		value := labels[key]
		metric.Label = append(metric.Label, &dto.LabelPair{
			Name:  &key,
			Value: &value,
		})
	}

	return metric
}

func createGaugeWithTimestamp(value float64, labels map[string]string, timestamp time.Time) *dto.Metric {
	metric := createGaugeMetric(value, labels)
	ts := timestamp.Unix() * 1000
	metric.TimestampMs = &ts
	return metric
}
