-- +goose Up
-- +goose StatementBegin
CREATE TABLE ticket (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  room_id INTEGER NOT NULL,
  FOREIGN KEY (room_id) REFERENCES room (id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table ticket;

-- +goose StatementEnd
