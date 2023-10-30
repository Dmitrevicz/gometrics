package server

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/Dmitrevicz/gometrics/internal/model"
	"github.com/Dmitrevicz/gometrics/internal/storage"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	storage storage.Storage
}

func NewHandlers(storage storage.Storage) *Handlers {
	return &Handlers{
		storage: storage,
	}
}

// Update
//
// > Сервер должен принимать данные в формате:
// http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
func (h *Handlers) Update(c *gin.Context) {
	mType, mName, mValue := c.Param("type"), c.Param("name"), c.Param("value")
	mType = strings.TrimSpace(mType)
	mName = strings.TrimSpace(mName)
	mValue = strings.TrimSpace(mValue)

	if mType == "" {
		http.Error(c.Writer, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}

	if mName == "" {
		http.Error(c.Writer, ErrMsgEmptyMetricName, http.StatusNotFound)
		return
	}

	if mValue == "" {
		http.Error(c.Writer, ErrMsgWrongMetricValue, http.StatusBadRequest)
		return
	}

	// split handlers for [/gauge, /counter] endpoints
	switch mType {
	case model.MetricTypeGauge:
		h.updateGauge(c, mName, mValue)
		return
	case model.MetricTypeCounter:
		h.updateCounter(c, mName, mValue)
		return
	default:
		http.Error(c.Writer, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}
}

func (h *Handlers) updateGauge(c *gin.Context, name, value string) {
	var gauge model.Gauge
	gauge, err := gauge.FromString(value)
	if err != nil {
		http.Error(c.Writer, ErrMsgWrongMetricValue, http.StatusBadRequest)
		return
	}

	h.storage.Gauges().Set(name, gauge)

	c.Status(http.StatusOK)
}

func (h *Handlers) updateCounter(c *gin.Context, name, value string) {
	var counter model.Counter
	counter, err := counter.FromString(value)
	if err != nil {
		http.Error(c.Writer, ErrMsgWrongMetricValue, http.StatusBadRequest)
		return
	}

	if counter < 0 {
		http.Error(c.Writer, ErrMsgNegativeCounter, http.StatusBadRequest)
		return
	}

	h.storage.Counters().Set(name, counter)

	c.Status(http.StatusOK)
}

// UpdateMetricByJSON
//
// > Для передачи метрик на сервер используйте Content-Type: application/json.
// В теле запроса должен быть описанный выше JSON. Передавать метрики нужно через
// POST update/. В теле ответа отправляйте JSON той же структуры с актуальным
// (изменённым) значением Value.
func (h *Handlers) UpdateMetricByJSON(c *gin.Context) {
	var req model.Metrics

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest) // StatusBadRequest or StatusInternalServerError when parsing json?
		return
	}

	req.ID = strings.TrimSpace(req.ID)
	req.MType = strings.TrimSpace(req.MType)

	if req.MType == "" {
		http.Error(c.Writer, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(c.Writer, ErrMsgEmptyMetricName, http.StatusNotFound)
		return
	}

	switch req.MType {
	case model.MetricTypeGauge:
		h.updateGaugeFromMetrics(c, req)
		return
	case model.MetricTypeCounter:
		h.updateCounterFromMetrics(c, req)
		return
	default:
		http.Error(c.Writer, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}
}

func (h *Handlers) updateGaugeFromMetrics(c *gin.Context, m model.Metrics) {
	if m.Value == nil {
		http.Error(c.Writer, ErrMsgWrongMetricValue, http.StatusNotFound)
		return
	}

	h.storage.Gauges().Set(m.ID, model.Gauge(*m.Value))
	m.Delta = nil

	c.JSON(http.StatusOK, m)
}

func (h *Handlers) updateCounterFromMetrics(c *gin.Context, m model.Metrics) {
	if m.Delta == nil {
		http.Error(c.Writer, ErrMsgWrongMetricValue, http.StatusNotFound)
		return
	}

	if *m.Delta < 0 {
		http.Error(c.Writer, ErrMsgNegativeCounter, http.StatusBadRequest)
		return
	}

	h.storage.Counters().Set(m.ID, model.Counter(*m.Delta))

	counter, ok := h.storage.Counters().Get(m.ID)
	if !ok {
		http.Error(c.Writer, ErrMsgNothingFound, http.StatusNotFound)
		return
	}

	f := float64(counter)
	m.Value = &f

	c.JSON(http.StatusOK, m)
}

// GetMetricByName returns metric value by its name
//
// > Доработайте сервер так, чтобы в ответ на запрос
// GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
// он возвращал текущее значение метрики в текстовом виде со статусом http.StatusOK.
// При попытке запроса неизвестной метрики сервер должен возвращать http.StatusNotFound.
func (h *Handlers) GetMetricByName(c *gin.Context) {
	mType, mName := c.Param("type"), c.Param("name")
	mType = strings.TrimSpace(mType)
	mName = strings.TrimSpace(mName)

	if mName == "" {
		http.Error(c.Writer, ErrMsgEmptyMetricName, http.StatusBadRequest)
		return
	}

	var (
		value interface{}
		ok    bool
	)

	switch mType {
	case model.MetricTypeGauge:
		value, ok = h.storage.Gauges().Get(mName)
	case model.MetricTypeCounter:
		value, ok = h.storage.Counters().Get(mName)
	default:
		http.Error(c.Writer, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}

	if !ok {
		http.Error(c.Writer, ErrMsgNothingFound, http.StatusNotFound)
		return
	}

	c.String(http.StatusOK, "%v", value)
}

// GetMetricByJSON returns metric value by name and type provided in json body.
//
// > Для получения метрик с сервера используйте Content-Type: application/json.
// В теле запроса должен быть описанный выше JSON с заполненными полями ID и
// MType. Запрашивать нужно через POST value/. В теле ответа должен приходить
// такой же JSON, но с уже заполненными значениями метрик.
func (h *Handlers) GetMetricByJSON(c *gin.Context) {
	var req model.Metrics

	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadRequest) // StatusBadRequest or StatusInternalServerError when parsing json?
		return
	}

	// sanitize inputs
	req.ID = strings.TrimSpace(req.ID)
	req.MType = strings.TrimSpace(req.MType)
	req.Delta = nil
	req.Value = nil

	if req.ID == "" {
		http.Error(c.Writer, ErrMsgEmptyMetricName, http.StatusBadRequest)
		return
	}

	var (
		value interface{}
		ok    bool
	)

	switch req.MType {
	case model.MetricTypeGauge:
		value, ok = h.storage.Gauges().Get(req.ID)
		f := float64(value.(model.Gauge))
		req.Value = &f
	case model.MetricTypeCounter:
		value, ok = h.storage.Counters().Get(req.ID)
		d := int64(value.(model.Counter))
		req.Delta = &d // автотесты требуют, чтобы counter отдавался в .Delta
	default:
		http.Error(c.Writer, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}

	if !ok {
		http.Error(c.Writer, ErrMsgNothingFound, http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, req)
}

type metricsResponse struct {
	Gauges   map[string]model.Gauge   `json:"gauges"`
	Counters map[string]model.Counter `json:"counters"`
}

// GetAllMetrics - just for debugging, returns list of all metrics.
func (h *Handlers) GetAllMetrics(c *gin.Context) {
	var metrics metricsResponse

	metrics.Gauges = h.storage.Gauges().GetAll()
	metrics.Counters = h.storage.Counters().GetAll()

	resp, err := json.Marshal(metrics)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = c.Writer.Write(resp)
}

// TODO: might move to file
var pageTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>gometrics</title>
</head>
	<h1>gometrics</h1>
	<hr>

	<h2>Counters</h2>
	<ul>
	{{range $key, $value := .Counters}}
		<li>{{$key}}: {{$value}}</li>
	{{end}}
	</ul>

	<h2>Gauges</h2>
	<ul>
	{{range $key, $value := .Gauges}}
		<li>{{$key}}: {{$value}}</li>
	{{end}}
	</ul>
</body>
</html>
`))

type indexPageData struct {
	Gauges   map[string]model.Gauge
	Counters map[string]model.Counter
}

func (h *Handlers) PageIndex(c *gin.Context) {
	var pData indexPageData

	pData.Gauges = h.storage.Gauges().GetAll()
	pData.Counters = h.storage.Counters().GetAll()

	if err := pageTmpl.Execute(c.Writer, pData); err != nil {
		http.Error(c.Writer, ErrMsgTemplateExec+": "+err.Error(), http.StatusInternalServerError)
	}
}
