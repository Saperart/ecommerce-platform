package postgres

import (
	"context"
	"fmt"
	"time"

	sqlcoutbox "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres/sqlc"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor"
	"github.com/jackc/pgx/v5/pgtype"
)

type postgresRepository struct {
	queries *sqlcoutbox.Queries
}

func NewPostgresRepository(db sqlcoutbox.DBTX) *postgresRepository {
	return &postgresRepository{queries: sqlcoutbox.New(db)}
}

func (r *postgresRepository) queriesFor(ctx context.Context) *sqlcoutbox.Queries {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *postgresRepository) SaveMessage(ctx context.Context, idempotencyKey string, kind Kind, payload []byte) error {
	if err := r.queriesFor(ctx).SaveMessage(ctx, sqlcoutbox.SaveMessageParams{
		IdempotencyKey: idempotencyKey,
		Kind:           string(kind),
		Payload:        payload,
	}); err != nil {
		return fmt.Errorf("save outbox message: %w", err)
	}
	return nil
}

func (r *postgresRepository) ClaimMessages(ctx context.Context, batchSize int, inProgressTTL time.Duration) ([]Message, error) {
	rows, err := r.queries.ClaimMessages(ctx, sqlcoutbox.ClaimMessagesParams{
		InProgressTtl: pgtype.Interval{
			Microseconds: inProgressTTL.Microseconds(),
			Valid:        true,
		},
		BatchSize: int32(batchSize),
	})
	if err != nil {
		return nil, fmt.Errorf("claim outbox messages: %w", err)
	}

	messages := make([]Message, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, Message{
			IdempotencyKey: row.IdempotencyKey,
			Kind:           Kind(row.Kind),
			Payload:        row.Payload,
		})
	}
	return messages, nil
}

func (r *postgresRepository) MarkProcessed(ctx context.Context, idempotencyKeys []string) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}
	if err := r.queries.MarkProcessed(ctx, idempotencyKeys); err != nil {
		return fmt.Errorf("mark outbox messages processed: %w", err)
	}
	return nil
}

func (r *postgresRepository) MarkRetryable(ctx context.Context, idempotencyKeys []string) error {
	if len(idempotencyKeys) == 0 {
		return nil
	}
	if err := r.queries.MarkRetryable(ctx, idempotencyKeys); err != nil {
		return fmt.Errorf("mark outbox messages retryable: %w", err)
	}
	return nil
}
