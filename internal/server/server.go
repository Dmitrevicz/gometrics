package server

import (
	"fmt"
	"net/http"

	"github.com/Dmitrevicz/gometrics/internal/storage/memstorage"
	"github.com/gin-gonic/gin"
)

type server struct {
	router   http.Handler
	handlers *Handlers
	// storage  *Storage
}

func New() *server {
	storage := memstorage.New()
	s := server{
		// router:   http.NewServeMux(),
		handlers: NewHandlers(storage),
		// storage:  storage,
	}

	// configure router
	gin.SetMode(gin.ReleaseMode) // make it not spam logs on startup
	r := gin.New()               // no middlewares

	// r.RedirectFixedPath
	r.RedirectTrailingSlash = false
	fmt.Println("r.RedirectFixedPath:", r.RedirectFixedPath)
	fmt.Println("r.RedirectTrailingSlash:", r.RedirectTrailingSlash)

	r.Use(gin.Recovery()) // only panic recover for now

	// TODO: move to separate func
	r.GET("/", s.handlers.PageIndex) // remake to become html page
	r.GET("/all", s.handlers.GetAllMetrics)
	r.GET("/value/:type/:name", s.handlers.GetMetricByName)
	r.POST("/update/*smth", s.handlers.Update) // have to use "/*smth" wildcard to workaround proper path segments checks and responses
	// r.POST("update/:type/:name/:value", s.handlers.Update) // FIXME: 307 redirects "/name/" --> "/name"

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
