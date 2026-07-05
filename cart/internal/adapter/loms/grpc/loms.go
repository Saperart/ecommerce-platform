package grpc

import (
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
)

//go:generate mockgen -destination=mocks/loms_client_mock.go -package=mocks github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1 LomsClient
//go:generate mockgen -destination=mocks/stocks_client_mock.go -package=mocks github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1 StocksClient

type lomsClient struct {
	stocksClient stockspb.StocksClient
	lomsClient   lomspb.LomsClient
}

func NewLOMSClient(stocksClient stockspb.StocksClient, lomsC lomspb.LomsClient) *lomsClient {
	return &lomsClient{
		stocksClient: stocksClient,
		lomsClient:   lomsC,
	}
}
