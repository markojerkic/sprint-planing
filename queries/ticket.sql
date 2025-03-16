-- name: CreateTicket :one
INSERT INTO
  public.ticket (name, description, room_id)
VALUES
  ($1, $2, $3) RETURNING *;

-- name: GetTicketsOfRoom :many
SELECT
    ticket.*,
    ticket_estimate_statistics.*,
    ticket_user_estimate.estimate IS NOT NULL as has_estimate,
    ticket_user_estimate.estimate as user_estimate,
    (
        SELECT
            COUNT(DISTINCT tue.user_id)
        FROM
            public.ticket_user_estimate tue
        WHERE
            tue.ticket_id = ticket.id
    ) AS users_estimated,
    (
        SELECT
            COUNT(DISTINCT ru.user_id)
        FROM
            public.room_user ru
        WHERE
            ru.room_id = ticket.room_id
    ) AS total_users_in_room
FROM
    public.ticket
LEFT JOIN public.ticket_estimate_statistics ON ticket.id = ticket_estimate_statistics.ticket_id
LEFT JOIN public.ticket_user_estimate ON ticket_user_estimate.ticket_id = ticket.id
    AND ticket_user_estimate.user_id = $1
WHERE
    ticket.room_id = $2
ORDER BY
    ticket.created_at DESC;

-- name: GetTicketAverageEstimation :one
SELECT
  ticket.*,
    (
        SELECT
            COUNT(DISTINCT tue.user_id)
        FROM
            public.ticket_user_estimate tue
        WHERE
            tue.ticket_id = ticket.id
    ) AS users_estimated,
    (
        SELECT
            COUNT(DISTINCT ru.user_id)
        FROM
            public.room_user ru
        WHERE
            ru.room_id = ticket.room_id
    ) AS total_users_in_room,
    ticket_estimate_statistics.*
FROM
  public.ticket
LEFT JOIN public.ticket_estimate_statistics ON ticket.id = ticket_estimate_statistics.ticket_id
WHERE
  ticket.id = $1;

-- name: GetHowManyUsersHaveEstimated :one
SELECT
  (
    SELECT
      COUNT(DISTINCT user_id)
    FROM
      public.ticket_user_estimate
    WHERE
        ticket_id = $1
  ) AS estimated_users,
  (
    SELECT
      COUNT(DISTINCT room_user.user_id)
    FROM
      public.room_user
      JOIN public.ticket ON ticket.room_id = room_user.room_id
    WHERE
      ticket.id = $1
  ) AS total_users;

-- name: EstimateTicket :one
INSERT INTO
  public.ticket_user_estimate (estimate, user_id, ticket_id)
VALUES
  ($1, $2, $3) RETURNING *;

-- name: CloseTicket :one
UPDATE
  public.ticket
SET closed_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetTicketEstimates :many
SELECT
    estimate
FROM
    public.ticket_user_estimate
WHERE
    ticket_id = $1
ORDER BY estimate ASC, created_at DESC;

-- name: ToggleTicketHidden :one
UPDATE
  public.ticket
SET hidden = NOT hidden
WHERE id = $1
RETURNING *;
