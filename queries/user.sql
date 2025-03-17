-- name: CreateUser :one
INSERT INTO
  public.user (created_at)
VALUES
  (NOW()) RETURNING *;

-- name: GetUser :one
SELECT
  *
FROM
  public.user
WHERE
  id = $1;
