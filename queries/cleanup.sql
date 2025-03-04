-- name: CleanupOldTickets :exec
DELETE FROM ticket_user_estimate
WHERE
  ticket_id IN (
    SELECT
      id
    FROM
      ticket
    WHERE
      closed_at IS NOT NULL
      AND datetime (closed_at) < datetime ('now', '-30 days')
  );

-- name: CleanupClosedTickets :exec
DELETE FROM ticket
WHERE
  closed_at IS NOT NULL
  AND datetime (closed_at) < datetime ('now', '-30 days');

-- name: CleanupUnusedRooms :exec
DELETE FROM room
WHERE
  id NOT IN (
    SELECT DISTINCT
      room_id
    FROM
      ticket
  )
  AND datetime (created_at) < datetime ('now', '-30 days');

-- name: CleanupUnusedUsers :exec
DELETE FROM user
WHERE
  id NOT IN (
    SELECT DISTINCT
      user_id
    FROM
      room_user
  )
  AND datetime (created_at) < datetime ('now', '-30 days');
