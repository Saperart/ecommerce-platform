package grpc

import (
	"context"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/errors"
	"github.com/igoroutine-courses/microservices.ecommerce.cart/internal/port"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *productClient) GetProduct(ctx context.Context, sku uint32) (*port.ProductInfo, error) {
	response, err := p.client.GetProduct(ctx, &productpb.GetProductRequest{Sku: sku})
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return nil, xerrors.ErrProductNotFound
		case codes.InvalidArgument:
			return nil, xerrors.ErrInvalidInput
		default:
			return nil, err
		}
	}
	return &port.ProductInfo{
		Name:  response.GetName(),
		Price: response.GetPrice(),
	}, nil
}
