package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

func dateExtractMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		date := chi.URLParam(r, "date")
		ctx := context.WithValue(r.Context(), "date", date)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func languageChecker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lang := strings.ToLower(chi.URLParam(r, "lang"))
		switch lang {
		case "en", "fr", "de", "es", "pt":
			ctx := context.WithValue(r.Context(), "lang", lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}

func useCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		ctx := context.WithValue(r.Context(), "cors", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Default().Handler)
	r.Use(middleware.Timeout(10 * time.Second))

	dofusdudeApiMajor := 1

	r.With(useCors).With(languageChecker).Route(fmt.Sprintf("/dofus3/v%d", dofusdudeApiMajor), func(r chi.Router) {
		r.Route("/meta/{lang}/almanax/bonuses", func(r chi.Router) {
			r.Get("/", ListBonuses)
			r.Get("/search", SearchBonuses)
		})

		r.Route("/{lang}/almanax", func(r chi.Router) {
			r.Get("/", RetrieveAlmanax)
			r.With(languageChecker).Put("/{lang}", UpdateAlmanax)
		})
	})

	return r
}
