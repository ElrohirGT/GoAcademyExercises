package get_summary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	get_random_user "github.com/ElrohirGT/GoAcademyExercises/FinalTest/services/getRandomUser"
	utils "github.com/ElrohirGT/GoAcademyExercises/FinalTest/utils"
)

type Params struct {
	TargetUsers uint
}

type TopCountries struct {
	data []get_random_user.TopCountry
}

func (s TopCountries) Len() int {
	return len(s.data)
}

func (s TopCountries) Less(a int, b int) bool {
	return s.data[a].Count > s.data[b].Count
}

func (s TopCountries) Swap(a int, b int) {
	s.data[a], s.data[b] = s.data[b], s.data[a]
}

func MapToAggregation(resp *[]get_random_user.APIUser) aggregation {
	userCount := uint(len(*resp))

	gender := get_random_user.GenderDistribution{}
	countriesMap := make(map[string]int)
	sumOfAges := uint(0)

	for _, apiUser := range *resp {
		sumOfAges += apiUser.Registered.Age
		if apiUser.Gender == "male" {
			gender.Male += 1
		} else if apiUser.Gender == "female" {
			gender.Female += 1
		}

		if _, found := countriesMap[apiUser.Location.Country]; !found {
			countriesMap[apiUser.Location.Country] = 0
		}
		countriesMap[apiUser.Location.Country] += 1
	}

	return aggregation{
		TotalUsers:         userCount,
		GenderDistribution: gender,
		SumAge:             float64(sumOfAges),
		TopCountries:       countriesMap,
	}
}

func MapAggregationToAPIResponse(resp *aggregation) Response {
	countries := TopCountries{data: []get_random_user.TopCountry{}}
	for k, v := range resp.TopCountries {
		countries.data = append(countries.data, get_random_user.TopCountry{
			Country: k,
			Count:   uint(v),
		})
	}

	sort.Sort(countries)

	if len(countries.data) > 3 {
		countries.data = countries.data[:3]
	}

	return Response{
		TotalUsers:         resp.TotalUsers,
		GenderDistribution: resp.GenderDistribution,
		AvgAge:             resp.SumAge / float64(resp.TotalUsers),
		TopCountries:       countries.data,
	}
}

func ParseParams(w http.ResponseWriter, r *http.Request) (Params, error) {
	params := Params{}

	resultsStr := r.URL.Query().Get("results")
	if resultsStr == "" {
		http.Error(w, "Invalid results query param", http.StatusBadRequest)
		return params, nil
	}

	targetUsers, err := strconv.Atoi(resultsStr)
	if err != nil {
		http.Error(w, "results query must be a number between 1000 and 15000", http.StatusBadRequest)
		return params, err
	}

	params.TargetUsers = uint(targetUsers)

	return params, nil
}

type aggregation struct {
	TotalUsers         uint
	GenderDistribution get_random_user.GenderDistribution
	SumAge             float64
	TopCountries       map[string]int
}

type Response struct {
	TotalUsers         uint
	GenderDistribution get_random_user.GenderDistribution
	AvgAge             float64
	TopCountries       []get_random_user.TopCountry
}

func CopyTo(resp *map[string]int, agg *map[string]int) {
	res := *resp

	for k, v := range *agg {
		if _, found := res[k]; !found {
			res[k] = 0
		}

		res[k] += v
	}
}

var CACHE = utils.Cache[get_random_user.APIUser]{
	Data:       []get_random_user.APIUser{},
	Length:     0,
	ExpireTime: time.Now(),
}

const MAX_RETRIES = 10

var RETRY_LIMIT = errors.New("MAX RETRY LIMIT")
var EMPTY_RESPONSE = errors.New("GET_RANDOM_USER EMPTY RESPONSE")

func GetSummary(w http.ResponseWriter, r *http.Request) {
	reqContext := r.Context()

	params, err := ParseParams(w, r)
	if err != nil {
		return
	}

	if params.TargetUsers < 1000 || params.TargetUsers > 15000 {
		http.Error(w, "results must be between 1000 and 15000", http.StatusBadRequest)
		return
	}

	// if params.TargetUsers < CACHE.Length {
	// }

	workerCount := params.TargetUsers/get_random_user.RESULTS_LIMIT + 1
	usersPerWorker := params.TargetUsers / uint(workerCount)
	outputResponse := make(chan aggregation, workerCount*3)

	log.Printf("Creating workers: #Workers: %d #UsersPerWorker: %d", workerCount, usersPerWorker)

	// Work divider thread
	workersTicket := make(chan uint, workerCount)
	workersReturnTicket := make(chan uint, workerCount)
	go func() {

		remainingUsers := params.TargetUsers

		for {
			workload := min(remainingUsers, get_random_user.RESULTS_LIMIT)
			if workload != 0 {
				select {
				case workersTicket <- workload:
					remainingUsers -= workload
				case remainingWorkload := <-workersReturnTicket:
					remainingUsers += remainingWorkload
				case <-reqContext.Done():
					return
				}
			} else {
				select {
				case <-reqContext.Done():
					return
				case remainingWorkload := <-workersReturnTicket:
					if remainingWorkload == 0 {
						return
					}

					remainingUsers += remainingWorkload
				}
			}
		}
	}()

	// Aggregator thread
	resultChannel := make(chan *aggregation, 1)
	aggGroup := sync.WaitGroup{}
	aggGroup.Add(1)
	go func() {
		defer aggGroup.Done()

		agr := aggregation{TopCountries: make(map[string]int)}
		for respAgr := range outputResponse {
			log.Println("The aggregation is:", respAgr)

			agr.SumAge += respAgr.SumAge
			agr.TotalUsers += respAgr.TotalUsers
			agr.GenderDistribution.Male += respAgr.GenderDistribution.Male
			agr.GenderDistribution.Female += respAgr.GenderDistribution.Female
			CopyTo(&agr.TopCountries, &respAgr.TopCountries)
		}
		resultChannel <- &agr
	}()

	// Worker threads
	workerCtx, cancelWorkers := context.WithCancelCause(reqContext)
	defer cancelWorkers(nil)

	workerGroup := sync.WaitGroup{}
	for i := range workerCount {
		workerGroup.Add(1)
		go func() {
			defer workerGroup.Done()
			workerLogPrefix := fmt.Sprintf("Worker %d:", i)

		workerLifeCycleLoop:
			for {
				select {
				case <-workerCtx.Done():
					return
				case workload := <-workersTicket:
					log.Println(workerLogPrefix, "Received workload", workload)
					workerUsers := make([]get_random_user.APIUser, 0, workload)
					if workload == 0 {
						log.Println(workerLogPrefix, "No more work to do, shutting down worker...")
						break workerLifeCycleLoop
					}

					log.Println(workerLogPrefix, "Creating URL...")
					url, err := url.Parse(get_random_user.BASE_URL)
					if err != nil {
						log.Panicf("Failed to parse URL: %s", err)
						cancelWorkers(err)
					}

					log.Println(workerLogPrefix, "Adding query to URL...")
					q := url.Query()
					q.Set("results", strconv.FormatInt(int64(workload), 10))
					url.RawQuery = q.Encode()

					urlString := url.String()

					reqError := errors.New("Dummy error to continue in loop")
					var retryCount int

				requestLoop:
					for retryCount = 0; reqError != nil && retryCount < MAX_RETRIES; retryCount += 1 {
						log.Printf("%s Attempt #%d/%d for worker %d", workerLogPrefix, retryCount+1, MAX_RETRIES, i)
						reqError = nil

						log.Println(workerLogPrefix, "Constructing request...")
						var req *http.Request
						req, reqError = http.NewRequestWithContext(workerCtx, http.MethodGet, urlString, nil)
						if reqError != nil {
							log.Fatal(workerLogPrefix, "Failed to create request.", reqError)
						}

						log.Println(workerLogPrefix, "Making request...")
						var resp *http.Response
						resp, reqError = http.DefaultClient.Do(req)
						if reqError != nil {
							log.Println(workerLogPrefix, "Failed to GET response.", reqError)
							continue requestLoop
						}

						log.Println(workerLogPrefix, "Reading all body bytes...")
						var respBody []byte
						respBody, reqError = io.ReadAll(resp.Body)
						if reqError != nil {
							log.Println(workerLogPrefix, "Failed to read the request body.", reqError)
							continue requestLoop
						}

						log.Println(workerLogPrefix, "Parsing response body...")
						var response get_random_user.APIResponse
						reqError = json.Unmarshal(respBody, &response)
						if reqError != nil {
							log.Println(workerLogPrefix, "Failed to unmarshal request body.", reqError)
							continue requestLoop
						}

						lengthResults := uint(len(response.Results))
						if lengthResults == 0 {
							reqError = EMPTY_RESPONSE
							log.Println(workerLogPrefix, "API returns 0 length array! ", reqError)
							continue requestLoop
						}

						log.Println(workerLogPrefix, "Appending response results to local worker users...", lengthResults)
						workerUsers = append(workerUsers, response.Results...)

						log.Println(workerLogPrefix, "Checking if some users are missing...")
						if lengthResults != workload {
							log.Printf("%s Response results isn't equal to workload. %d != %d", workerLogPrefix, lengthResults, workload)

							missingUsers := (workload - lengthResults)
							log.Println(workerLogPrefix, "Returning workload of ", missingUsers)
							select {
							case <-workerCtx.Done():
								return
							case workersReturnTicket <- missingUsers:
								continue workerLifeCycleLoop
							}
						}

						log.Println(workerLogPrefix, "No users are missing...")
						mappedResp := MapToAggregation(&workerUsers)
						log.Println(workerLogPrefix, "The amount of users on the worker local DB:", len(workerUsers))
						outputResponse <- mappedResp
						return // End worker execution
					}

					log.Println(workerLogPrefix, "Checking if it ended because of retries or not...")
					if retryCount == MAX_RETRIES {
						log.Println(workerLogPrefix, "Reached retry limit because:", reqError)
						cancelWorkers(RETRY_LIMIT)
						return
					} else {
						log.Panic(workerLogPrefix, " Ended even though retryCount is not maxtries: ", retryCount, reqError)
					}
				}
			}
		}()
	}

	log.Println("Waiting for all workers...")
	workerGroup.Wait()
	close(outputResponse)
	close(workersTicket)
	close(workersReturnTicket)

	workersErr := workerCtx.Err()
	if workersErr != nil {
		log.Println("Workers failed because ", workersErr)
		http.Error(w, "Reached max retries!", http.StatusInternalServerError)
		return
	}

	log.Println("Waiting for aggregator...")
	aggGroup.Wait()

	log.Println("Generating final response...")
	select {
	case <-reqContext.Done():
		return
	case finalAgr := <-resultChannel:
		finalResponse := MapAggregationToAPIResponse(finalAgr)
		log.Println("Final response:", finalResponse)
		var respBytes []byte
		respBytes, err = json.Marshal(finalResponse)
		if err != nil {
			log.Println("Error on response", err)
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(respBytes)
	}

}
