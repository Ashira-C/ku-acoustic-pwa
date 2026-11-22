-- name: CreateUser :one
INSERT INTO users (student_id, name, nickname, email, role, status)
VALUES (@student_id, @name, @nickname, @email, @role, @status)
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

-- name: UpsertImportedUser :one
-- Used only by the one-time importer: re-running updates the same rows.
INSERT INTO users (student_id, name, nickname, role)
VALUES (@student_id, @name, @nickname, @role)
ON CONFLICT (student_id) DO UPDATE
SET name       = EXCLUDED.name,
    nickname   = EXCLUDED.nickname,
    role       = EXCLUDED.role,
    updated_at = now()
RETURNING id;

-- name: DeleteUserParts :exec
DELETE FROM user_parts
WHERE user_id = @user_id;

-- name: CountActivatedUsers :one
SELECT count(*) FROM users
WHERE password_hash IS NOT NULL;