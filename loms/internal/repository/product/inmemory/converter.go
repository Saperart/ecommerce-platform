package inmemory

import "github.com/igoroutine-courses/microservices.ecommerce.loms/internal/entity"

func rowToProduct(r *RowProduct) *entity.Product {
	if r == nil {
		return nil
	}
	return &entity.Product{
		SKU:   r.SKU,
		Name:  r.Name,
		Price: r.Price,
	}
}
