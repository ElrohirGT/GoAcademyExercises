package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	shared "github.com/ElrohirGT/GoAcademyExercises/ConcurrencyExercise/Shared"
)

type Params struct {
	SourcePath      string
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

func (self *Server) MountHandlers(db *[]shared.APIUser) {
	self.Router.HandleFunc("/users", get_users(db))
}

func ParseParams() Params {
	params := Params{}

	flag.DurationVar(&params.GracefulTimeout, "graceful-timeout", time.Second*10, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	params.SourcePath = os.Getenv("SOURCE_PATH")
	if params.SourcePath == "" {
		params.SourcePath = "./source.json"
	}

	return params
}

func main() {
	params := ParseParams()

	log.Printf("Attempting to read source in: %s", params.SourcePath)
	sourceBytes, err := os.ReadFile(params.SourcePath)
	if err != nil {
		panic("Can't read the source file")
	}

	log.Printf("Parsing source file...")
	var resp shared.APIResponse
	err = json.Unmarshal(sourceBytes, &resp)
	if err != nil {
		panic("Can't parse response in source file")
	}

	server := CreateNewServer()
	server.MountHandlers(&resp.Results)

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
