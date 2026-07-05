-- name: AddItem :exec
INSERT INTO cart.items (user_id, sku, count)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, sku) DO UPDATE SET count = cart.items.count + EXCLUDED.count;

-- name: ListItemsByUserID :many
SELECT sku, count
FROM cart.items
WHERE user_id = $1
ORDER BY sku;

-- name: DeleteItemsByUserID :exec
DELETE FROM cart.items
WHERE user_id = $1;

-- name: DeleteItem :exec
DELETE FROM cart.items
WHERE user_id = $1
  AND sku = $2;
