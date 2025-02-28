-- name: CreateUser :one
INSERT INTO
  user (id)
VALUES
  (DEFAULT) RETURNING *;

-- name: CreateSession :one
INSERT INTO
  session (user_id)
VALUES
  (?) RETURNING *;
