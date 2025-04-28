package payments

import (
	"context"
	"log"
)

type PaymentProcessor struct {
}

func (self PaymentProcessor) Process(ctx context.Context, payment Payable) bool {
	if _, ok := payment.(Refundable); ok {
		log.Default().Println("The payment is refundable!")
	}

	return payment.Execute(ctx)
}
