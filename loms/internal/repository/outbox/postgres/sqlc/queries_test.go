package sqlc_test

import (
	"context"
	"testing"

	sqlcoutbox "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres/sqlc"
	mockssqlc "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres/sqlc/mocks"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestQueriesClaimMessages(t *testing.T) {
	t.Parallel()

	params := sqlcoutbox.ClaimMessagesParams{
		InProgressTtl: pgtype.Interval{Microseconds: 10, Valid: true},
		BatchSize:     10,
	}

	tests := []struct {
		name       string
		setupMocks func(db *mockssqlc.MockDBTX, rows *mockssqlc.MockRows)
		wantItems  []sqlcoutbox.ClaimMessagesRow
		wantErr    error
	}{
		{
			name: "query error",
			setupMocks: func(db *mockssqlc.MockDBTX, _ *mockssqlc.MockRows) {
				db.EXPECT().
					Query(gomock.Any(), gomock.Any(), params.InProgressTtl, params.BatchSize).
					Return(nil, context.DeadlineExceeded)
			},
			wantItems: nil,
			wantErr:   context.DeadlineExceeded,
		},
		{
			name: "scan error",
			setupMocks: func(db *mockssqlc.MockDBTX, rows *mockssqlc.MockRows) {
				db.EXPECT().
					Query(gomock.Any(), gomock.Any(), params.InProgressTtl, params.BatchSize).
					Return(rows, nil)
				rows.EXPECT().Close()
				rows.EXPECT().Next().Return(true)
				rows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
			},
			wantItems: nil,
			wantErr:   context.DeadlineExceeded,
		},
		{
			name: "rows error",
			setupMocks: func(db *mockssqlc.MockDBTX, rows *mockssqlc.MockRows) {
				db.EXPECT().
					Query(gomock.Any(), gomock.Any(), params.InProgressTtl, params.BatchSize).
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
					Query(gomock.Any(), gomock.Any(), params.InProgressTtl, params.BatchSize).
					Return(rows, nil)
				rows.EXPECT().Close()
				gomock.InOrder(
					rows.EXPECT().Next().Return(true),
					rows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(dest ...any) error {
						*(dest[0].(*string)) = "order-status:1001:paid"
						*(dest[1].(*string)) = "notification"
						*(dest[2].(*[]byte)) = []byte(`{"order_id":1001}`)
						return nil
					}),
					rows.EXPECT().Next().Return(false),
				)
				rows.EXPECT().Err().Return(nil)
			},
			wantItems: []sqlcoutbox.ClaimMessagesRow{
				{
					IdempotencyKey: "order-status:1001:paid",
					Kind:           "notification",
					Payload:        []byte(`{"order_id":1001}`),
				},
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			db := mockssqlc.NewMockDBTX(ctrl)
			rows := mockssqlc.NewMockRows(ctrl)
			test.setupMocks(db, rows)

			items, err := sqlcoutbox.New(db).ClaimMessages(context.Background(), params)

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

func TestQueriesOutboxExecCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		run        func(queries *sqlcoutbox.Queries) error
		setupMocks func(db *mockssqlc.MockDBTX)
		wantErr    error
	}{
		{
			name: "save message success",
			run: func(queries *sqlcoutbox.Queries) error {
				return queries.SaveMessage(context.Background(), sqlcoutbox.SaveMessageParams{
					IdempotencyKey: "order-status:1001:paid",
					Kind:           "notification",
					Payload:        []byte(`{"order_id":1001}`),
				})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(
						gomock.Any(),
						gomock.Any(),
						"order-status:1001:paid",
						"notification",
						[]byte(`{"order_id":1001}`),
					).
					Return(pgconn.NewCommandTag("INSERT 0 1"), nil)
			},
			wantErr: nil,
		},
		{
			name: "save message error",
			run: func(queries *sqlcoutbox.Queries) error {
				return queries.SaveMessage(context.Background(), sqlcoutbox.SaveMessageParams{
					IdempotencyKey: "order-status:1001:paid",
					Kind:           "notification",
					Payload:        []byte(`{"order_id":1001}`),
				})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(
						gomock.Any(),
						gomock.Any(),
						"order-status:1001:paid",
						"notification",
						[]byte(`{"order_id":1001}`),
					).
					Return(pgconn.CommandTag{}, context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "mark processed success",
			run: func(queries *sqlcoutbox.Queries) error {
				return queries.MarkProcessed(context.Background(), []string{"order-status:1001:paid"})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), []string{"order-status:1001:paid"}).
					Return(pgconn.NewCommandTag("UPDATE 1"), nil)
			},
			wantErr: nil,
		},
		{
			name: "mark retryable success",
			run: func(queries *sqlcoutbox.Queries) error {
				return queries.MarkRetryable(context.Background(), []string{"order-status:1001:paid"})
			},
			setupMocks: func(db *mockssqlc.MockDBTX) {
				db.EXPECT().
					Exec(gomock.Any(), gomock.Any(), []string{"order-status:1001:paid"}).
					Return(pgconn.NewCommandTag("UPDATE 1"), nil)
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			db := mockssqlc.NewMockDBTX(ctrl)
			test.setupMocks(db)

			err := test.run(sqlcoutbox.New(db))

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
