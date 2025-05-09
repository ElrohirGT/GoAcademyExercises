package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	summary "github.com/ElrohirGT/GoAcademyExercises/FinalTest/routes/GetSummary"
	ready "github.com/ElrohirGT/GoAcademyExercises/FinalTest/routes/Ready"
)

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/demographic-summary", summary.GetSummary)
	router.HandleFunc("/ready", ready.Ready)

	srv := &http.Server{
		Addr:         "0.0.0.0:4040",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		fmt.Println("Listening on: ", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	wg.Wait()
}
