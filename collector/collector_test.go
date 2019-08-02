package collector

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
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
	//collected []*dto.Metric
	collected []*collectedMetric
}

type collectedMetric struct {
	desc   *prometheus.Desc
	metric *dto.Metric
}

func (c *collectedMetric) String() string {
	var labels string
	for _, l := range c.metric.Label {
		labels += fmt.Sprintf(" %s=%s,", *l.Name, *l.Value)
	}
	return fmt.Sprintf("Labels:%s Value: %f Desc: %s", labels, *c.metric.Gauge.Value, c.desc.String())
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
