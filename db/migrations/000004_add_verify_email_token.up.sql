CREATE TABLE "email_verification_tokens" (
    "id" bigserial PRIMARY KEY,
    "username" varchar NOT NULL,
    "token" varchar NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "expires_at" timestamptz NOT NULL,
    "used_at" timestamptz,
    "is_verified" boolean NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX ON "email_verification_tokens" ("username", "token");

CREATE INDEX ON "email_verification_tokens" ("username");

CREATE INDEX ON "email_verification_tokens" ("token");
CREATE INDEX idx_verified_email_tokens
ON email_verification_tokens (username)
WHERE is_verified = true;

ALTER TABLE "email_verification_tokens" ADD FOREIGN KEY ("username") REFERENCES "users" ("username");