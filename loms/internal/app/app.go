package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	notificationsadapter "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/adapter/notifications/grpc"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller"
	lomscontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/loms"
	productcontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/product"
	stockscontroller "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/controller/stocks"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/outbox"
	orderpostgres "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/order/postgres"
	outboxpostgres "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/outbox/postgres"
	productpostgres "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/product/postgres"
	stockspostgres "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/stocks/postgres"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/internal/repository/transactor"
	lomsusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/loms"
	productusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/product"
	stocksusecase "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/usecase/stocks"
	"github.com/igoroutine-courses/microservices.ecommerce.loms/migrations"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
	notificationspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/notifications/api/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const shutdownDelay = 100 * time.Millisecond

func Run(logger *zap.Logger, cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool := setupPostgres(ctx, logger, cfg)
	defer pool.Close()

	productRepo := productpostgres.NewPostgresRepository(pool)
	stockRepo := stockspostgres.NewPostgresRepository(pool)
	orderRepo := orderpostgres.NewPostgresRepository(pool)
	outboxRepository := outboxpostgres.NewPostgresRepository(pool)
	tx := transactor.New(pool)

	notificationsConn, err := grpc.NewClient(
		cfg.Clients.NotificationsGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("failed to connect to notifications", zap.Error(err))
	}
	defer func() {
		_ = notificationsConn.Close()
	}()

	notificationsClient := notificationsadapter.NewNotificationsClient(
		notificationspb.NewNotificationsClient(notificationsConn),
	)

	productUC := productusecase.NewProductService(productRepo)
	stocksUC := stocksusecase.NewStocksService(stockRepo)
	lomsUC := lomsusecase.NewLomsService(
		orderRepo,
		stockRepo,
		lomsusecase.WithTransactor(tx),
		lomsusecase.WithOutboxRepository(outboxRepository),
		lomsusecase.WithNotificationsClient(notificationsClient),
	)

	ctrl := controller.New(
		productcontroller.NewProductServer(productUC),
		stockscontroller.NewStocksServer(stocksUC),
		lomscontroller.NewLomsServer(lomsUC),
	)

	outboxCore := outbox.New(logger, outboxRepository, func(kind outboxpostgres.Kind) (outbox.Handler, error) {
		switch kind {
		case outboxpostgres.KindNotification:
			return lomsUC.OrderStatusChangedNotificationKindHandler, nil
		default:
			return nil, fmt.Errorf("unsupported outbox kind: %s", kind)
		}
	})
	outboxCore.Start(
		ctx,
		cfg.Outbox.Workers,
		cfg.Outbox.BatchSize,
		cfg.Outbox.FetchPeriod,
		cfg.Outbox.InProgressTTL,
	)

	go runGrpc(logger, cfg, ctrl)
	go runRest(ctx, logger, cfg)

	<-ctx.Done()
	time.Sleep(shutdownDelay)
}

func setupPostgres(ctx context.Context, logger *zap.Logger, cfg *config.Config) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, cfg.PostgresDSN())
	if err != nil {
		logger.Fatal("connect postgres", zap.Error(err))
	}
	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("ping postgres", zap.Error(err))
	}

	db := stdlib.OpenDBFromPool(pool)
	if err := migrations.Up(db); err != nil {
		logger.Fatal("run migrations", zap.Error(err))
	}
	return pool
}

func runRest(ctx context.Context, logger *zap.Logger, cfg *config.Config) {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := net.JoinHostPort("localhost", cfg.GRPC.Port)

	if err := lomspb.RegisterLomsHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		logger.Fatal("register loms gateway", zap.Error(err))
	}
	if err := productpb.RegisterProductServiceHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		logger.Fatal("register product gateway", zap.Error(err))
	}
	if err := stockspb.RegisterStocksHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		logger.Fatal("register stocks gateway", zap.Error(err))
	}

	addr := ":" + cfg.GRPC.GatewayPort
	logger.Info("loms gateway", zap.String("addr", addr))
	if err := http.ListenAndServe(addr, corsHandler(mux)); err != nil {
		logger.Error("gateway serve", zap.Error(err))
		os.Exit(1)
	}
}

func runGrpc(logger *zap.Logger, cfg *config.Config, ctrl *controller.API) {
	port := ":" + cfg.GRPC.Port

	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Fatal("cannot open tcp socket", zap.Error(err))
	}

	s := grpc.NewServer()

	productpb.RegisterProductServiceServer(s, ctrl.Product)
	stockspb.RegisterStocksServer(s, ctrl.Stocks)
	lomspb.RegisterLomsServer(s, ctrl.Loms)

	logger.Info("grpc server listening", zap.String("port", port))
	if err := s.Serve(lis); err != nil {
		logger.Fatal("grpc server listen error", zap.Error(err))
	}
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:5173"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
