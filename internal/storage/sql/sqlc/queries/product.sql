-- name: CreateProduct :one
INSERT INTO products (website_id, category_id, title, price, link, image, description, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at;

-- name: GetProductByID :one
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
WHERE id = $1;

-- name: UpdateProduct :one
UPDATE products
SET website_id = $2, category_id = $3, title = $4, price = $5, link = $6, image = $7, description = $8, status = $9, updated_at = CURRENT_TIMESTAMP
WHERE id = $1   
RETURNING id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

-- name: ListProductsByWebsiteID :many
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
WHERE website_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListProductsByCategoryID :many
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
WHERE category_id = $1 
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveProductsByWebsiteID :many
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
WHERE website_id = $1 AND status = 'active'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveProductsByCategoryID :many
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
WHERE category_id = $1 AND status = 'active'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListProductsByStatus :many
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListActiveProductsByStatus :many
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
WHERE status = $1 AND status = 'active'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllProducts :many
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAllActiveProducts :many
SELECT id, website_id, category_id, title, price, link, image, description, status, created_at, updated_at
FROM products
WHERE status = 'active'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

