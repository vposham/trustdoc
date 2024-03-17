-- name: GetUser :one
SELECT *
FROM users
WHERE email_id = $1
LIMIT 1;


-- name: AddUser :one
INSERT INTO users (email_id, first_name, last_name, status)
VALUES ($1, $2, $3, $4)
RETURNING *;
