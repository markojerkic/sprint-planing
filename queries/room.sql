-- name: CreateRoom :one
INSERT INTO
  room (name, created_by)
VALUES
  (?, ?) RETURNING *;
