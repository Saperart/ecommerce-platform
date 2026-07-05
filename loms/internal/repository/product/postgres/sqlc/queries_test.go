package sqlc_test

import (
	"context"
	"testing"

	sqlcproduct "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/product/postgres/sqlc"
	mockssqlc "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/product/postgres/sqlc/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestQueriesCreateProduct(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	db := mockssqlc.NewMockDBTX(ctrl)
	row := mockssqlc.NewMockRow(ctrl)

	db.EXPECT().
		QueryRow(gomock.Any(), gomock.Any(), "phone", int32(100)).
		Return(row)
	row.EXPECT().
		Scan(gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(dest ...any) error {
			*(dest[0].(*int32)) = 10
			*(dest[1].(*string)) = "phone"
			*(dest[2].(*int32)) = 100
			return nil
		})

	product, err := sqlcproduct.New(db).CreateProduct(context.Background(), sqlcproduct.CreateProductParams{
		Name:  "phone",
		Price: 100,
	})

	require.NoError(t, err)
	require.Equal(t, sqlcproduct.CreateProductRow{Sku: 10, Name: "phone", Price: 100}, product)
}

func TestQueriesGetProduct(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow)
		wantRow    sqlcproduct.GetProductRow
		wantErr    error
	}{
		{
			name: "scan error",
			setupMocks: func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow) {
				db.EXPECT().
					QueryRow(gomock.Any(), gomock.Any(), int32(10)).
					Return(row)
				row.EXPECT().
					Scan(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(pgx.ErrNoRows)
			},
			wantRow: sqlcproduct.GetProductRow{},
			wantErr: pgx.ErrNoRows,
		},
		{
			name: "success",
			setupMocks: func(db *mockssqlc.MockDBTX, row *mockssqlc.MockRow) {
				db.EXPECT().
					QueryRow(gomock.Any(), gomock.Any(), int32(10)).
					Return(row)
				row.EXPECT().
					Scan(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(dest ...any) error {
						*(dest[0].(*int32)) = 10
						*(dest[1].(*string)) = "phone"
						*(dest[2].(*int32)) = 100
						return nil
					})
			},
			wantRow: sqlcproduct.GetProductRow{Sku: 10, Name: "phone", Price: 100},
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

			product, err := sqlcproduct.New(db).GetProduct(context.Background(), 10)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Equal(t, test.wantRow, product)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantRow, product)
		})
	}
}
