CREATE TABLE "users" (
    "username" varchar PRIMARY KEY,
    "email" varchar UNIQUE NOT NULL,
    "hashed_password" varchar NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now()),
    "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);