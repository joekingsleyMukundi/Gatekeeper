-- name: CreateEmailVerifyToken :one
INSERT INTO email_verification_tokens (
    username,
    token,
    expires_at 
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetEmailVerifyToken :one
SELECT * FROM email_verification_tokens
WHERE token = $1 LIMIT 1;

-- name: GetActiveEmailVerifyToken :one
SELECT * 
FROM "email_verification_tokens"
WHERE "token" = $1
    AND ("used_at" IS NULL OR "expires_at" < NOW());

-- name: UpdateEmailVerifyToken :exec
UPDATE email_verification_tokens
SET
    used_at= NOW()
WHERE
    token = $1;