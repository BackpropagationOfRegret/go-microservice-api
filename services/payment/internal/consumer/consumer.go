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
	failRate          float64
}

func NewProcessor(repo *repository.Repository, processed, failed *kafka.Producer, failRate float64) *Processor {
	return &Processor{repo: repo, processedProducer: processed, failedProducer: failed, failRate: failRate}
}

func (p *Processor) HandleOrderCreated(ctx context.Context, env events.Envelope) error {
	order, err := kafka.DecodePayload[events.OrderCreated](env)
	if err != nil {
		return err
	}

	log.Printf("[payment-svc] processing payment for order %s, amount %.2f", order.OrderID, order.TotalAmount)

	// Simulate payment gateway latency
	time.Sleep(500 * time.Millisecond)

	if rand.Float64() < p.failRate {
		reason := "insufficient funds"
		_ = p.repo.Create(ctx, &repository.Payment{
			OrderID: order.OrderID,
			Amount:  order.TotalAmount,
			Method:  "card",
			Status:  "FAILED",
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
		Method:        "card",
		Status:        "SUCCESS",
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
