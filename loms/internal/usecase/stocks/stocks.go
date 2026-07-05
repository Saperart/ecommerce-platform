package stocks

import (
	"context"
)

//go:generate mockgen -source=stocks.go -destination=mocks/stocks_mock.go -package=mocks

type (
	stocksRepository interface {
		GetStock(ctx context.Context, sku uint32) (uint64, error)
		SetStock(ctx context.Context, sku uint32, count uint64) error
	}
)

type stocksService struct {
	stocksRepository stocksRepository
}

func NewStocksService(stocksRepository stocksRepository) *stocksService {
	return &stocksService{
		stocksRepository: stocksRepository,
	}
}
