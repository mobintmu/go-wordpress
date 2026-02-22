-- name: CreateWebsite :one
INSERT INTO websites (name, domain)
VALUES ($1, $2)
RETURNING id, name, domain;

-- name: GetWebsiteByID :one
SELECT id, name, domain, status, created_at, updated_at
FROM websites
WHERE id = $1;

-- name: UpdateWebsite :one
UPDATE websites
SET name = $2, domain = $3, status = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, name, domain, status, created_at, updated_at; 

-- name: DeleteWebsite :exec
DELETE FROM websites
WHERE id = $1;  

-- name: ListWebsites :many
SELECT id, name, domain, status, created_at, updated_at
FROM websites
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListWebsitesByStatus :many
SELECT id, name, domain, status, created_at, updated_at
FROM websites
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllWebsites :many
SELECT id, name, domain, status, created_at, updated_at
FROM websites
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

