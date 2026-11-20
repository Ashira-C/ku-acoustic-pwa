-- name: CreateUser :one
INSERT INTO users (student_id, name, email, role, status)
VALUES (@student_id, @name, @email, @role, @status)
RETURNING *;

-- name: GetUserByStudentID :one
SELECT * FROM users
WHERE student_id = @student_id;

-- name: ListParts :many
SELECT * FROM parts
ORDER BY display_order;

-- name: AddUserPart :exec
INSERT INTO user_parts (user_id, part_id)
VALUES (@user_id, @part_id)
ON CONFLICT DO NOTHING;