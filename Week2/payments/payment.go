package payments

import (
	"context"
)

const MIN_PAYMENT_AMOUNT = 1.0
const MAX_PAYMENT_AMOUNT = 1000.0

type Payable interface {
	Execute(ctx context.Context) bool
}

type Refundable interface {
	Payable
	Refund() bool
}
