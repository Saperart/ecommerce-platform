package grpc_test

import (
	"context"
	"testing"

	notificationsgrpc "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/adapter/notifications/grpc"
	mocksgrpc "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/adapter/notifications/grpc/mocks"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/port"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNotificationsClientSendOrderStatusChangedNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		notification port.OrderStatusChangedNotification
		setupMocks   func(client *mocksgrpc.MockNotificationsClient)
		wantErr      error
	}{
		{
			name: "send error",
			notification: port.OrderStatusChangedNotification{
				UserID:  42,
				OrderID: 1001,
				Status:  port.OrderStatusPaid,
			},
			setupMocks: func(client *mocksgrpc.MockNotificationsClient) {
				client.EXPECT().
					SendOrderStatusChangedNotification(
						gomock.Any(),
						&notificationspb.OrderStatusChangedNotificationRequest{
							UserId:  42,
							OrderId: 1001,
							Status:  notificationspb.OrderStatus_ORDER_STATUS_PAID,
						},
					).
					Return(nil, context.DeadlineExceeded)
			},
			wantErr: context.DeadlineExceeded,
		},
		{
			name: "success",
			notification: port.OrderStatusChangedNotification{
				UserID:  42,
				OrderID: 1001,
				Status:  port.OrderStatusAwaitingPayment,
			},
			setupMocks: func(client *mocksgrpc.MockNotificationsClient) {
				client.EXPECT().
					SendOrderStatusChangedNotification(
						gomock.Any(),
						&notificationspb.OrderStatusChangedNotificationRequest{
							UserId:  42,
							OrderId: 1001,
							Status:  notificationspb.OrderStatus_ORDER_STATUS_AWAITING_PAYMENT,
						},
					).
					Return(&emptypb.Empty{}, nil)
			},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			client := mocksgrpc.NewMockNotificationsClient(ctrl)
			test.setupMocks(client)

			adapter := notificationsgrpc.NewNotificationsClient(client)

			err := adapter.SendOrderStatusChangedNotification(context.Background(), test.notification)

			if test.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
