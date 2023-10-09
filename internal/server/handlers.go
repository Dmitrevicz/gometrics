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
	case "gauge":
		h.updateGauge(c, mName, mValue)
		return
	case "counter":
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
	case "gauge":
		value, ok = h.storage.Gauges().Get(mName)
	case "counter":
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

type metricsResponse struct {
	Gauges   map[string]model.Gauge   `json:"gauges"`
	Counters map[string]model.Counter `json:"counters"`
}

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
