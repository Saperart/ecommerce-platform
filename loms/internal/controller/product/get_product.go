package product

import (
	"context"
	"errors"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (p *productServer) GetProduct(
	ctx context.Context, req *productpb.GetProductRequest,
) (*productpb.GetProductResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation: %v", err)
	}

	product, err := p.productService.GetProduct(ctx, req.GetSku())
	if err != nil {
		switch {
		case errors.Is(err, xerrors.ErrProductNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, xerrors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	return &productpb.GetProductResponse{Name: product.Name, Price: product.Price}, nil
}
