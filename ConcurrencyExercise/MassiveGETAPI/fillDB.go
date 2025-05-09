package main

import (
	"context"
	"database/sql"
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
	"github.com/lib/pq"
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
const RETRY_LIMIT = 5

var RETRY_LIMIT_EXCEEDED_ERR error = errors.New("Retry limit exceeded")

var DB = make([]User, 0, TARGET_USER_COUNT)

func FillDBData(APIUrl string, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		reqContext := r.Context()

		var reqUrl *url.URL
		reqUrl, err = url.Parse(APIUrl)
		if err != nil {
			panic("Failed to parse APIUrl!")
		}

		q := reqUrl.Query()
		q.Set("results", strconv.FormatInt(USERS_PER_REQUEST, 10))

		reqUrl.RawQuery = q.Encode()
		reqUrlString := reqUrl.String()
		log.Printf("Target of GET requests: %s", reqUrlString)

		log.Printf("Initializing transaction...")
		var tx *sql.Tx
		tx, err = db.BeginTx(reqContext, nil)
		if err != nil {
			log.Printf("Failed to open transaction: `%s`", err)
			http.Error(w, "Failed to open query transaction", http.StatusInternalServerError)
			return
		}

		defer func() {
			var comRollErr error
			if err != nil {
				comRollErr = tx.Rollback()
			} else {
				comRollErr = tx.Commit()
			}

			if comRollErr != nil {
				log.Panicf("Error rolling back/committing transaction! %s", err)
			}
		}()

		processedUsers := make(chan User, USERS_PER_REQUEST*2)

		workTickets := make(chan bool, WORKER_COUNT*2)
		// Signal work thread
		go func() {
			defer close(workTickets)

			for range TARGET_USER_COUNT / USERS_PER_REQUEST {
				select {
				case <-reqContext.Done():
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

		workerCtx, cancelWorkerCtx := context.WithCancelCause(reqContext)
		defer cancelWorkerCtx(nil)

		workersGroup := sync.WaitGroup{}
		for workerIdx := range WORKER_COUNT {
			workersGroup.Add(1)
			go func() {
				defer func() {
					log.Println("Worker ", workerIdx, " Done!")
					workersGroup.Done()
				}()

			workerLoop:
				for {
					select {
					case <-workerCtx.Done():
						log.Println("Stopping worker due to context being canceled...")
						return
					case shouldKeep := <-workTickets:
						if !shouldKeep {
							break workerLoop // Exit processing loop once makeReqSignal closes
						}

						err = errors.New("Dummy error for execution purposes")
						retryCounter := 0

					requestLoop:
						for retryCounter = 0; err != nil && retryCounter < RETRY_LIMIT; retryCounter += 1 {
							err = nil

							select {
							case <-workerCtx.Done():
								log.Println("Stopping worker due to context being canceled...")
								return

							default:
								var req *http.Request
								req, err = http.NewRequestWithContext(workerCtx, http.MethodGet, reqUrlString, nil)
								if err != nil {
									log.Println("Error: Creating GET request")
									continue requestLoop
								}

								var resp *http.Response
								resp, err = http.DefaultClient.Do(req)
								if err != nil {
									log.Printf("Error: Doing GET request. `%s`", err)
									continue requestLoop
								}

								var respBytes []byte
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

								var stmt *sql.Stmt
								stmt, err = tx.Prepare(pq.CopyIn("db_user", "id", "gender", "name", "location", "city", "state", "country", "email", "phone"))
								if err != nil {
									log.Printf("Error: Failed to create prepared statement for batch insert!")
									continue requestLoop
								}

								if useStatement(&apiResponse, stmt, err, processedUsers) {
									continue requestLoop
								}
							}
						}

						// If we reached the retry limit and didn't succeed then we need to respond as failure and cancel the request.
						if retryCounter == RETRY_LIMIT && err != nil {
							log.Printf("Error: Worker %d reached RETRY_LIMIT (%d). Err: %s\n", workerIdx, RETRY_LIMIT, err)
							cancelWorkerCtx(RETRY_LIMIT_EXCEEDED_ERR)
							break workerLoop
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

		err = workerCtx.Err()
		if err == RETRY_LIMIT_EXCEEDED_ERR {
			log.Printf("Workers failed because: %s", err.Error())
			http.Error(w, "RETRY LIMIT EXCEEDED", http.StatusInternalServerError)
			return
		} else if err != nil {
			log.Printf("Workers failed because: %s", err)
			http.Error(w, "INTERNAL SERVER ERROR", http.StatusInternalServerError)
			return
		}

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

// Returns true if it should skip the rest of the iteration
func useStatement(apiResponse *shared.APIResponse,
	stmt *sql.Stmt,
	err error,
	processedUsers chan<- User) bool {
	defer stmt.Close()

	for _, v := range apiResponse.Results {
		user := NewUserFromAPI(v, apiResponse.Info)
		_, err = stmt.Exec(
			user.Id,
			user.Gender,
			user.Name,
			user.Location,
			user.City,
			user.State,
			user.Country,
			user.Email,
			user.Phone,
		)
		if err != nil {
			log.Printf("Error: Failed to add user to prepared statement! `%s`", err)
			return true
		}
		processedUsers <- user
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Printf("Error: Failed to execute prepared statement! `%s`", err)
		return true
	}

	return false
}
