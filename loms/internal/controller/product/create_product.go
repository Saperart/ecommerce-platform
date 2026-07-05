package product

import (
	"context"
	"errors"

	xerrors "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/errors"
	productpb "github.com/igoroutine-courses/microservices.ecommerce.pkg/generated/loms/api/product/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *productServer) CreateProduct(
	ctx context.Context,
	req *productpb.CreateProductRequest,
) (*productpb.CreateProductResponse, error) {
	if err := req.ValidateAll(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation: %v", err)
	}

	sku, err := s.productService.CreateProduct(ctx, req.GetName(), req.GetPrice())
	if err != nil {
		switch {
		case errors.Is(err, xerrors.ErrInvalidInput):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &productpb.CreateProductResponse{Sku: sku}, nil
}
