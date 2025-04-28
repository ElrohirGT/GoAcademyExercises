package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"alive": true}`)
}

func complexGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	category := chi.URLParam(r, "category")
	id := chi.URLParam(r, "id")

	io.WriteString(w, fmt.Sprintf("Tried to GET article with category %s and id %s", category, id))
}

// HTTP middleware setting a value on the request context
func MyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create new context from `r` request context, and assign key `"user"`
		// to value of `"123"`
		ctx := context.WithValue(r.Context(), "user", "123")

		// call the next handler in the chain, passing the response writer and
		// the updated request object with the new context value.
		//
		// note: context.Context values are nested, so any previously set
		// values will be accessible as well, and the new `"user"` key
		// will be accessible from this point forward.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func middleGet(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintf("The Middleware value is: %s", r.Context().Value("user")))
	// panic("ASDF")
}
