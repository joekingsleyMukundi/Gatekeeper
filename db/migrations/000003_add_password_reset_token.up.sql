CREATE TABLE "password_reset_tokens" (
    "id" bigserial PRIMARY KEY,
    "owner" varchar NOT NULL,
    "token" varchar NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "expires_at" timestamptz NOT NULL,
    "used_at" timestamptz
);

CREATE INDEX ON "password_reset_tokens" ("owner");

CREATE INDEX ON "password_reset_tokens" ("token");

ALTER TABLE "password_reset_tokens" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");