package licitations

import (
	"encoding/json"
	"errors"
	"net/http"

	"goacademy.com/week3/api"
	core_licitations "goacademy.com/week3/core/licitations"

	"goacademy.com/week3/services/countries"
	"goacademy.com/week3/services/tenders"
)

func GETLicitationsWithCountryName(w http.ResponseWriter, r *http.Request) {
	countryCode := r.PathValue("countryCode")
	if countryCode == "" || countryCode != "hu" {
		api.NewFromError(errors.New("Invalid country code supplied")).JSONIfyAndRespond(w, http.StatusBadRequest)
		return
	}

	deps := core_licitations.GetTendersByCountryCodeDeps{
		GetCountryNameByCode: countries.GetCountryByCode,
		GetTendersByCode:     tenders.GetTendersByCountryCode,
	}
	coreResponse, err := core_licitations.GetTendersByCountryCode(r.Context(), deps, countryCode)
	if err != nil {
		api.NewFromError(err).JSONIfyAndRespond(w, http.StatusInternalServerError)
		return
	}

	// Uncomment if you wan't to see what happens when the gracefull shutdown times out!
	// time.Sleep(10 * time.Second)

	respBytes, err := json.Marshal(coreResponse)
	if err != nil {
		api.NewFromError(err).JSONIfyAndRespond(w, http.StatusInternalServerError)
		return
	}
	w.Write(respBytes)
}
