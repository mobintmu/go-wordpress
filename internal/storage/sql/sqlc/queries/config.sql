-- name: CreateConfig :one
INSERT INTO configs (website_id, key, value)
VALUES ($1, $2, $3)
RETURNING id, website_id, key, value, created_at, updated_at;

-- name: GetConfigByID :one
SELECT id, website_id, key, value, created_at, updated_at
FROM configs
WHERE id = $1;

-- name: UpdateConfig :one
UPDATE configs
SET website_id = $2, key = $3, value = $4, status = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, website_id, key, value, created_at, updated_at;

-- name: DeleteConfig :exec
DELETE FROM configs
WHERE id = $1;  

-- name: ListConfigsByWebsiteID :many
SELECT id, website_id, key, value, created_at, updated_at
FROM configs
WHERE website_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveConfigsByWebsiteID :many
SELECT id, website_id, key, value, created_at, updated_at
FROM configs
WHERE website_id = $1 AND status = 'active'
ORDER BY created_at DESC    
LIMIT $2 OFFSET $3;

-- name: ListConfigsByStatus :many
SELECT id, website_id, key, value, created_at, updated_at
FROM configs
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveConfigsByKey :many
SELECT id, website_id, key, value, created_at, updated_at
FROM configs
WHERE key = $1 AND status = 'active'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllConfigs :many
SELECT id, website_id, key, value, created_at, updated_at
FROM configs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2; 

