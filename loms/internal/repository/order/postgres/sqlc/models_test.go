package sqlc_test

import (
	"testing"

	sqlcorder "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order/postgres/sqlc"
	"github.com/stretchr/testify/require"
)

func TestLomsOrderStatusScan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		src        any
		wantStatus sqlcorder.LomsOrderStatus
		wantErr    bool
	}{
		{name: "bytes", src: []byte("paid"), wantStatus: sqlcorder.LomsOrderStatusPaid, wantErr: false},
		{name: "string", src: "cancelled", wantStatus: sqlcorder.LomsOrderStatusCancelled, wantErr: false},
		{name: "unsupported", src: 1, wantStatus: "", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var status sqlcorder.LomsOrderStatus
			err := status.Scan(test.src)

			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.wantStatus, status)
		})
	}
}

func TestNullLomsOrderStatus(t *testing.T) {
	t.Parallel()

	var status sqlcorder.NullLomsOrderStatus

	err := status.Scan(nil)

	require.NoError(t, err)
	require.False(t, status.Valid)

	err = status.Scan("paid")

	require.NoError(t, err)
	require.True(t, status.Valid)
	require.Equal(t, sqlcorder.LomsOrderStatusPaid, status.LomsOrderStatus)

	value, err := status.Value()

	require.NoError(t, err)
	require.Equal(t, "paid", value)

	value, err = (sqlcorder.NullLomsOrderStatus{}).Value()

	require.NoError(t, err)
	require.Nil(t, value)
}
