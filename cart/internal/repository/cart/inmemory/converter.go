package inmemory

import "github.com/igoroutine-courses/microservices.ecommerce.cart/internal/entity"

func rowToEntity(r *ItemRow) *entity.Item {
	if r == nil {
		return nil
	}
	return &entity.Item{
		SKU:   r.SKU,
		Count: r.Count,
	}
}

func entityToRow(o *entity.Item) *ItemRow {
	if o == nil {
		return nil
	}
	return &ItemRow{
		SKU:   o.SKU,
		Count: o.Count,
	}
}
