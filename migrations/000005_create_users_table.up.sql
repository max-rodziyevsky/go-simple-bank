CREATE TABLE "users" (
    "username" varchar PRIMARY KEY NOT NULL,
    "full_name" varchar NOT NULL,
    "email" varchar UNIQUE NOT NULL,
    "hash_password" varchar NOT NULL,
    "change_password_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "accounts" ADD FOREIGN KEY ("owner") REFERENCES "users" ("username");

-- CREATE UNIQUE INDEX ON "accounts" ("owner", "currency");
alter table "accounts" add constraint "owner_currency_key" unique ("owner", "currency")