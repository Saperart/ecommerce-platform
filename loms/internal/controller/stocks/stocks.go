package stocks

import (
	"context"

	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
)

//go:generate mockgen -source=stocks.go -destination=mocks/stocks_mock.go -package=mocks

type stocksService interface {
	GetStock(ctx context.Context, sku uint32) (uint64, error)
	SetStock(ctx context.Context, sku uint32, count uint64) error
}

type stocksServer struct {
	stockspb.UnimplementedStocksServer
	stocksService stocksService
}

func NewStocksServer(stocksService stocksService) *stocksServer {
	return &stocksServer{
		stocksService: stocksService,
	}
}
