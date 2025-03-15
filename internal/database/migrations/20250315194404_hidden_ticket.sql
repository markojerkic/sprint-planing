-- +goose Up
-- +goose StatementBegin
ALTER TABLE ticket
ADD COLUMN hidden BOOLEAN DEFAULT FALSE NOT NULL;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE ticket
DROP COLUMN hidden;

-- +goose StatementEnd
