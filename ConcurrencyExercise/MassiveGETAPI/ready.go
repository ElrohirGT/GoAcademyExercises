package main

import "net/http"

func health(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "", http.StatusNoContent)
}
