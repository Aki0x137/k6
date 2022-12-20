package v1

import (
	"encoding/json"
	"net/http"
	"time"
)

func handleGetMetrics(cs *ControlSurface, rw http.ResponseWriter, _ *http.Request) {
	var t time.Duration
	if cs.Scheduler != nil {
		t = cs.Scheduler.GetState().GetCurrentTestRunDuration()
	}

	metrics := newMetricsJSONAPI(cs.MetricsEngine.ObservedMetrics(), t)

	data, err := json.Marshal(metrics)
	if err != nil {
		apiError(rw, "Encoding error", err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = rw.Write(data)
}

func handleGetMetric(cs *ControlSurface, rw http.ResponseWriter, _ *http.Request, id string) {
	var t time.Duration
	if cs.Scheduler != nil {
		t = cs.Scheduler.GetState().GetCurrentTestRunDuration()
	}

	metric, ok := cs.MetricsEngine.ObservedMetricByID(id)
	if !ok {
		apiError(rw, "Not Found", "No metric with that ID was found", http.StatusNotFound)
		return
	}
	wrappedMetric := newMetricEnvelope(&metric, t)

	data, err := json.Marshal(wrappedMetric)
	if err != nil {
		apiError(rw, "Encoding error", err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = rw.Write(data)
}
