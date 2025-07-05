-- name: GetAuthByGUID :one
SELECT * FROM auths WHERE guid = $1;