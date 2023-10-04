package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/storage"
)

type Handlers struct {
	storage *storage.Storage
}

func NewHandlers(storage *storage.Storage) *Handlers {
	return &Handlers{
		storage: storage,
	}
}

// Сервер должен принимать данные в формате:
// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, ErrMsgMethodOnlyPOST, http.StatusMethodNotAllowed)
		return
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/update")

	// head stands for "metric type", tail - "metric name"
	var head string
	head, r.URL.Path = ShiftPath(r.URL.Path)

	// check metric type
	if head == "" {
		http.Error(w, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}

	// check metric name
	if r.URL.Path == "/" {
		http.Error(w, ErrMsgEmptyMetricName, http.StatusNotFound)
		return
	}

	// check metric value
	_, metricValue := ShiftPath(r.URL.Path)
	metricValue = strings.TrimPrefix(metricValue, "/")
	if metricValue == "" {
		http.Error(w, ErrMsgWrongMetricValue, http.StatusBadRequest)
		return
	}

	// split handlers for [/gauge, /counter] endpoints
	switch head {
	case "gauge":
		h.updateGauge(w, r)
		return
	case "counter":
		h.updateCounter(w, r)
		return
	default:
		http.Error(w, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}
}

func (h *Handlers) updateGauge(w http.ResponseWriter, r *http.Request) {
	// At this point it is expected for the Path to be at least of 2 parts,
	// e.g. /gauge-name/gauge-value.
	//
	// Split in 3 parts so this last 3rd item will contain all possible
	// unnecessary trash and the 2nd will be nothing but the metric value
	paths := strings.SplitN(r.URL.Path[1:], "/", 3)

	val, err := strconv.ParseFloat(paths[1], 64)
	if err != nil {
		http.Error(w, ErrMsgWrongMetricValue, http.StatusBadRequest)
		return
	}

	h.storage.Gauges.Set(paths[0], model.Gauge(val))

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) updateCounter(w http.ResponseWriter, r *http.Request) {
	paths := strings.SplitN(r.URL.Path[1:], "/", 3)

	val, err := strconv.ParseInt(paths[1], 10, 64)
	if err != nil {
		http.Error(w, ErrMsgWrongMetricValue, http.StatusBadRequest)
		return
	}

	if val < 0 {
		http.Error(w, ErrMsgNegativeCounter, http.StatusBadRequest)
		return
	}

	h.storage.Counters.Set(paths[0], model.Counter(val))

	w.WriteHeader(http.StatusOK)
}

type metricsResponse struct {
	Gauges   map[string]model.Gauge   `json:"gauges"`
	Counters map[string]model.Counter `json:"counters"`
}

func (h *Handlers) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		// avoid cases when std mux drops in this handler when not needed
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, ErrMsgMethodOnlyGET, http.StatusMethodNotAllowed)
		return
	}

	var metrics metricsResponse

	metrics.Gauges = h.storage.Gauges.GetAll()
	metrics.Counters = h.storage.Counters.GetAll()

	resp, err := json.Marshal(metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = io.WriteString(w, string(resp))
}
