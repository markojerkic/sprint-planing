-- name: CleanupOldTickets :exec
DELETE FROM public.ticket_user_estimate
WHERE
  ticket_id IN (
    SELECT
      id
    FROM
      public.ticket
    WHERE
      closed_at IS NOT NULL
      AND closed_at < NOW () - INTERVAL '30 days'
  );

-- name: CleanupClosedTickets :exec
DELETE FROM public.ticket
WHERE
  closed_at IS NOT NULL
  AND closed_at < NOW () - INTERVAL '30 days';

-- name: CleanupUnusedRooms :exec
DELETE FROM public.room
WHERE
  id NOT IN (
    SELECT DISTINCT
      room_id
    FROM
      public.ticket
  )
  AND created_at < NOW () - INTERVAL '30 days';

-- name: CleanupUnusedUsers :exec
DELETE FROM public.user
WHERE
  id NOT IN (
    SELECT DISTINCT
      user_id
    FROM
      public.room_user
  )
  AND created_at < NOW () - INTERVAL '30 days';
