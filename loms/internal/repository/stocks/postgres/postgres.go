package postgres

import (
	"context"
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	sqlcstocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/stocks/postgres/sqlc"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool    *pgxpool.Pool
	queries *sqlcstocks.Queries
}

func NewPostgresRepository(pool *pgxpool.Pool) *postgresRepository {
	return &postgresRepository{
		pool:    pool,
		queries: sqlcstocks.New(pool),
	}
}

func (r *postgresRepository) queriesFor(ctx context.Context) *sqlcstocks.Queries {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *postgresRepository) GetStock(ctx context.Context, sku uint32) (uint64, error) {
	count, err := r.queriesFor(ctx).GetStock(ctx, int32(sku))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return uint64(count), nil
}

func (r *postgresRepository) ReserveStocks(ctx context.Context, items []entity.OrderItem) error {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return r.reserveStocks(ctx, r.queries.WithTx(tx), items)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := r.reserveStocks(ctx, r.queries.WithTx(tx), items); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *postgresRepository) reserveStocks(ctx context.Context, q *sqlcstocks.Queries, items []entity.OrderItem) error {
	for _, item := range items {
		rows, err := q.ReserveStock(ctx, sqlcstocks.ReserveStockParams{
			Sku:   int32(item.SKU),
			Count: int64(item.Count),
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			if _, err := q.GetStock(ctx, int32(item.SKU)); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return xerrors.ErrInsufficientStock
				}
				return err
			}
			return xerrors.ErrInsufficientStock
		}
	}
	return nil
}

func (r *postgresRepository) ReleaseStock(ctx context.Context, sku uint32, count uint64) error {
	rows, err := r.queriesFor(ctx).ReleaseStock(ctx, sqlcstocks.ReleaseStockParams{
		Sku:   int32(sku),
		Count: int64(count),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return xerrors.ErrStockNotFound
	}
	return nil
}

func (r *postgresRepository) SetStock(ctx context.Context, sku uint32, count uint64) error {
	return r.queriesFor(ctx).UpsertStock(ctx, sqlcstocks.UpsertStockParams{
		Sku:   int32(sku),
		Count: int64(count),
	})
}
