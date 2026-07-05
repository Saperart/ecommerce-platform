package sqlc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	sqlcorder "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order/postgres/sqlc"
	mockssqlc "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order/postgres/sqlc/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestQueriesCreateOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow)
		wantOrder  sqlcorder.LomsOrder
		wantErr    error
	}{
		{
			name: "scan error",
			setupMocks: func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow) {
				db.EXPECT().
					QueryRow(gomock.Any(), gomock.Any(), int64(42), sqlcorder.LomsOrderStatusAwaitingPayment).
					Return(row)
				row.EXPECT().
					Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(context.DeadlineExceeded)
			},
			wantOrder: sqlcorder.LomsOrder{},
			wantErr:   context.DeadlineExceeded,
		},
		{
			name: "success",
			setupMocks: func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow) {
				db.EXPECT().
					QueryRow(gomock.Any(), gomock.Any(), int64(42), sqlcorder.LomsOrderStatusAwaitingPayment).
					Return(row)
				row.EXPECT().
					Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(dest ...any) error {
						*(dest[0].(*int64)) = 1001
						*(dest[1].(*int64)) = 42
						*(dest[2].(*sqlcorder.LomsOrderStatus)) = sqlcorder.LomsOrderStatusAwaitingPayment
						*(dest[3].(*pgtype.Timestamptz)) = pgtype.Timestamptz{Time: time.Unix(10, 0), Valid: true}
						*(dest[4].(*pgtype.Timestamptz)) = pgtype.Timestamptz{Time: time.Unix(20, 0), Valid: true}
						return nil
					})
			},
			wantOrder: sqlcorder.LomsOrder{
				ID:        1001,
				UserID:    42,
				Status:    sqlcorder.LomsOrderStatusAwaitingPayment,
				CreatedAt: pgtype.Timestamptz{Time: time.Unix(10, 0), Valid: true},
				UpdatedAt: pgtype.Timestamptz{Time: time.Unix(20, 0), Valid: true},
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			db := mockssqlc.NewMockDBTX(ctrl)
			row := mockssqlc.NewMockRow(ctrl)
			test.setupMocks(db, row)

			order, err := sqlcorder.New(db).CreateOrder(context.Background(), sqlcorder.CreateOrderParams{
				UserID: 42,
				Status: sqlcorder.LomsOrderStatusAwaitingPayment,
			})

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantOrder, order)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantOrder, order)
		})
	}
}

func TestQueriesExecCommands(t *testing.T) {
	t.Parallel()

	errExec := errors.New("exec error")

	tests := []struct {
		name       string
		run        func(queries *sqlcorder.Queries) (int64, error)
		setupMocks func(db *mockssqlc.MockDBTX)
		wantRows   int64
		wantErr    error
	}{
		{
			name: "create order item error",
			run: func(queries *sqlcorder.Queries) (int64, error) {
				return 0, queries.CreateOrderItem(context.Background(), sqlcorder.CreateOrderItemParams{
					OrderID: 1001,
					Sku:     10,
					Count:   2,
				})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), int64(1001), int32(10), int32(2)).
					Return(pgconn.CommandTag{}, errExec)
			},
			wantRows: 0,
			wantErr:  errExec,
		},
		{
			name: "delete order success",
			run: func(queries *sqlcorder.Queries) (int64, error) {
				return 0, queries.DeleteOrder(context.Background(), 1001)
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), int64(1001)).
					Return(pgconn.NewCommandTag("DELETE 1"), nil)
			},
			wantRows: 0,
			wantErr:  nil,
		},
		{
			name: "set order status success",
			run: func(queries *sqlcorder.Queries) (int64, error) {
				return queries.SetOrderStatus(context.Background(), sqlcorder.SetOrderStatusParams{
					ID:     1001,
					Status: sqlcorder.LomsOrderStatusPaid,
				})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), int64(1001), sqlcorder.LomsOrderStatusPaid).
					Return(pgconn.NewCommandTag("UPDATE 1"), nil)
			},
			wantRows: 1,
			wantErr:  nil,
		},
		{
			name: "set order status error",
			run: func(queries *sqlcorder.Queries) (int64, error) {
				return queries.SetOrderStatus(context.Background(), sqlcorder.SetOrderStatusParams{
					ID:     1001,
					Status: sqlcorder.LomsOrderStatusPaid,
				})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), int64(1001), sqlcorder.LomsOrderStatusPaid).
					Return(pgconn.CommandTag{}, errExec)
			},
			wantRows: 0,
			wantErr:  errExec,
		},
		{
			name: "transit order status success",
			run: func(queries *sqlcorder.Queries) (int64, error) {
				return queries.TransitOrderStatus(context.Background(), sqlcorder.TransitOrderStatusParams{
					ID:       1001,
					Status:   sqlcorder.LomsOrderStatusAwaitingPayment,
					Status_2: sqlcorder.LomsOrderStatusCancelled,
				})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(
						gomock.Any(),
						gomock.Any(),
						int64(1001),
						sqlcorder.LomsOrderStatusAwaitingPayment,
						sqlcorder.LomsOrderStatusCancelled,
					).
					Return(pgconn.NewCommandTag("UPDATE 1"), nil)
			},
			wantRows: 1,
			wantErr:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			db := mockssqlc.NewMockDBTX(ctrl)
			test.setupMocks(db)

			rows, err := test.run(sqlcorder.New(db))

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantRows, rows)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantRows, rows)
		})
	}
}

func TestQueriesGetOrder(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	db := mockssqlc.NewMockDBTX(ctrl)
	row := mockssqlc.NewMockRow(ctrl)

	db.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), int64(1001)).
		Return(row)
	row.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(pgx.ErrNoRows)

	order, err := sqlcorder.New(db).GetOrder(context.Background(), 1001)

	require.ErrorIs(t, err, pgx.ErrNoRows)
	require.Equal(t, sqlcorder.LomsOrder{}, order)
}

func TestQueriesListOrderItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(db *mockssqlc.MockDBTX, rows *mockssqlc.MockRows)
		wantItems  []sqlcorder.ListOrderItemsRow
		wantErr    error
	}{
		{
			name: "query error",
			setupMocks: func(db *mockssqlc.MockDBTX, _ *mockssqlc.MockRows) {
				db.EXPECT().
					Query(gomock.Any(), gomock.Any(), int64(1001)).
					Return(nil, context.DeadlineExceeded)
			},
			wantItems: nil,
			wantErr:   context.DeadlineExceeded,
		},
		{
			name: "scan error",
			setupMocks: func(db *mockssqlc.MockDBTX, rows *mockssqlc.MockRows) {
				db.EXPECT().
					Query(gomock.Any(), gomock.Any(), int64(1001)).
					Return(rows, nil)
				rows.EXPECT().Close()
				rows.EXPECT().Next().Return(true)
				rows.EXPECT().Scan(gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
			},
			wantItems: nil,
			wantErr:   context.DeadlineExceeded,
		},
		{
			name: "rows error",
			setupMocks: func(db *mockssqlc.MockDBTX, rows *mockssqlc.MockRows) {
				db.EXPECT().
					Query(gomock.Any(), gomock.Any(), int64(1001)).
					Return(rows, nil)
				rows.EXPECT().Close()
				rows.EXPECT().Next().Return(false)
				rows.EXPECT().Err().Return(context.DeadlineExceeded)
			},
			wantItems: nil,
			wantErr:   context.DeadlineExceeded,
		},
		{
			name: "success",
			setupMocks: func(db *mockssqlc.MockDBTX, rows *mockssqlc.MockRows) {
				db.EXPECT().
					Query(gomock.Any(), gomock.Any(), int64(1001)).
					Return(rows, nil)
				rows.EXPECT().Close()
				gomock.InOrder(
					rows.EXPECT().Next().Return(true),
					rows.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...any) error {
						*(dest[0].(*int32)) = 10
						*(dest[1].(*int32)) = 2
						return nil
					}),
					rows.EXPECT().Next().Return(false),
				)
				rows.EXPECT().Err().Return(nil)
			},
			wantItems: []sqlcorder.ListOrderItemsRow{{Sku: 10, Count: 2}},
			wantErr:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			db := mockssqlc.NewMockDBTX(ctrl)
			rows := mockssqlc.NewMockRows(ctrl)
			test.setupMocks(db, rows)

			items, err := sqlcorder.New(db).ListOrderItems(context.Background(), 1001)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantItems, items)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantItems, items)
		})
	}
}
