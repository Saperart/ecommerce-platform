-- name: GetStock :one
SELECT count
FROM loms.stocks
WHERE sku = $1;

-- name: ReserveStock :execrows
UPDATE loms.stocks
SET count = count - $2,
    updated_at = NOW()
WHERE sku = $1
  AND count >= $2;

-- name: ReleaseStock :execrows
UPDATE loms.stocks
SET count = count + $2
WHERE sku = $1;

-- name: UpsertStock :exec
INSERT INTO loms.stocks (sku, count)
VALUES ($1, $2)
ON CONFLICT (sku) DO UPDATE SET count = EXCLUDED.count;
