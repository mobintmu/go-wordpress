-- name: CreateCategory :one
INSERT INTO categories (website_id, name, link)
VALUES ($1, $2, $3)
RETURNING id, website_id, name, link, status, created_at, updated_at;

-- name: GetCategoryByID :one
SELECT id, website_id, name, link, status, created_at, updated_at
FROM categories
WHERE id = $1;

-- name: UpdateCategory :one
UPDATE categories
SET website_id = $2, name = $3, link = $4, status = $5, updated_at = CURRENT_TIMESTAMP
WHERE id = $1   
RETURNING id, website_id, name, link, status, created_at, updated_at;

-- name: DeleteCategory :exec
DELETE FROM categories
WHERE id = $1;  

-- name: ListCategoriesByWebsiteID :many
SELECT id, website_id, name, link, status, created_at, updated_at
FROM categories
WHERE website_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveCategoriesByWebsiteID :many
SELECT id, website_id, name, link, status, created_at, updated_at
FROM categories
WHERE website_id = $1 AND status = 'active'
ORDER BY created_at DESC    
LIMIT $2 OFFSET $3;

-- name: ListCategoriesByStatus :many
SELECT id, website_id, name, link, status, created_at, updated_at
FROM categories
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveCategoriesByStatus :many
SELECT id, website_id, name, link, status, created_at, updated_at
FROM categories
WHERE status = $1 AND status = 'active'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllCategories :many
SELECT id, website_id, name, link, status, created_at, updated_at
FROM categories
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAllActiveCategories :many
SELECT id, website_id, name, link, status, created_at, updated_at
FROM categories
WHERE status = 'active'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

