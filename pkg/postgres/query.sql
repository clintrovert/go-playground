-- name: GetUser :one
SELECT * FROM users
WHERE user_id = $1 LIMIT 1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = $1;

-- name: UpdateUser :exec
UPDATE users SET
    name = $1, email = $2, password = $3, modified_at = now()::timestamp
WHERE user_id = $4;

-- name: CreateUser :exec
INSERT INTO users (
    user_id, name, email, password, created_at, modified_at
) VALUES (
    $1, $2, $3, $4, now()::timestamp, now()::timestamp
);

-- name: GetProduct :one
SELECT * FROM products
WHERE product_id = $1 LIMIT 1;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE product_id = $1;

-- name: UpdateProduct :exec
UPDATE products SET
     name = $1, price = $2, modified_at = now()::timestamp
WHERE product_id = $3;