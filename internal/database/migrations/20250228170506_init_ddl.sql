-- +goose Up
-- +goose StatementBegin
CREATE TABLE user (
  id INTEGER PRIMARY KEY,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE session (
  id INTEGER PRIMARY KEY,
  user_id INTEGER REFERENCES user (id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE room (
  id INTEGER PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by INTEGER REFERENCES user (id)
);

CREATE TABLE room_user (
  room_id INTEGER REFERENCES room (id),
  user_id INTEGER REFERENCES user (id),
  PRIMARY KEY (room_id, user_id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE room_user;

DROP TABLE room;

DROP TABLE session;

DROP TABLE user;

-- +goose StatementEnd
