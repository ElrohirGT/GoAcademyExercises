package countries

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type CountryNameResponse struct {
	CountryNameResult string
}

type SOAPResponseBody struct {
	CountryNameResponse CountryNameResponse
}

type SOAPResponse struct {
	Body SOAPResponseBody
}

func GetCountryByCode(ctx context.Context, code string) (CountryNameResponse, error) {
	reqBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	  <soap:Body>
	    <CountryName xmlns="http://www.oorsprong.org/websamples.countryinfo">
	      <sCountryISOCode>%s</sCountryISOCode>
	    </CountryName>
	  </soap:Body>
	</soap:Envelope>`, strings.ToUpper(code))
	bodyReader := strings.NewReader(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://webservices.oorsprong.org/websamples.countryinfo/CountryInfoService.wso", bodyReader)
	if err != nil {
		return CountryNameResponse{}, err
	}

	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	req.Header.Add("SOAPAction", "\"http://www.oorsprong.org/websamples.countryinfo/CountryName\"")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CountryNameResponse{}, nil
	}

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return CountryNameResponse{}, nil
	}
	fmt.Println("The response bytes: ", string(respBodyBytes))

	var soapResponse SOAPResponse
	err = xml.Unmarshal(respBodyBytes, &soapResponse)
	if err != nil {
		return CountryNameResponse{}, nil
	}

	return soapResponse.Body.CountryNameResponse, nil
}
