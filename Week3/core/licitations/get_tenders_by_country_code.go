package licitations

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"goacademy.com/week3/core"
	"goacademy.com/week3/services/countries"
	"goacademy.com/week3/services/tenders"
)

type GetTendersByCountryCodeDeps struct {
	GetCountryNameByCode func(context.Context, string) (countries.CountryNameResponse, error)
	GetTendersByCode     func(context.Context, string) (tenders.GuruResponse, error)
}

type Task[T any] struct {
	result T
	error  error
}

func GetTendersByCountryCode(ctx context.Context, deps GetTendersByCountryCodeDeps, countryCode string) (core.FinalResponse, error) {

	group := sync.WaitGroup{}

	group.Add(1)
	restResult := make(chan Task[[]core.Tender], 1)
	defer close(restResult)

	go func() {
		defer group.Done()

		respTenders, err := deps.GetTendersByCode(ctx, countryCode)
		if err != nil {
			restResult <- Task[[]core.Tender]{error: err}
			return
		}

		tenders := []core.Tender{}
		for i, elm := range respTenders.Data {

			if i > 15 {
				break
			}

			supplierName := ""
			if len(elm.Awarded) > 0 && len(elm.Awarded[0].Suppliers) > 0 {
				supplierName = elm.Awarded[0].Suppliers[0].Name
			}

			awardValue, err := strconv.ParseFloat(elm.Awarded_value, 64)
			if err != nil {
				restResult <- Task[[]core.Tender]{error: err}
				return
			}

			tenders = append(tenders, core.Tender{
				Id:               elm.Id,
				Title:            elm.Title,
				Date:             elm.Date,
				Description:      elm.Description,
				Place:            elm.Place,
				Awarded_value:    awardValue,
				Awarded_currency: elm.Awarded_currency,
				Purchaser_name:   elm.Purcharser.Name,
				Supplier_name:    supplierName,
			})
		}

		restResult <- Task[[]core.Tender]{result: tenders}
	}()

	group.Add(1)
	soapResult := make(chan Task[countries.CountryNameResponse], 1)
	defer close(soapResult)

	go func() {
		defer group.Done()

		soapResponse, err := deps.GetCountryNameByCode(ctx, countryCode)
		if err != nil {
			soapResult <- Task[countries.CountryNameResponse]{error: err}
			return
		}

		soapResult <- Task[countries.CountryNameResponse]{result: soapResponse}
	}()

	group.Wait()

	// go func() {
	// 	group.Wait()
	// 	close(soapResult)
	// 	close(restResult)
	// }()

	select {
	case <-ctx.Done():
		return core.FinalResponse{}, errors.New("Server closing!")
	default:
		soapResponse := <-soapResult
		restResponse := <-restResult

		if soapResponse.error != nil {
			return core.FinalResponse{}, soapResponse.error
		}

		if restResponse.error != nil {
			return core.FinalResponse{}, restResponse.error
		}

		return core.FinalResponse{
			CountryName: soapResponse.result.CountryNameResult,
			Tenders:     restResponse.result,
		}, nil
	}
}
