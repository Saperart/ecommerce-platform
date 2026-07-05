package transactor

import (
	"context"
	"errors"
	"testing"

	mockstransactor "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTransactorWithTx(t *testing.T) {
	t.Parallel()

	errBegin := errors.New("begin error")
	errCallback := errors.New("callback error")
	errCommit := errors.New("commit error")

	tests := []struct {
		name       string
		setupMocks func(db *mockstransactor.MockDB, tx *mockstransactor.MockTx)
		callback   func(ctx context.Context) error
		wantErr    error
		wantMsg    string
	}{
		{
			name: "begin error",
			setupMocks: func(db *mockstransactor.MockDB, _ *mockstransactor.MockTx) {
				db.EXPECT().
					Begin(gomock.Any()).
					Return(nil, errBegin)
			},
			callback: func(context.Context) error {
				return nil
			},
			wantErr: errBegin,
			wantMsg: "begin tx",
		},
		{
			name: "callback error rolls back",
			setupMocks: func(db *mockstransactor.MockDB, tx *mockstransactor.MockTx) {
				db.EXPECT().
					Begin(gomock.Any()).
					Return(tx, nil)
				tx.EXPECT().
					Rollback(gomock.Any()).
					Return(nil)
			},
			callback: func(ctx context.Context) error {
				_, err := ExtractTx(ctx)
				require.NoError(t, err)
				return errCallback
			},
			wantErr: errCallback,
			wantMsg: "execute in tx",
		},
		{
			name: "commit error",
			setupMocks: func(db *mockstransactor.MockDB, tx *mockstransactor.MockTx) {
				db.EXPECT().
					Begin(gomock.Any()).
					Return(tx, nil)
				tx.EXPECT().
					Commit(gomock.Any()).
					Return(errCommit)
			},
			callback: func(ctx context.Context) error {
				_, err := ExtractTx(ctx)
				require.NoError(t, err)
				return nil
			},
			wantErr: errCommit,
			wantMsg: "commit tx",
		},
		{
			name: "success",
			setupMocks: func(db *mockstransactor.MockDB, tx *mockstransactor.MockTx) {
				db.EXPECT().
					Begin(gomock.Any()).
					Return(tx, nil)
				tx.EXPECT().
					Commit(gomock.Any()).
					Return(nil)
			},
			callback: func(ctx context.Context) error {
				_, err := ExtractTx(ctx)
				require.NoError(t, err)
				return nil
			},
			wantErr: nil,
			wantMsg: "",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			db := mockstransactor.NewMockDB(ctrl)
			tx := mockstransactor.NewMockTx(ctrl)
			test.setupMocks(db, tx)

			transactor := &Transactor{db: db}

			err := transactor.WithTx(context.Background(), test.callback)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				require.Contains(t, err.Error(), test.wantMsg)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestTransactorWithTxUsesExistingTx(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	db := mockstransactor.NewMockDB(ctrl)
	tx := mockstransactor.NewMockTx(ctrl)
	ctx := injectTx(context.Background(), tx)

	transactor := &Transactor{db: db}

	err := transactor.WithTx(ctx, func(ctx context.Context) error {
		extractedTx, err := ExtractTx(ctx)
		require.NoError(t, err)
		require.Equal(t, tx, extractedTx)
		return nil
	})

	require.NoError(t, err)
}

func TestInjectTx(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	tx := mockstransactor.NewMockTx(ctrl)

	ctx := injectTx(context.Background(), tx)

	extractedTx, err := ExtractTx(ctx)
	require.NoError(t, err)
	require.Equal(t, tx, extractedTx)
}

func TestExtractTxNotFound(t *testing.T) {
	t.Parallel()

	tx, err := ExtractTx(context.Background())

	require.Nil(t, tx)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrTxNotFound)
}
