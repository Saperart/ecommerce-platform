package app

import (
	"context"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/controller"
	notificationscontroller "github.com/igoroutine-courses/microservices.ecommerce.notifications/internal/controller/notifications"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const shutdownDelay = 100 * time.Millisecond

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	notificationsServer := notificationscontroller.NewNotificationsServer(logger, cfg)
	ctrl := controller.New(notificationsServer)

	go runGrpc(logger, cfg, ctrl)

	<-ctx.Done()
	time.Sleep(shutdownDelay)
}

func runGrpc(logger *zap.Logger, cfg *config.Config, ctrl *controller.API) {
	port := ":" + cfg.GRPC.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatal("cannot open tcp socket", zap.Error(err))
	}

	s := grpc.NewServer()
	notificationspb.RegisterNotificationsServer(s, ctrl.Notifications)

	logger.Info("notifications grpc server listening", zap.String("port", port))
	if err := s.Serve(lis); err != nil {
		logger.Fatal("notifications grpc server listen error", zap.Error(err))
	}
}
