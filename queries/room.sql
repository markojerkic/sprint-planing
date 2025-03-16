-- name: CreateRoom :one
INSERT INTO
  public.room (name, created_by)
VALUES
  ($1, $2) RETURNING *;

-- name: GetMyRooms :many
SELECT
    room.id,
    room.name,
    room.created_at,
    room.created_by = $1 as is_owner
FROM public.room_user
JOIN public.room ON room.id = room_user.room_id
WHERE room_user.user_id = $1
ORDER BY room.created_at DESC;

-- name: GetRoomDetails :one
SELECT room.*,
       room.created_by = $1 as is_owner
FROM public.room
WHERE room.id = $2;

-- name: AddUserToRoom :exec
INSERT INTO
  public.room_user (room_id, user_id)
VALUES ($1, $2)
ON CONFLICT (room_id, user_id) DO NOTHING;
