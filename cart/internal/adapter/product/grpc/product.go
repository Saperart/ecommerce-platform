package grpc

import (
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
)

//go:generate mockgen -destination=mocks/product_mock.go -package=mocks github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1 ProductServiceClient
type productClient struct {
	client productpb.ProductServiceClient
}

func NewProductClient(client productpb.ProductServiceClient) *productClient {
	return &productClient{
		client: client,
	}
}
