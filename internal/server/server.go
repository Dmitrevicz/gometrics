package server

import (
	"net/http"

	"github.com/Dmitrevicz/gometrics/internal/storage/memstorage"
	"github.com/gin-gonic/gin"
)

type server struct {
	router   http.Handler
	handlers *Handlers
}

func New() *server {
	storage := memstorage.New()
	s := server{
		// router:   http.NewServeMux(),
		handlers: NewHandlers(storage),
		// storage:  storage,
	}

	// configure router
	gin.SetMode(gin.ReleaseMode)    // make it not spam logs on startup
	r := gin.New()                  // no middlewares
	r.RedirectTrailingSlash = false // autotests fail without it
	r.Use(gin.Recovery())           // only panic recover for now
	r.Use(RequestLogger())          // custom logger middleware from the lesson

	// TODO: move routes configuration to separate func
	r.GET("/", s.handlers.PageIndex)
	r.GET("/all", s.handlers.GetAllMetrics)
	r.GET("/value/:type/:name", s.handlers.GetMetricByName)
	r.POST("/update/:type/:name/:value", s.handlers.Update)
	// For endpoint "/update/:type/:name/:value" decided to use readable params
	// definition. Because instead you have to use *wildcard like "update/:type/*params"
	// or smth like this if needed to treat params errors more precisely

	s.router = r

	return &s
}

// requires router of type *http.ServeMux
// func (s *server) configureRouter() {
// 	s.Router.HandleFunc(`/update/`, internal.UpdateHandler)
// }

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
