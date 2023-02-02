package v1

import (
	"bytes"
	"encoding/json"
	"time"

	"gopkg.in/guregu/null.v3"

	"go.k6.io/k6/metrics"
)

type NullMetricType struct {
	Type  metrics.MetricType
	Valid bool
}

func (t NullMetricType) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return t.Type.MarshalJSON()
}

func (t *NullMetricType) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		t.Valid = false
		return nil
	}
	t.Valid = true
	return json.Unmarshal(data, &t.Type)
}

type NullValueType struct {
	Type  metrics.ValueType
	Valid bool
}

func (t NullValueType) MarshalJSON() ([]byte, error) {
	if !t.Valid {
		return []byte("null"), nil
	}
	return t.Type.MarshalJSON()
}

func (t *NullValueType) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		t.Valid = false
		return nil
	}
	t.Valid = true
	return json.Unmarshal(data, &t.Type)
}

type Metric struct {
	Name string `json:"-" yaml:"name"`

	Type     NullMetricType `json:"type" yaml:"type"`
	Contains NullValueType  `json:"contains" yaml:"contains"`
	Tainted  null.Bool      `json:"tainted" yaml:"tainted"`

	Sample map[string]float64 `json:"sample" yaml:"sample"`
}

// NewMetric constructs a new Metric
func NewMetric(om metrics.ObservedMetric, t time.Duration) Metric {
	m := Metric{
		Name:     om.Name,
		Type:     NullMetricType{om.Type, true},
		Contains: NullValueType{om.Contains, true},
		Tainted:  om.Tainted,
	}

	switch sink := om.Sink.(type) {
	case *metrics.CounterSink:
		m.Sample = map[string]float64{
			"count": sink.Value,
			"rate":  sink.Rate(t),
		}
	case *metrics.GaugeSink:
		m.Sample = map[string]float64{"value": sink.Value}
	case *metrics.RateSink:
		m.Sample = map[string]float64{"rate": sink.Rate()}
	case *metrics.TrendSink:
		m.Sample = map[string]float64{
			"min":   sink.Min,
			"max":   sink.Max,
			"avg":   sink.Avg,
			"med":   sink.P(0.5),
			"p(90)": sink.P(0.90),
			"p(95)": sink.P(0.95),
		}
	}

	return m
}
