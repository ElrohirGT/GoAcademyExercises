package main

import (
	"log"
	"net/http"
)

type Server struct {
	Router *http.ServeMux
}

func CreateNewServer() *Server {
	s := &Server{}
	s.Router = http.NewServeMux()
	return s
}

type HTTPMiddleware = func(http.Handler) http.Handler

func applyMiddleware(middlewares []HTTPMiddleware, handler func(http.ResponseWriter, *http.Request)) http.Handler {
	var accum http.Handler = http.HandlerFunc(handler)

	// Apply middleware from last to first
	for i := range len(middlewares) {
		idx := len(middlewares) - 1 - i
		fun := middlewares[idx]
		accum = fun(accum)
	}

	return accum
}

func (s *Server) MountHandlers() {
	r := s.Router
	BASE_MIDDLEWARE := []HTTPMiddleware{
		loggingMiddleware,
		myMiddleware,
		corsMiddlware,
	}

	r.Handle("/health", applyMiddleware(BASE_MIDDLEWARE, healthCheck))
	r.Handle("/article/{category}/{id}", applyMiddleware(BASE_MIDDLEWARE, complexGet))
	r.Handle("/middle", applyMiddleware(BASE_MIDDLEWARE, middleGet))

	notFoundHandler := func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		w.Write([]byte("CUSTOM 404 NOT FOUND!"))
	}
	r.Handle("/*", http.HandlerFunc(notFoundHandler))
}
