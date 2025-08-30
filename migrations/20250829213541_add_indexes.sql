-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_events_user_date ON events(user_id, event_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_events_user_date;
-- +goose StatementEnd
