package ready

import "net/http"

func Ready(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid method returned", http.StatusMethodNotAllowed)
		return
	}

	w.Write([]byte("SUCCESS"))
}
