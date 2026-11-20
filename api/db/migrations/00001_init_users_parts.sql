-- +goose Up
CREATE TYPE user_role AS ENUM ('MEMBER', 'MANAGER', 'ADMIN');
CREATE TYPE user_status AS ENUM ('ACTIVE', 'INACTIVE', 'ALUMNI');

CREATE TABLE users (
    id            uuid        PRIMARY KEY DEFAULT uuidv7(),
    student_id    text        NOT NULL UNIQUE,
    password_hash text,                          -- NULL until the account is activated
    name          text        NOT NULL,
    email         text,
    role          user_role   NOT NULL DEFAULT 'MEMBER',
    status        user_status NOT NULL DEFAULT 'ACTIVE',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- Case-insensitive uniqueness; multiple NULL emails are allowed.
CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email));

CREATE TABLE parts (
    id            uuid     PRIMARY KEY DEFAULT uuidv7(),
    name          text     NOT NULL UNIQUE,
    display_order smallint NOT NULL UNIQUE
);

CREATE TABLE user_parts (
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    part_id uuid NOT NULL REFERENCES parts (id) ON DELETE RESTRICT,
    PRIMARY KEY (user_id, part_id)
);

-- PK (user_id, part_id) already serves "parts of a user";
-- this index serves the reverse: "all users in a part".
CREATE INDEX user_parts_part_id_idx ON user_parts (part_id);

-- +goose Down
DROP TABLE user_parts;
DROP TABLE parts;
DROP TABLE users;
DROP TYPE user_status;
DROP TYPE user_role;
