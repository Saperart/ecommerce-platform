package inmemory

type ItemRow struct {
	SKU   uint32 `db:"sku"`
	Count uint32 `db:"count"`
}
