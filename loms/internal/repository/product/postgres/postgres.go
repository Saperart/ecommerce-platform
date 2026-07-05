package postgres

import (
	"context"
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	sqlcproduct "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/product/postgres/sqlc"
	"github.com/jackc/pgx/v5"
)

type postgresRepository struct {
	queries *sqlcproduct.Queries
}

func NewPostgresRepository(db sqlcproduct.DBTX) *postgresRepository {
	return &postgresRepository{queries: sqlcproduct.New(db)}
}

func (r *postgresRepository) GetProduct(ctx context.Context, sku uint32) (*entity.Product, error) {
	row, err := r.queries.GetProduct(ctx, int32(sku))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrProductNotFound
		}
		return nil, err
	}
	return &entity.Product{
		SKU:   uint32(row.Sku),
		Name:  row.Name,
		Price: uint32(row.Price),
	}, nil
}

func (r *postgresRepository) CreateProduct(ctx context.Context, name string, price uint32) (uint32, error) {
	row, err := r.queries.CreateProduct(ctx, sqlcproduct.CreateProductParams{
		Name:  name,
		Price: int32(price),
	})
	if err != nil {
		return 0, err
	}
	return uint32(row.Sku), nil
}
