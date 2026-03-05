package service

import (
	paymentsvc "github.com/tormenta-bot/internal/services/payment"
)

type PaymentService struct {
	core *paymentsvc.Service
}

func NewPaymentService() *PaymentService {
	return &PaymentService{core: paymentsvc.NewService()}
}

func (s *PaymentService) ConfirmByTxID(txid, source string) (paymentsvc.ConfirmResult, error) {
	return s.core.ConfirmByTxID(txid, source)
}
