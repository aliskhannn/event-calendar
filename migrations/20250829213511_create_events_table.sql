-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events
(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_date  DATE NOT NULL,
    title       TEXT NOT NULL,
    description TEXT,
    created_at  TIMESTAMP        DEFAULT now(),
    updated_at  TIMESTAMP        DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
