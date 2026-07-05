-- name: SaveMessage :exec
INSERT INTO loms.outbox (idempotency_key, kind, payload)
VALUES (
    $1,
    $2,
    sqlc.arg(payload)::jsonb
)
ON CONFLICT (idempotency_key) DO NOTHING;

-- name: ClaimMessages :many
UPDATE loms.outbox
SET status = 'in_progress',
    attempts = attempts + 1,
    updated_at = NOW()
WHERE id IN (
    SELECT id
    FROM loms.outbox
    WHERE status = 'created'
       OR (
            status = 'in_progress'
            AND updated_at + sqlc.arg(in_progress_ttl)::interval < NOW()
       )
    ORDER BY created_at
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE SKIP LOCKED
)
RETURNING idempotency_key, kind, payload;

-- name: MarkProcessed :exec
UPDATE loms.outbox
SET status = 'processed',
    updated_at = NOW()
WHERE idempotency_key = ANY($1::text[]);

-- name: MarkRetryable :exec
UPDATE loms.outbox
SET status = 'created',
    updated_at = NOW()
WHERE idempotency_key = ANY($1::text[]);
