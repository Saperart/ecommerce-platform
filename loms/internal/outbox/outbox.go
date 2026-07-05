package outbox

import (
	"context"
	"sync"
	"time"

	outboxrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres"
	"go.uber.org/zap"
)

//go:generate mockgen -source=outbox.go -destination=mocks/outbox_mock.go -package=mocks

type Repository interface {
	ClaimMessages(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]outboxrepo.Message, error)
	MarkProcessed(ctx context.Context, idempotencyKeys []string) error
	MarkRetryable(ctx context.Context, idempotencyKeys []string) error
}

type Handler func(ctx context.Context, payload []byte) error
type HandlerByKind func(kind outboxrepo.Kind) (Handler, error)

type Outbox struct {
	logger        *zap.Logger
	repository    Repository
	handlerByKind HandlerByKind
}

func New(logger *zap.Logger, repository Repository, handlerByKind HandlerByKind) *Outbox {
	return &Outbox{
		logger:        logger,
		repository:    repository,
		handlerByKind: handlerByKind,
	}
}

func (o *Outbox) Start(
	ctx context.Context,
	workers int,
	batchSize int,
	fetchPeriod time.Duration,
	inProgressTTL time.Duration,
) {
	if workers <= 0 || batchSize <= 0 || fetchPeriod <= 0 {
		return
	}

	wg := new(sync.WaitGroup)
	for workerID := 1; workerID <= workers; workerID++ {
		wg.Add(1)
		go o.worker(ctx, wg, workerID, batchSize, fetchPeriod, inProgressTTL)
	}
}

func (o *Outbox) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	workerID int,
	batchSize int,
	fetchPeriod time.Duration,
	inProgressTTL time.Duration,
) {
	defer wg.Done()

	ticker := time.NewTicker(fetchPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			o.processBatch(ctx, workerID, batchSize, inProgressTTL)
		}
	}
}

func (o *Outbox) processBatch(ctx context.Context, workerID int, batchSize int, inProgressTTL time.Duration) {
	messages, err := o.repository.ClaimMessages(ctx, batchSize, inProgressTTL)
	if err != nil {
		o.logger.Error("claim outbox messages", zap.Int("worker_id", workerID), zap.Error(err))
		return
	}
	if len(messages) == 0 {
		return
	}

	successKeys := make([]string, 0, len(messages))
	failedKeys := make([]string, 0, len(messages))

	for _, message := range messages {
		handler, err := o.handlerByKind(message.Kind)
		if err != nil {
			o.logger.Error("unsupported outbox kind", zap.String("kind", string(message.Kind)), zap.Error(err))
			failedKeys = append(failedKeys, message.IdempotencyKey)
			continue
		}

		if err := handler(ctx, message.Payload); err != nil {
			o.logger.Error("process outbox message", zap.String("idempotency_key", message.IdempotencyKey), zap.Error(err))
			failedKeys = append(failedKeys, message.IdempotencyKey)
			continue
		}
		successKeys = append(successKeys, message.IdempotencyKey)
	}

	if err := o.repository.MarkProcessed(ctx, successKeys); err != nil {
		o.logger.Error("mark outbox messages processed", zap.Int("worker_id", workerID), zap.Error(err))
	}
	if err := o.repository.MarkRetryable(ctx, failedKeys); err != nil {
		o.logger.Error("mark outbox messages retryable", zap.Int("worker_id", workerID), zap.Error(err))
	}
}
