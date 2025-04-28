package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"alive": true}`)
}

func complexGet(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	category := r.PathValue("category")
	id := r.PathValue("id")

	io.WriteString(w, fmt.Sprintf("Tried to GET article with category %s and id %s", category, id))
}

// HTTP middleware setting a value on the request context
func myMiddleware(next http.Handler) http.Handler {
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving:", r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func corsMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("AllowedOrigins", "https://*,http://*")
		w.Header().Add("AllowedMethods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Add("AllowedHeaders", "Accept,Authorization,Content-Type,X-CSRF-Token")
		w.Header().Add("ExposedHeaders", "Link")
		next.ServeHTTP(w, r)
	})
}

func middleGet(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintf("The Middleware value is: %s", r.Context().Value("user")))
	// panic("ASDF")
}
