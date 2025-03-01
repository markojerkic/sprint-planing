-- +goose Up
-- +goose StatementBegin
CREATE TABLE ticket_user_estimate (
  id INTEGER PRIMARY KEY,
  estimate INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  ticket_id INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES user (id),
  FOREIGN KEY (ticket_id) REFERENCES ticket (id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table ticket_user_estimate;

-- +goose StatementEnd
