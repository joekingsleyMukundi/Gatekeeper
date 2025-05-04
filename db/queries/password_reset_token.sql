-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (
    owner,
    token,
    expires_at 
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE token = $1 LIMIT 1;

-- name: GetActivePasswordResetToken :one
SELECT * 
FROM "password_reset_tokens"
WHERE "token" = $1
    AND ("used_at" IS NULL OR "expires_at" < NOW());

-- name: UpdatePasswordResetToken :exec
UPDATE password_reset_tokens
SET
    used_at= NOW()
WHERE
    token = $1;