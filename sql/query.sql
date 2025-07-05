-- name: CreateAuth :one
INSERT INTO auths 
  (id, guid, refresh_token_hash, ip_address, user_agent, refreshed_at)
VALUES 
  ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAuthById :one
SELECT * FROM auths WHERE id = $1;

-- name: GetAuthByRefreshToken :one
SELECT * FROM auths WHERE refresh_token_hash = $1;

-- name: UpdateAuthRefreshToken :exec
UPDATE auths SET refresh_token_hash = $1, refreshed_at = NOW() WHERE id = $2;

-- name: DeleteAuthById :exec
DELETE FROM auths WHERE id = $1;

-- name: GetNextAuthId :one
SELECT nextval('auths_id_seq');