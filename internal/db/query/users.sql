-- name: GetUser :one
SELECT *
FROM users
WHERE user_id = $1
LIMIT 1;


-- name: AddUser :one
INSERT INTO users (user_id, first_name, last_name, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;
