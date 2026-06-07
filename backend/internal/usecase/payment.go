package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/your-handle/brewly/internal/domain"
)

// validPaymentMethods is the set of accepted payment method strings.
var validPaymentMethods = map[string]bool{
	domain.PaymentMethodCash: true,
	domain.PaymentMethodQRIS: true,
	domain.PaymentMethodCard: true,
}

// PaymentUsecase handles payment business logic.
type PaymentUsecase struct {
	paymentRepo domain.PaymentRepository
	orderRepo   domain.OrderRepository
}

// NewPaymentUsecase constructs a PaymentUsecase.
func NewPaymentUsecase(paymentRepo domain.PaymentRepository, orderRepo domain.OrderRepository) *PaymentUsecase {
	return &PaymentUsecase{paymentRepo: paymentRepo, orderRepo: orderRepo}
}

// Record creates a payment entry for an order.
func (u *PaymentUsecase) Record(
	ctx context.Context,
	orderID uuid.UUID,
	recordedByUserID uuid.UUID,
	method string,
	amountMinor int64,
	receivedMinor int64,
) (*domain.Payment, error) {
	if !validPaymentMethods[method] {
		return nil, fmt.Errorf("usecase.Payment.Record: invalid method %q", method)
	}

	if _, err := u.orderRepo.FindByID(ctx, orderID); err != nil {
		return nil, fmt.Errorf("usecase.Payment.Record: %w", err)
	}

	p := &domain.Payment{
		OrderID:          orderID,
		Method:           method,
		AmountMinor:      amountMinor,
		ReceivedMinor:    receivedMinor,
		RecordedByUserID: recordedByUserID,
	}
	if err := u.paymentRepo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("usecase.Payment.Record: %w", err)
	}
	return p, nil
}

// ListByOrder returns all payments for a given order.
func (u *PaymentUsecase) ListByOrder(ctx context.Context, orderID uuid.UUID) ([]domain.Payment, error) {
	payments, err := u.paymentRepo.ListByOrder(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("usecase.Payment.ListByOrder: %w", err)
	}
	return payments, nil
}
