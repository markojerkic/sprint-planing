-- name: CreateTicket :one
INSERT INTO
  ticket (name, description, room_id)
VALUES
  (?, ?, ?) RETURNING *;

-- name: GetTicketsOfRoom :many
SELECT
    ticket.*,
    ticket_user_estimate_avg.avg_estimate,
    ticket_user_estimate.estimate IS NOT NULL as has_estimate,
    ticket_user_estimate.estimate as user_estimate,
    (
        SELECT
            COUNT(DISTINCT tue.user_id)
        FROM
            ticket_user_estimate tue
        WHERE
            tue.ticket_id = ticket.id
    ) AS users_estimated,
    (
        SELECT
            COUNT(DISTINCT ru.user_id)
        FROM
            room_user ru
        WHERE
            ru.room_id = ticket.room_id
    ) AS total_users_in_room
FROM
    ticket
    LEFT JOIN ticket_user_estimate_avg ON ticket_user_estimate_avg.ticket_id = ticket.id
    LEFT JOIN ticket_user_estimate ON ticket_user_estimate.ticket_id = ticket.id
        AND ticket_user_estimate.user_id = :user_id
WHERE
    ticket.room_id = :room_id
ORDER BY
    ticket.created_at DESC;

-- name: GetTicketAverageEstimation :one
select
  ticket.*,
    ticket_user_estimate_avg.avg_estimate,
    (
        SELECT
            COUNT(DISTINCT tue.user_id)
        FROM
            ticket_user_estimate tue
        WHERE
            tue.ticket_id = ticket.id
    ) AS users_estimated,
    (
        SELECT
            COUNT(DISTINCT ru.user_id)
        FROM
            room_user ru
        WHERE
            ru.room_id = ticket.room_id
    ) AS total_users_in_room
from
  ticket
  left join ticket_user_estimate_avg on ticket.id = ticket_user_estimate_avg.ticket_id
where
  ticket.id = :ticket_id;

-- name: GetHowManyUsersHaveEstimated :one
SELECT
  (
    SELECT
      COUNT(DISTINCT user_id)
    FROM
      ticket_user_estimate
    WHERE
        ticket_id = :ticket_id
  ) AS estimated_users,
  (
    SELECT
      COUNT(DISTINCT room_user.user_id)
    FROM
      room_user
      JOIN ticket ON ticket.room_id = room_user.room_id
    WHERE
      ticket.id = :ticket_id
  ) AS total_users;

-- name: EstimateTicket :one
INSERT INTO
  ticket_user_estimate (estimate, user_id, ticket_id)
VALUES
  (?, ?, ?) RETURNING *;


-- name: CloseTicket :one
UPDATE
  ticket SET closed_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;
