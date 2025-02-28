-- name: CreateUser :one
INSERT INTO
  user (created_at)
VALUES
  (CURRENT_TIMESTAMP) RETURNING *;

-- name: GetUser :one
SELECT
  *
FROM
  user
WHERE
  id = ?;
