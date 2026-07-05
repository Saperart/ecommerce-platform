package inmemory

type RowProduct struct {
	SKU   uint32 `db:"sku"`
	Name  string `db:"name"`
	Price uint32 `db:"price"`
}
