package outbox

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	mocksoutbox "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/outbox/mocks"
	outboxrepo "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	repository := mocksoutbox.NewMockRepository(ctrl)
	handlerByKind := func(outboxrepo.Kind) (Handler, error) {
		return nil, nil
	}

	service := New(zap.NewNop(), repository, handlerByKind)

	require.NotNil(t, service)
	require.Equal(t, repository, service.repository)
	require.NotNil(t, service.logger)
	require.NotNil(t, service.handlerByKind)
}

func TestOutboxProcessBatch(t *testing.T) {
	t.Parallel()

	errClaim := errors.New("claim error")
	errKind := errors.New("kind error")
	errHandler := errors.New("handler error")
	errMark := errors.New("mark error")

	tests := []struct {
		name          string
		setupMocks    func(repository *mocksoutbox.MockRepository)
		handlerByKind HandlerByKind
	}{
		{
			name: "claim messages error",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(nil, errClaim)
			},
			handlerByKind: func(_ outboxrepo.Kind) (Handler, error) {
				return nil, nil
			},
		},
		{
			name: "empty batch",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(nil, nil)
			},
			handlerByKind: func(_ outboxrepo.Kind) (Handler, error) {
				return nil, nil
			},
		},
		{
			name: "success",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				messages := []outboxrepo.Message{
					{
						IdempotencyKey: "order-status:1:paid",
						Kind:           outboxrepo.KindNotification,
						Payload:        []byte(`{"order_id":1}`),
					},
					{
						IdempotencyKey: "order-status:2:cancelled",
						Kind:           outboxrepo.KindNotification,
						Payload:        []byte(`{"order_id":2}`),
					},
				}

				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(messages, nil)
				repository.EXPECT().
					MarkProcessed(gomock.Any(), []string{"order-status:1:paid", "order-status:2:cancelled"}).
					Return(nil)
				repository.EXPECT().
					MarkRetryable(gomock.Any(), []string{}).
					Return(nil)
			},
			handlerByKind: func(kind outboxrepo.Kind) (Handler, error) {
				if kind != outboxrepo.KindNotification {
					return nil, errKind
				}
				return func(_ context.Context, payload []byte) error {
					if len(payload) == 0 {
						return errors.New("empty payload")
					}
					return nil
				}, nil
			},
		},
		{
			name: "handler by kind error",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				messages := []outboxrepo.Message{
					{
						IdempotencyKey: "order-status:1:paid",
						Kind:           outboxrepo.Kind("unknown"),
						Payload:        []byte(`{"order_id":1}`),
					},
				}

				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(messages, nil)
				repository.EXPECT().
					MarkProcessed(gomock.Any(), []string{}).
					Return(nil)
				repository.EXPECT().
					MarkRetryable(gomock.Any(), []string{"order-status:1:paid"}).
					Return(nil)
			},
			handlerByKind: func(_ outboxrepo.Kind) (Handler, error) {
				return nil, errKind
			},
		},
		{
			name: "handler error",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				messages := []outboxrepo.Message{
					{
						IdempotencyKey: "order-status:1:paid",
						Kind:           outboxrepo.KindNotification,
						Payload:        []byte(`{"order_id":1}`),
					},
				}

				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(messages, nil)
				repository.EXPECT().
					MarkProcessed(gomock.Any(), []string{}).
					Return(nil)
				repository.EXPECT().
					MarkRetryable(gomock.Any(), []string{"order-status:1:paid"}).
					Return(nil)
			},
			handlerByKind: func(_ outboxrepo.Kind) (Handler, error) {
				return func(context.Context, []byte) error {
					return errHandler
				}, nil
			},
		},
		{
			name: "mark results errors",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				messages := []outboxrepo.Message{
					{
						IdempotencyKey: "order-status:1:paid",
						Kind:           outboxrepo.KindNotification,
						Payload:        []byte(`{"order_id":1}`),
					},
					{
						IdempotencyKey: "order-status:2:paid",
						Kind:           outboxrepo.KindNotification,
						Payload:        []byte(`{"order_id":2}`),
					},
				}

				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(messages, nil)
				repository.EXPECT().
					MarkProcessed(gomock.Any(), []string{"order-status:1:paid"}).
					Return(errMark)
				repository.EXPECT().
					MarkRetryable(gomock.Any(), []string{"order-status:2:paid"}).
					Return(errMark)
			},
			handlerByKind: func(_ outboxrepo.Kind) (Handler, error) {
				return func(_ context.Context, payload []byte) error {
					if string(payload) == `{"order_id":2}` {
						return errHandler
					}
					return nil
				}, nil
			},
		},
		{
			name: "mixed success unsupported and failed handler",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				messages := []outboxrepo.Message{
					{
						IdempotencyKey: "order-status:1:paid",
						Kind:           outboxrepo.KindNotification,
						Payload:        []byte(`{"order_id":1}`),
					},
					{
						IdempotencyKey: "order-status:2:failed",
						Kind:           outboxrepo.Kind("unsupported"),
						Payload:        []byte(`{"order_id":2}`),
					},
					{
						IdempotencyKey: "order-status:3:cancelled",
						Kind:           outboxrepo.KindNotification,
						Payload:        []byte(`{"order_id":3}`),
					},
				}

				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(messages, nil)
				repository.EXPECT().
					MarkProcessed(gomock.Any(), []string{"order-status:1:paid"}).
					Return(nil)
				repository.EXPECT().
					MarkRetryable(gomock.Any(), []string{"order-status:2:failed", "order-status:3:cancelled"}).
					Return(nil)
			},
			handlerByKind: func(kind outboxrepo.Kind) (Handler, error) {
				if kind != outboxrepo.KindNotification {
					return nil, errKind
				}
				return func(_ context.Context, payload []byte) error {
					if string(payload) == `{"order_id":3}` {
						return errHandler
					}
					return nil
				}, nil
			},
		},
		{
			name: "all failed and mark processed with empty slice",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				messages := []outboxrepo.Message{
					{
						IdempotencyKey: "order-status:10:paid",
						Kind:           outboxrepo.Kind("unsupported"),
						Payload:        []byte(`{"order_id":10}`),
					},
				}

				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(messages, nil)
				repository.EXPECT().
					MarkProcessed(gomock.Any(), []string{}).
					Return(nil)
				repository.EXPECT().
					MarkRetryable(gomock.Any(), []string{"order-status:10:paid"}).
					Return(nil)
			},
			handlerByKind: func(_ outboxrepo.Kind) (Handler, error) {
				return nil, errKind
			},
		},
		{
			name: "all success and mark retryable with empty slice",
			setupMocks: func(repository *mocksoutbox.MockRepository) {
				messages := []outboxrepo.Message{
					{
						IdempotencyKey: "order-status:11:paid",
						Kind:           outboxrepo.KindNotification,
						Payload:        []byte(`{"order_id":11}`),
					},
				}

				repository.EXPECT().
					ClaimMessages(gomock.Any(), 10, time.Second).
					Return(messages, nil)
				repository.EXPECT().
					MarkProcessed(gomock.Any(), []string{"order-status:11:paid"}).
					Return(nil)
				repository.EXPECT().
					MarkRetryable(gomock.Any(), []string{}).
					Return(nil)
			},
			handlerByKind: func(kind outboxrepo.Kind) (Handler, error) {
				if kind != outboxrepo.KindNotification {
					return nil, errKind
				}
				return func(_ context.Context, _ []byte) error {
					return nil
				}, nil
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			repository := mocksoutbox.NewMockRepository(ctrl)
			test.setupMocks(repository)

			service := New(zap.NewNop(), repository, test.handlerByKind)

			service.processBatch(context.Background(), 1, 10, time.Second)
		})
	}
}

func TestOutboxStartInvalidSettings(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	repository := mocksoutbox.NewMockRepository(ctrl)
	service := New(zap.NewNop(), repository, func(outboxrepo.Kind) (Handler, error) {
		return nil, nil
	})

	service.Start(context.Background(), 0, 10, time.Millisecond, time.Second)
	service.Start(context.Background(), 1, 0, time.Millisecond, time.Second)
	service.Start(context.Background(), 1, 10, 0, time.Second)
}

func TestOutboxStartProcessesBatch(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	repository := mocksoutbox.NewMockRepository(ctrl)
	repository.EXPECT().
		ClaimMessages(gomock.Any(), 1, time.Second).
		DoAndReturn(func(context.Context, int, time.Duration) ([]outboxrepo.Message, error) {
			cancel()
			return []outboxrepo.Message{
				{
					IdempotencyKey: "order-status:1001:paid",
					Kind:           outboxrepo.KindNotification,
					Payload:        []byte(`{"order_id":1001}`),
				},
			}, nil
		})
	repository.EXPECT().
		MarkProcessed(gomock.Any(), []string{"order-status:1001:paid"}).
		Return(nil)
	repository.EXPECT().
		MarkRetryable(gomock.Any(), []string{}).
		DoAndReturn(func(context.Context, []string) error {
			wg.Done()
			return nil
		})

	service := New(zap.NewNop(), repository, func(outboxrepo.Kind) (Handler, error) {
		return func(context.Context, []byte) error {
			return nil
		}, nil
	})

	service.Start(ctx, 1, 1, time.Millisecond, time.Second)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "outbox worker did not process batch")
	}
}

func TestOutboxWorkerStopsOnCanceledContext(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	repository := mocksoutbox.NewMockRepository(ctrl)
	service := New(zap.NewNop(), repository, func(outboxrepo.Kind) (Handler, error) {
		return nil, nil
	})

	var wg sync.WaitGroup
	wg.Add(1)

	done := make(chan struct{})
	go func() {
		service.worker(ctx, &wg, 1, 10, time.Millisecond, time.Second)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		require.Fail(t, "worker did not stop on canceled context")
	}
}

func TestOutboxProcessBatchNoMessagesAfterClaim(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)

	repository := mocksoutbox.NewMockRepository(ctrl)
	repository.EXPECT().
		ClaimMessages(gomock.Any(), 5, time.Second).
		Return([]outboxrepo.Message{}, nil)

	service := New(zap.NewNop(), repository, func(outboxrepo.Kind) (Handler, error) {
		return nil, nil
	})

	service.processBatch(context.Background(), 1, 5, time.Second)
}
