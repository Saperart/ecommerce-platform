-- name: CreateOrder :one
INSERT INTO loms.orders (user_id, status)
VALUES ($1, $2)
RETURNING id, user_id, status, created_at, updated_at;

-- name: CreateOrderItem :exec
INSERT INTO loms.order_items (order_id, sku, count)
VALUES ($1, $2, $3);

-- name: GetOrder :one
SELECT id, user_id, status, created_at, updated_at
FROM loms.orders
WHERE id = $1;

-- name: ListOrderItems :many
SELECT sku, count
FROM loms.order_items
WHERE order_id = $1
ORDER BY sku;

-- name: SetOrderStatus :execrows
UPDATE loms.orders
SET status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: TransitOrderStatus :execrows
UPDATE loms.orders
SET status = $3,
    updated_at = NOW()
WHERE id = $1
  AND status = $2;

-- name: DeleteOrder :exec
DELETE FROM loms.orders
WHERE id = $1;
