package server

import (
	"fmt"
	"net/http"

	"github.com/chaikadn/url-shortener/internal/config"
	"github.com/chaikadn/url-shortener/internal/handler"
	"github.com/chaikadn/url-shortener/internal/middleware"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	log *zap.Logger
	srv *http.Server
}

func New(log *zap.Logger, cfg *config.Config, hnd *handler.Handler) *Server {
	return &Server{
		log: log,
		srv: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Handler: registerRoutes(log, cfg, hnd),
		},
	}
}

func (s *Server) Run() error {
	s.log.Info("Starting server", zap.String("addr", s.srv.Addr))
	return s.srv.ListenAndServe()
}

// func (s *Server) Shutdown(ctx context.Context) error {
//     return s.srv.Shutdown(ctx)
// }

func registerRoutes(log *zap.Logger, cfg *config.Config, hnd *handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logging(log))
	r.Use(middleware.Compression(log))

	r.Group(func(r chi.Router) {
		r.Get("/ping", hnd.HandlePingStorage)
		r.Get("/{key}", hnd.HandleRedirect)
	})

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/login", hnd.HandleLogin)
			r.Post("/register", hnd.HandleRegister)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(cfg.JWTsecret))

			r.Group(func(r chi.Router) {
				r.Post("/shorten", hnd.HandleShorten)
				r.Get("/urls", hnd.HandleUserURLs)
			})

			// Админские роуты
			// r.Group(func(r chi.Router) {
			// 	r.Use(middleware.RequireRole("admin"))
			// 	r.Get("/admin/stats", hnd.HandleStats)
			// })
		})
	})

	return r
}
