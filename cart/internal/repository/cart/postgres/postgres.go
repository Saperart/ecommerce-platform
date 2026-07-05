package postgres

import (
	"context"

	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"
	sqlccart "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/repository/cart/postgres/sqlc"
)

type postgresRepository struct {
	queries *sqlccart.Queries
}

func NewPostgresRepository(db sqlccart.DBTX) *postgresRepository {
	return &postgresRepository{queries: sqlccart.New(db)}
}

func (r *postgresRepository) AddItem(ctx context.Context, userID int64, item *entity.Item) error {
	return r.queries.AddItem(ctx, sqlccart.AddItemParams{
		UserID: userID,
		Sku:    int32(item.SKU),
		Count:  int32(item.Count),
	})
}

func (r *postgresRepository) GetItemsByUserID(ctx context.Context, userID int64) ([]*entity.Item, error) {
	rows, err := r.queries.ListItemsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]*entity.Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, &entity.Item{
			SKU:   uint32(row.Sku),
			Count: uint32(row.Count),
		})
	}
	return items, nil
}

func (r *postgresRepository) DeleteItemsByUserID(ctx context.Context, userID int64) error {
	return r.queries.DeleteItemsByUserID(ctx, userID)
}

func (r *postgresRepository) DeleteItem(ctx context.Context, userID int64, sku uint32) error {
	return r.queries.DeleteItem(ctx, sqlccart.DeleteItemParams{
		UserID: userID,
		Sku:    int32(sku),
	})
}
