-- name: CreateTicket :one
INSERT INTO
  ticket (name, description, room_id)
VALUES
  (?, ?, ?) RETURNING *;

-- name: GetTicketsOfRoom :many
SELECT
    ticket.*,
    ticket_user_estimate_avg.weeks,
    ticket_user_estimate_avg.days,
    ticket_user_estimate_avg.hours,
    ticket_user_estimate.estimate IS NOT NULL as has_estimate,
    ticket_user_estimate.estimate as user_estimate
FROM
  ticket
    LEFT JOIN ticket_user_estimate_avg ON ticket_user_estimate_avg.ticket_id = ticket.id
    LEFT JOIN ticket_user_estimate ON ticket_user_estimate.ticket_id = ticket.id
        AND ticket_user_estimate.user_id = :user_id
WHERE
room_id = :room_id
ORDER BY
  ticket.created_at DESC;

-- name: EstimateTicket :one
INSERT INTO
  ticket_user_estimate (estimate, user_id, ticket_id)
VALUES
  (?, ?, ?) RETURNING *;
