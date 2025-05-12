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
FROM email_verification_tokens
WHERE token = $1
  AND used_at IS NULL
  AND expires_at > NOW()
LIMIT 1;

-- name: UpdateEmailVerifyToken :exec
UPDATE email_verification_tokens
SET
    used_at= NOW(),
    is_verified=true
WHERE
    token = $1;

-- name: IsUserEmailVerified :one
SELECT EXISTS (
    SELECT 1
    FROM email_verification_tokens
    WHERE username = $1
        AND is_verified = true
);