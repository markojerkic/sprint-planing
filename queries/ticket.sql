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

-- name: EstimateTicket :one
INSERT INTO
  ticket_user_estimate (estimate, user_id, ticket_id)
VALUES
  (?, ?, ?) RETURNING *;
