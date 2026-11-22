-- +goose Up
-- NOT NULL without a default is safe only because users is still empty everywhere.
ALTER TABLE users
    ADD COLUMN nickname text NOT NULL CHECK (btrim(nickname) <> '');

-- +goose Down
ALTER TABLE users DROP COLUMN nickname;
