-- name: GetDoc :one
SELECT *
FROM documents
WHERE doc_id = $1
LIMIT 1;

-- name: AddDoc :one
INSERT INTO documents (doc_id, title, description, file_name, doc_hash, doc_minted_id, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetDocByHash :one
SELECT *
FROM documents
WHERE doc_hash = $1
LIMIT 1;
