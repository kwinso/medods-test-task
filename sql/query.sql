-- name: CreateAuth :one
INSERT INTO auths 
  (guid, refresh_token_hash, ip_address, user_agent, refreshed_at)
VALUES 
  ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAuthById :one
SELECT * FROM auths WHERE id = $1;

-- name: DeleteAuthById :exec
DELETE FROM auths WHERE id = $1;