package sqlc_test

import (
	"context"
	"testing"

	sqlcstocks "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/stocks/postgres/sqlc"
	mockssqlc "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/stocks/postgres/sqlc/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestQueriesGetStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow)
		wantCount  int64
		wantErr    error
	}{
		{
			name: "scan error",
			setupMocks: func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow) {
				db.EXPECT().
					QueryRow(gomock.Any(), gomock.Any(), int32(10)).
					Return(row)
				row.EXPECT().
					Scan(gomock.Any()).
					Return(pgx.ErrNoRows)
			},
			wantCount: 0,
			wantErr:   pgx.ErrNoRows,
		},
		{
			name: "success",
			setupMocks: func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow) {
				db.EXPECT().
					QueryRow(gomock.Any(), gomock.Any(), int32(10)).
					Return(row)
				row.EXPECT().
					Scan(gomock.Any()).
					DoAndReturn(func(dest ...any) error {
						*(dest[0].(*int64)) = 5
						return nil
					})
			},
			wantCount: 5,
			wantErr:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			db := mockssqlc.NewMockDBTX(ctrl)
			row := mockssqlc.NewMockRow(ctrl)
			test.setupMocks(db, row)

			count, err := sqlcstocks.New(db).GetStock(context.Background(), 10)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantCount, count)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantCount, count)
		})
	}
}

func TestQueriesStockExecCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		run        func(queries *sqlcstocks.Queries) (int64, error)
		setupMocks func(db *mockssqlc.MockDBTX)
		wantRows   int64
		wantErr    error
	}{
		{
			name: "release stock success",
			run: func(queries *sqlcstocks.Queries) (int64, error) {
				return queries.ReleaseStock(context.Background(), sqlcstocks.ReleaseStockParams{Sku: 10, Count: 2})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), int32(10), int64(2)).
					Return(pgconn.NewCommandTag("UPDATE 1"), nil)
			},
			wantRows: 1,
			wantErr:  nil,
		},
		{
			name: "release stock error",
			run: func(queries *sqlcstocks.Queries) (int64, error) {
				return queries.ReleaseStock(context.Background(), sqlcstocks.ReleaseStockParams{Sku: 10, Count: 2})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), int32(10), int64(2)).
					Return(pgconn.CommandTag{}, context.DeadlineExceeded)
			},
			wantRows: 0,
			wantErr:  context.DeadlineExceeded,
		},
		{
			name: "reserve stock success",
			run: func(queries *sqlcstocks.Queries) (int64, error) {
				return queries.ReserveStock(context.Background(), sqlcstocks.ReserveStockParams{Sku: 10, Count: 2})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), int32(10), int64(2)).
					Return(pgconn.NewCommandTag("UPDATE 1"), nil)
			},
			wantRows: 1,
			wantErr:  nil,
		},
		{
			name: "upsert stock success",
			run: func(queries *sqlcstocks.Queries) (int64, error) {
				return 0, queries.UpsertStock(context.Background(), sqlcstocks.UpsertStockParams{Sku: 10, Count: 2})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), int32(10), int64(2)).
					Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
			},
			wantRows: 0,
			wantErr:  nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			db := mockssqlc.NewMockDBTX(ctrl)
			test.setupMocks(db)

			rows, err := test.run(sqlcstocks.New(db))

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
