package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/ElrohirGT/GoAcademyExercises/ConcurrencyExercise/Shared"
)

type User struct {
	Id       uint
	Gender   string
	Name     string // -> first and last
	Location string // -> street -> number and name
	City     string
	State    string
	Country  string
	Email    string
	Phone    string

	Info map[string]any
}

func NewUserFromAPI(u shared.APIUser, info map[string]any) User {
	return User{
		Gender:   u.Gender,
		Name:     fmt.Sprintf("%s %s", u.Name.First, u.Name.Last),
		Location: fmt.Sprintf("%d %s", u.Location.Street.Number, u.Location.Street.Name),
		City:     u.Location.City,
		State:    u.Location.State,
		Country:  u.Location.Country,
		Email:    u.Email,
		Phone:    u.Phone,
		Info:     info,
	}
}

const TARGET_USER_COUNT = 10_000
const WORKER_COUNT = 50
const USERS_PER_REQUEST = 500

var DB = make([]User, 0, TARGET_USER_COUNT)

func FillDBData(APIUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqUrl, err := url.Parse(APIUrl)
		if err != nil {
			panic("Failed to parse APIUrl!")
		}

		q := reqUrl.Query()
		q.Set("results", strconv.FormatInt(USERS_PER_REQUEST, 10))

		reqUrl.RawQuery = q.Encode()
		reqUrlString := reqUrl.String()
		log.Printf("Target of GET requests: %s", reqUrlString)

		processedUsers := make(chan User, USERS_PER_REQUEST*2)

		workTickets := make(chan bool, WORKER_COUNT*2)
		// Signal work thread
		go func() {
			defer close(workTickets)

			for range TARGET_USER_COUNT / USERS_PER_REQUEST {
				select {
				case <-r.Context().Done():
					return
				case workTickets <- true:
				}
			}
		}()

		// Append to DB
		appenderGroup := sync.WaitGroup{}
		appenderGroup.Add(1)
		go func() {
			defer appenderGroup.Done()
			for user := range processedUsers {
				user.Id = uint(len(DB))
				DB = append(DB, user)
				log.Printf("Adding user with ID %d: %v", user.Id, user)
			}
		}()

		workersGroup := sync.WaitGroup{}
		for i := range WORKER_COUNT {
			workersGroup.Add(1)
			go func() {
				defer func() {
					log.Println("Worker ", i, " Done!")
					workersGroup.Done()
				}()

			workerLoop:
				for {
					select {
					case <-r.Context().Done():
						log.Println("Stopping worker due to context being canceled...")
						return
					case shouldKeep := <-workTickets:
						if !shouldKeep {
							break workerLoop // Exit processing loop once makeReqSignal closes
						}

						var err error = errors.New("Dummy error for execution purposes")

						var resp *http.Response
						var req *http.Request
						var respBytes []byte
					requestLoop:
						for err != nil {
							err = nil

							select {
							case <-r.Context().Done():
								log.Println("Stopping worker due to context being canceled...")
								return
							default:

								req, err = http.NewRequestWithContext(r.Context(), http.MethodGet, reqUrlString, nil)
								if err != nil {
									log.Println("Error: Creating GET request")
									continue requestLoop
								}

								resp, err = http.DefaultClient.Do(req)
								if err != nil {
									log.Printf("Error: Doing GET request. `%s`", err)
									continue requestLoop
								}

								respBytes, err = io.ReadAll(resp.Body)
								if err != nil {
									log.Println("Error: Reading HTTP body")
									continue requestLoop
								}

								var apiResponse shared.APIResponse
								err = json.Unmarshal(respBytes, &apiResponse)
								if err != nil {
									log.Printf("Error: Unmarshalling HTTP response body, maybe it was an error? (%s)", err)
									log.Printf("Body: %s", string(respBytes))

									var apiResponse shared.APIError
									err = json.Unmarshal(respBytes, &apiResponse)
									if err != nil {
										log.Println("Error: Failed even to parse as an error!")
										continue requestLoop
									}

									continue requestLoop
								}

								if len(apiResponse.Results) != USERS_PER_REQUEST {
									log.Printf("Error: Not enough users returned by the API: %d != %d", len(apiResponse.Results), TARGET_USER_COUNT)
									err = errors.New("No results from the API")
									continue requestLoop
								}

								for _, v := range apiResponse.Results {
									user := NewUserFromAPI(v, apiResponse.Info)
									processedUsers <- user
								}
							}
						}
					}
				}
			}()
		}

		log.Println("Waiting for workers...")
		workersGroup.Wait()
		close(processedUsers)

		log.Println("Waiting for appenders...")
		appenderGroup.Wait()

		log.Println("Marshalling JSON response...")
		respBytes, err := json.Marshal(&DB)
		if err != nil {
			panic("Can't marshal response!")
		}
		log.Println("Sending JSON response!")
		w.Header().Add("Content-Type", "application/json")
		w.Write(respBytes)
	}
}
