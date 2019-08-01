package collector

import (
	"bytes"
	"sort"
	"strings"

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
	logged    string
	errors    *dto.Metric
	collected []*dto.Metric
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

	var metrics []*dto.Metric
	for metric := range ch {
		v := &dto.Metric{}
		err := metric.Write(v)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, v)
	}

	errs := &dto.Metric{}
	if err := testErrors.WithLabelValues(errLabel).Write(errs); err != nil {
		return nil, err
	}

	return &collectResult{
		logged:    strings.Trim(logbuf.String(), "\n"),
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
