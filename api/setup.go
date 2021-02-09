package api

import (
	"compress/flate"
	"github.com/evilsocket/shieldwall/mailer"
	"net/http"

	"github.com/evilsocket/islazy/log"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

type API struct {
	config   Config
	mail     EmailConfig
	sendmail *mailer.Mailer
	router   *chi.Mux
}

func Setup(config Config, email EmailConfig, sendmail *mailer.Mailer) *API {
	api := &API{
		config:   config,
		mail:     email,
		sendmail: sendmail,
		router:   chi.NewRouter(),
	}

	compressor := middleware.NewCompressor(flate.DefaultCompression)

	api.router.Use(compressor.Handler)

	api.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	api.router.Route("/api", func(r chi.Router) {
		r.Route("/v1", func(r chi.Router) {
			r.Get("/rules", api.GetRules)

			r.Route("/user", func(r chi.Router) {
				r.Post("/register", api.UserRegister)
				r.Get("/verify/{verification:[A-Fa-f0-9]{64}}", api.UserVerify)
				r.Post("/login", api.UserLogin)
			})
		})
	})

	return api
}

func (api *API) Run() {
	log.Info("api starting on %s", api.config.Address)
	log.Fatal("%v", http.ListenAndServe(api.config.Address, api.router))
}