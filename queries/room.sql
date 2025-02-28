-- name: CreateRoom :one
INSERT INTO
  room (name, created_by)
VALUES
  (?, ?) RETURNING *;

-- name: AddUserToRoom :exec
INSERT INTO
  room_user (room_id, user_id)
VALUES  (?, ?);

-- name: GetMyRooms :many
SELECT
    room.id,
    room.name,
    room.created_at,
    room.created_by = :id as is_owner
    FROM room_user
JOIN room ON room.id = room_user.room_id
WHERE room_user.user_id = :id;
