-- name: CreateTicket :one
INSERT INTO
  ticket (name, description, room_id)
VALUES
  (?, ?, ?) RETURNING *;

-- name: GetTicketsOfRoom :many
SELECT
  *
FROM
  ticket
WHERE
  room_id = ?;
