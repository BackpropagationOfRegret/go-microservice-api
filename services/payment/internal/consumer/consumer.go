package consumer

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/kostayne/go-microservice/pkg/events"
	"github.com/kostayne/go-microservice/pkg/kafka"
	"github.com/kostayne/go-microservice/services/payment/internal/repository"
)

type Processor struct {
	repo              *repository.Repository
	processedProducer *kafka.Producer
	failedProducer    *kafka.Producer
	refundedProducer  *kafka.Producer
	failRate          float64
	processLatency    time.Duration
	refundLatency     time.Duration
	paymentMethod     string
}

func NewProcessor(
	repo *repository.Repository,
	processed, failed, refunded *kafka.Producer,
	failRate float64,
	processLatency, refundLatency time.Duration,
	paymentMethod string,
) *Processor {
	return &Processor{
		repo:              repo,
		processedProducer: processed,
		failedProducer:    failed,
		refundedProducer:  refunded,
		failRate:          failRate,
		processLatency:    processLatency,
		refundLatency:     refundLatency,
		paymentMethod:     paymentMethod,
	}
}

func (p *Processor) HandleOrderCreated(ctx context.Context, env events.Envelope) error {
	order, err := kafka.DecodePayload[events.OrderCreated](env)
	if err != nil {
		return err
	}

	existing, err := p.repo.GetByOrderID(ctx, order.OrderID)
	if err == nil {
		switch existing.Status {
		case repository.StatusSuccess:
			return nil
		case repository.StatusFailed:
			return nil
		case repository.StatusRefunded:
			return nil
		}
	}

	log.Printf("[payment-svc] processing payment for order %s, amount %.2f", order.OrderID, order.TotalAmount)
	time.Sleep(p.processLatency)

	if rand.Float64() < p.failRate {
		reason := "insufficient funds"
		_ = p.repo.Create(ctx, &repository.Payment{
			OrderID: order.OrderID,
			Amount:  order.TotalAmount,
			Method:  p.paymentMethod,
			Status:  repository.StatusFailed,
		})
		log.Printf("[payment-svc] payment failed for order %s: %s", order.OrderID, reason)
		return p.failedProducer.Publish(ctx, events.TopicPaymentFailed, events.PaymentFailed{
			OrderID: order.OrderID,
			Reason:  reason,
		})
	}

	txID := uuid.New().String()
	if err := p.repo.Create(ctx, &repository.Payment{
		OrderID:       order.OrderID,
		Amount:        order.TotalAmount,
		Method:        p.paymentMethod,
		Status:        repository.StatusSuccess,
		TransactionID: txID,
	}); err != nil {
		return err
	}

	log.Printf("[payment-svc] payment success for order %s, tx %s", order.OrderID, txID)
	return p.processedProducer.Publish(ctx, events.TopicPaymentProcessed, events.PaymentProcessed{
		OrderID:       order.OrderID,
		TransactionID: txID,
		Amount:        order.TotalAmount,
	})
}

func (p *Processor) HandleRefundRequested(ctx context.Context, env events.Envelope) error {
	req, err := kafka.DecodePayload[events.PaymentRefundRequested](env)
	if err != nil {
		return err
	}

	payment, err := p.repo.GetByOrderID(ctx, req.OrderID)
	if err != nil {
		log.Printf("[payment-svc] refund skipped, no payment for order %s", req.OrderID)
		return nil
	}
	if payment.Status == repository.StatusRefunded {
		return nil
	}
	if payment.Status != repository.StatusSuccess {
		log.Printf("[payment-svc] refund skipped, payment status %s for order %s", payment.Status, req.OrderID)
		return nil
	}

	log.Printf("[payment-svc] refunding order %s, amount %.2f, reason: %s", req.OrderID, payment.Amount, req.Reason)
	time.Sleep(p.refundLatency)

	refundTxID := uuid.New().String()
	if err := p.repo.MarkRefunded(ctx, req.OrderID, refundTxID, req.Reason); err != nil {
		return err
	}

	log.Printf("[payment-svc] refund completed for order %s, refund_tx %s", req.OrderID, refundTxID)
	return p.refundedProducer.Publish(ctx, events.TopicPaymentRefunded, events.PaymentRefunded{
		OrderID:             req.OrderID,
		RefundTransactionID: refundTxID,
		Amount:              payment.Amount,
		Reason:              req.Reason,
	})
}
