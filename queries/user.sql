-- name: CreateUser :one
INSERT INTO
  user (id)
VALUES
  (DEFAULT) RETURNING *;

-- name: GetUser :one
SELECT
  *
FROM
  user
WHERE
  id = ?;
