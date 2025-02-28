-- +goose Up
-- +goose StatementBegin
CREATE TABLE user (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE session (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES user (id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE room (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by INT REFERENCES user (id)
);

CREATE TABLE room_user (
  room_id INT REFERENCES room (id),
  user_id INT REFERENCES user (id),
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
