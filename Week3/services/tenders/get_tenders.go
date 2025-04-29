package tenders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type GuruResponse struct {
	Data []ReqLicitation `json:"data"`
}

func GetTendersByCountryCode(ctx context.Context, countryCode string) (GuruResponse, error) {
	if countryCode == "" || countryCode != "hu" {
		err := errors.New("Invalid country code")
		return GuruResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://tenders.guru/api/%s/tenders", countryCode), nil)
	if err != nil {
		return GuruResponse{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return GuruResponse{}, err
	}

	var guruResponse GuruResponse
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return GuruResponse{}, nil
	}
	fmt.Println("The response bytes: ", string(respBodyBytes))

	err = json.Unmarshal(respBodyBytes, &guruResponse)
	if err != nil {
		return GuruResponse{}, nil
	}

	return guruResponse, nil
}
