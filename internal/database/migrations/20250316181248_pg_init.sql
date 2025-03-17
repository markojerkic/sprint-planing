-- +goose Up
-- +goose StatementBegin
CREATE TABLE public.user
(
    id         SERIAL
        PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.room
(
    id         SERIAL
        PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER      NOT NULL
        REFERENCES public.user
);

CREATE TABLE public.room_user
(
    room_id INTEGER NOT NULL
        REFERENCES public.room,
    user_id INTEGER NOT NULL
        REFERENCES public.user,
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE public.ticket
(
    id          SERIAL
        PRIMARY KEY,
    name        TEXT                                NOT NULL,
    description TEXT                                NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    room_id     INTEGER                             NOT NULL
        REFERENCES public.room,
    closed_at   TIMESTAMP,
    hidden      BOOLEAN   DEFAULT FALSE             NOT NULL
);

CREATE TABLE public.ticket_user_estimate
(
    id         SERIAL
        PRIMARY KEY,
    estimate   INTEGER                             NOT NULL,
    user_id    INTEGER                             NOT NULL
        REFERENCES public.user,
    ticket_id  INTEGER                             NOT NULL
        REFERENCES public.ticket,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE VIEW public.ticket_estimate_statistics AS
WITH
  estimate_data AS (
    SELECT
      ticket_id,
      estimate
    FROM
      public.ticket_user_estimate
    WHERE
      estimate IS NOT NULL
  )
SELECT
  ticket_id,
  AVG(estimate) AS avg_estimate,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY estimate) AS median_estimate,
  STDDEV(estimate) AS std_dev_estimate
FROM
  estimate_data
GROUP BY
  ticket_id;

CREATE VIEW public.ticket_user_estimate_avg AS
SELECT
    ticket_id,
    AVG(COALESCE(estimate, 0)) AS avg_estimate,
    CAST(AVG(estimate) / 40 AS INTEGER) AS weeks,
    CAST((AVG(estimate) % 40) / 8 AS INTEGER) AS days,
    (AVG(estimate) % 40) % 8 AS hours
FROM
  public.ticket_user_estimate
GROUP BY ticket_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW public.ticket_user_estimate_avg;
DROP VIEW public.ticket_estimate_statistics;
DROP TABLE public.ticket_user_estimate;
DROP TABLE public.ticket;
DROP TABLE public.room_user;
DROP TABLE public.room;
DROP TABLE public.user;
-- +goose StatementEnd
