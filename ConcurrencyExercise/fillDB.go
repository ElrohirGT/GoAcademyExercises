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
)

type APIUserName struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

type APIStreetLocation struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
}
type APILocation struct {
	Street  APIStreetLocation `json:"street"`
	City    string            `json:"city"`
	State   string            `json:"state"`
	Country string            `json:"country"`
}

type APIUser struct {
	Gender   string      `json:"gender"`
	Name     APIUserName `json:"name"`
	Location APILocation `json:"location"`
	Email    string      `json:"email"`
	Phone    string      `json:"phone"`
}

type APIResponse struct {
	Results []APIUser      `json:"results"`
	Info    map[string]any `json:"info"`
}

type APIError struct {
	Error string `json:"error"`
}

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

func NewUserFromAPI(u APIUser, info map[string]any) User {
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
const USER_PER_REQUEST = 500

var DB = make([]User, 0, TARGET_USER_COUNT)

func FillDBData(APIUrl string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		reqUrl, err := url.Parse(APIUrl)
		if err != nil {
			panic("Failed to parse APIUrl!")
		}
		reqUrl.Query().Set("results", strconv.FormatInt(USER_PER_REQUEST, 10))
		reqUrlString := reqUrl.String()

		outputChannel := make(chan User, TARGET_USER_COUNT/WORKER_COUNT)

		makeReqSignal := make(chan bool, WORKER_COUNT*2)
		// Signal work thread
		go func() {
			defer close(makeReqSignal)

			for range TARGET_USER_COUNT / USER_PER_REQUEST {
				select {
				case <-r.Context().Done():
					return
				case makeReqSignal <- true:
				}
			}
		}()

		// Append to DB
		appenderGroup := sync.WaitGroup{}
		appenderGroup.Add(1)
		go func() {
			defer appenderGroup.Done()
			for user := range outputChannel {
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

				for {
					select {
					case <-r.Context().Done():
						log.Println("Stopping worker due to context being canceled...")
						return
					case shouldKeep := <-makeReqSignal:
						if !shouldKeep {
							break // Exit processing loop once makeReqSignal closes
						}

						var err error = errors.New("Dummy error for execution purposes")

						var resp *http.Response
						var req *http.Request
						var respBytes []byte
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
									continue
								}

								resp, err = http.DefaultClient.Do(req)
								if err != nil {
									log.Println("Error: Doing GET request")
									continue
								}

								respBytes, err = io.ReadAll(resp.Body)
								if err != nil {
									log.Println("Error: Reading HTTP body")
									continue
								}

								var apiResponse APIResponse
								err = json.Unmarshal(respBytes, &apiResponse)
								if err != nil {
									log.Printf("Error: Unmarshalling HTTP response body, maybe it was an error? (%s)", err)
									log.Printf("Body: %s", string(respBytes))

									var apiResponse APIError
									err = json.Unmarshal(respBytes, &apiResponse)
									if err != nil {
										log.Println("Error: Failed even to parse as an error!")
										continue
									}

									continue
								}

								if len(apiResponse.Results) != TARGET_USER_COUNT {
									log.Println("Error: Not enough users returned by the API: %d", len(apiResponse.Results))
									err = errors.New("No results from the API")
									continue
								}

								for _, v := range apiResponse.Results {
									user := NewUserFromAPI(v, apiResponse.Info)
									outputChannel <- user
								}
							}
						}
					}
				}
			}()
		}

		log.Println("Waiting for workers...")
		workersGroup.Wait()
		close(outputChannel)

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
