package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	shared "github.com/ElrohirGT/GoAcademyExercises/ConcurrencyExercise/Shared"
)

type params struct {
	UserCount uint64
}

func parseParams(w http.ResponseWriter, r *http.Request) (params, error) {
	var err error
	var params params

	var userCount uint64 = 1
	resultsQuery := r.URL.Query().Get("results")
	if resultsQuery != "" {
		userCount, err = strconv.ParseUint(resultsQuery, 10, 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return params, err
		}
	}

	params.UserCount = userCount
	return params, nil

}

func mod(a, b uint64) uint64 {
	if a < b {
		return a
	} else {
		return a % b
	}
}

func get_users(db *[]shared.APIUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dbLength uint64 = uint64(len((*db)))
		params, err := parseParams(w, r)
		if err != nil {
			return
		}

		users := make([]shared.APIUser, 0, params.UserCount)
		idxs := rand.Perm(int(dbLength))

		for i := range params.UserCount {
			idx := idxs[mod(i, dbLength)]
			user := (*db)[idx]
			users = append(users, user)
		}

		respBytes, err := json.Marshal(users)
		if err != nil {
			panic("Can't marshall the users response to json!")
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(respBytes)
	}
}
