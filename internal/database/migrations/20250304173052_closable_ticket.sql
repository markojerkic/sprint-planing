-- +goose Up
-- +goose StatementBegin
ALTER TABLE ticket
ADD COLUMN closed_at TIMESTAMP;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE ticket
DROP COLUMN closed_at;

-- +goose StatementEnd
