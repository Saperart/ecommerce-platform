package app

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	lomsadapter "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/loms/grpc"
	productadapter "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/adapter/product/grpc"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/config"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller"
	cartcontroller "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/controller/cart"
	cartpostgres "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/repository/cart/postgres"
	cartusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/cart"
	itemusecase "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/usecase/item"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/migrations"
	cartpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/cart/api/cart/v1"
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
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

	lomsConn, err := grpc.NewClient(
		cfg.Clients.LOMSGrpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatal("failed to connect to loms", zap.Error(err))
	}
	defer func() {
		_ = lomsConn.Close()
	}()

	productGrpcClient := productpb.NewProductServiceClient(lomsConn)
	stocksGrpcClient := stockspb.NewStocksClient(lomsConn)
	lomsGrpcClient := lomspb.NewLomsClient(lomsConn)

	productClient := productadapter.NewProductClient(productGrpcClient)
	lomsClient := lomsadapter.NewLOMSClient(stocksGrpcClient, lomsGrpcClient)

	pool := setupPostgres(ctx, logger, cfg)
	defer pool.Close()

	cartRepo := cartpostgres.NewPostgresRepository(pool)
	itemService := itemusecase.NewItemService(cartRepo, productClient, lomsClient)
	cartService := cartusecase.NewCartService(cartRepo, productClient, lomsClient)

	ctrl := controller.New(cartcontroller.NewCartServer(itemService, cartService))

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

	if err := cartpb.RegisterCartHandlerFromEndpoint(ctx, mux, endpoint, opts); err != nil {
		logger.Fatal("register gateway", zap.Error(err))
	}

	addr := ":" + cfg.GRPC.GatewayPort
	logger.Info("cart gateway", zap.String("addr", addr))
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
	cartpb.RegisterCartServer(s, ctrl.Cart)

	logger.Info("cart grpc server listening", zap.String("port", port))

	if err := s.Serve(lis); err != nil {
		logger.Fatal("cart grpc server listen error", zap.Error(err))
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
