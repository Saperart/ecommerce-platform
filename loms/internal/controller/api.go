package controller

import (
	lomspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/loms/v1"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	stockspb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/stocks/v1"
)

type API struct {
	Product productpb.ProductServiceServer
	Stocks  stockspb.StocksServer
	Loms    lomspb.LomsServer
}

func New(product productpb.ProductServiceServer, stocks stockspb.StocksServer, loms lomspb.LomsServer) *API {
	return &API{
		Product: product,
		Stocks:  stocks,
		Loms:    loms,
	}
}
