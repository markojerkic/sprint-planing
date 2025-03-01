-- +goose Up
-- +goose StatementBegin
create table ticket (
  id integer constraint ticket_pk primary key autoincrement,
  name strig not null,
  created_at timestamp default CURRENT_TIMESTAMP not null,
  room_id integer not null constraint ticket_room_id_fk references room
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table ticket;

-- +goose StatementEnd
