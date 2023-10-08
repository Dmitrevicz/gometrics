package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
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
	/* method check is handled by gin now
	if c.Request.Method != http.MethodPost {
		http.Error(c.Writer, ErrMsgMethodOnlyPOST, http.StatusMethodNotAllowed)
		return
	} */

	// It just works but might refactor later.
	// Handler can have much less code as we now use gin as router.
	// TODO: refactor "/update" handler

	fmt.Println("params:", c.Params)
	fmt.Println("r.URL.Path:", c.Request.URL.Path)
	c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/update")

	// head stands for "metric type", tail - "metric name"
	var head string
	head, c.Request.URL.Path = ShiftPath(c.Request.URL.Path)

	// check metric type
	if head == "" {
		http.Error(c.Writer, ErrMsgWrongMetricType, http.StatusBadRequest)
		return
	}

	metricName, metricValue := ShiftPath(c.Request.URL.Path)
	metricName = strings.TrimSpace(metricName)

	// check metric name
	if c.Request.URL.Path == "/" || metricName == "" {
		http.Error(c.Writer, ErrMsgEmptyMetricName, http.StatusNotFound)
		return
	}

	// check metric value
	metricValue = strings.TrimPrefix(metricValue, "/")
	if metricValue == "" {
		http.Error(c.Writer, ErrMsgWrongMetricValue, http.StatusBadRequest)
		return
	}

	// split handlers for [/gauge, /counter] endpoints
	switch head {
	case "gauge":
		h.updateGauge(c.Writer, c.Request)
		return
	case "counter":
		h.updateCounter(c.Writer, c.Request)
		return
	default:
		http.Error(c.Writer, ErrMsgWrongMetricType, http.StatusBadRequest)
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

	h.storage.Gauges().Set(paths[0], model.Gauge(val))

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

	h.storage.Counters().Set(paths[0], model.Counter(val))

	w.WriteHeader(http.StatusOK)
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
	/* deprecated as gin is used now
	if c.Request.URL.Path != "/" {
		// avoid cases when std mux drops in this handler when not needed
		http.NotFound(c.Writer, c.Request)
		return
	} */

	if c.Request.Method != http.MethodGet {
		http.Error(c.Writer, ErrMsgMethodOnlyGET, http.StatusMethodNotAllowed)
		return
	}

	var metrics metricsResponse

	metrics.Gauges = h.storage.Gauges().GetAll()
	metrics.Counters = h.storage.Counters().GetAll()

	resp, err := json.Marshal(metrics)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = io.WriteString(c.Writer, string(resp))
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
