-- +goose Up
-- +goose StatementBegin
DROP TABLE session;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
  'down SQL query';

-- +goose StatementEnd
