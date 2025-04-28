package payments

import (
	"context"
	"log"
	"time"
)

type CardPayment struct {
	amount float64
}

func NewCardPayment(amount float64) CardPayment {
	return CardPayment{
		amount: amount,
	}
}

func (self CardPayment) Execute(ctx context.Context) bool {
	select {
	case <-ctx.Done(): //context cancelled
		return false
	// case <-time.After(2 * time.Second):
	default:
		if self.amount >= MIN_PAYMENT_AMOUNT || self.amount <= MAX_PAYMENT_AMOUNT {
			log.Default().Println("Card payment with: ", self.amount)

			time.Sleep(time.Second)
			return true
		}
		return false
	}
}

func (self CardPayment) Refund() bool {
	log.Default().Println("Card payment refunded!")
	return true
}
