package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Params struct {
	APIUrl          string
	GracefulTimeout time.Duration
}

type Server struct {
	Router *http.ServeMux
}

func CreateNewServer() *Server {
	return &Server{
		Router: http.NewServeMux(),
	}
}

func (self *Server) MountHandlers(params *Params) {
	self.Router.HandleFunc("/fillDB", FillDBData(params.APIUrl))
}

func ParseParams() Params {
	params := Params{}

	flag.DurationVar(&params.GracefulTimeout, "graceful-timeout", time.Second*10, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	params.APIUrl = os.Getenv("API_URL")
	if params.APIUrl == "" {
		params.APIUrl = "https://randomuser.me/api/"
	}

	return params
}

func main() {
	params := ParseParams()

	server := CreateNewServer()
	server.MountHandlers(&params)

	srv := &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      server.Router,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Println("Listening on:", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), params.GracefulTimeout)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := srv.Shutdown(ctx); err != nil {
		log.Panic("Failed to shutdown the server: ", err)
	}
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)

}
