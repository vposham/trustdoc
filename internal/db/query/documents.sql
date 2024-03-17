-- name: GetDoc :one
SELECT *
FROM documents
WHERE doc_id = $1
LIMIT 1;

-- name: AddDoc :one
INSERT INTO documents (doc_id, title, description, file_name, uploaded_by, blockchain_hash)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
