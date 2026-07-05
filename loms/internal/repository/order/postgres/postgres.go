package postgres

import (
	"context"
	"errors"

	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"
	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	sqlcorder "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order/postgres/sqlc"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor"
	"github.com/jackc/pgx/v5"
)

type postgresRepository struct {
	queries *sqlcorder.Queries
}

func NewPostgresRepository(db sqlcorder.DBTX) *postgresRepository {
	return &postgresRepository{queries: sqlcorder.New(db)}
}

func (r *postgresRepository) queriesFor(ctx context.Context) *sqlcorder.Queries {
	if tx, err := transactor.ExtractTx(ctx); err == nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *postgresRepository) CreateOrder(ctx context.Context, userID int64, items []entity.OrderItem) (int64, error) {
	q := r.queriesFor(ctx)
	order, err := q.CreateOrder(ctx, sqlcorder.CreateOrderParams{
		UserID: userID,
		Status: toSQLCStatus(entity.OrderStatusAwaitingPayment),
	})
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		if err := q.CreateOrderItem(ctx, sqlcorder.CreateOrderItemParams{
			OrderID: order.ID,
			Sku:     int32(item.SKU),
			Count:   int32(item.Count),
		}); err != nil {
			return 0, err
		}
	}

	return order.ID, nil
}

func (r *postgresRepository) DeleteOrder(ctx context.Context, orderID int64) error {
	return r.queriesFor(ctx).DeleteOrder(ctx, orderID)
}

func (r *postgresRepository) GetOrder(ctx context.Context, orderID int64) (*entity.Order, error) {
	q := r.queriesFor(ctx)
	order, err := q.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, xerrors.ErrOrderNotFound
		}
		return nil, err
	}

	items, err := q.ListOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	result := &entity.Order{
		ID:     order.ID,
		UserID: order.UserID,
		Status: fromSQLCStatus(order.Status),
		Items:  make([]entity.OrderItem, 0, len(items)),
	}
	if order.CreatedAt.Valid {
		result.CreatedAt = order.CreatedAt.Time
	}
	if order.UpdatedAt.Valid {
		result.UpdatedAt = order.UpdatedAt.Time
	}

	for _, item := range items {
		result.Items = append(result.Items, entity.OrderItem{
			SKU:   uint32(item.Sku),
			Count: uint32(item.Count),
		})
	}

	return result, nil
}

func (r *postgresRepository) SetOrderStatus(ctx context.Context, orderID int64, status entity.OrderStatus) error {
	if status == entity.OrderStatusPaid || status == entity.OrderStatusCancelled {
		return r.transitOrderStatus(ctx, orderID, entity.OrderStatusAwaitingPayment, status)
	}

	rows, err := r.queriesFor(ctx).SetOrderStatus(ctx, sqlcorder.SetOrderStatusParams{
		ID:     orderID,
		Status: toSQLCStatus(status),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return xerrors.ErrOrderNotFound
	}
	return nil
}

func (r *postgresRepository) transitOrderStatus(
	ctx context.Context,
	orderID int64,
	from entity.OrderStatus,
	to entity.OrderStatus,
) error {
	q := r.queriesFor(ctx)
	rows, err := q.TransitOrderStatus(ctx, sqlcorder.TransitOrderStatusParams{
		ID:       orderID,
		Status:   toSQLCStatus(from),
		Status_2: toSQLCStatus(to),
	})
	if err != nil {
		return err
	}
	if rows > 0 {
		return nil
	}

	if _, err := q.GetOrder(ctx, orderID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return xerrors.ErrOrderNotFound
		}
		return err
	}
	return xerrors.ErrInvalidOrderStatus
}

func toSQLCStatus(status entity.OrderStatus) sqlcorder.LomsOrderStatus {
	switch status {
	case entity.OrderStatusNew:
		return sqlcorder.LomsOrderStatusNew
	case entity.OrderStatusAwaitingPayment:
		return sqlcorder.LomsOrderStatusAwaitingPayment
	case entity.OrderStatusFailed:
		return sqlcorder.LomsOrderStatusFailed
	case entity.OrderStatusPaid:
		return sqlcorder.LomsOrderStatusPaid
	case entity.OrderStatusCancelled:
		return sqlcorder.LomsOrderStatusCancelled
	default:
		return sqlcorder.LomsOrderStatusNew
	}
}

func fromSQLCStatus(status sqlcorder.LomsOrderStatus) entity.OrderStatus {
	switch status {
	case sqlcorder.LomsOrderStatusNew:
		return entity.OrderStatusNew
	case sqlcorder.LomsOrderStatusAwaitingPayment:
		return entity.OrderStatusAwaitingPayment
	case sqlcorder.LomsOrderStatusFailed:
		return entity.OrderStatusFailed
	case sqlcorder.LomsOrderStatusPaid:
		return entity.OrderStatusPaid
	case sqlcorder.LomsOrderStatusCancelled:
		return entity.OrderStatusCancelled
	default:
		return entity.OrderStatusUnspecified
	}
}
