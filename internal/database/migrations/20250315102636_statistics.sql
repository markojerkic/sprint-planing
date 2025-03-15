-- +goose Up
-- +goose StatementBegin
DROP VIEW IF EXISTS ticket_estimate_statistics;

CREATE VIEW ticket_estimate_statistics AS
WITH
  estimate_data AS (
    SELECT
      ticket_id,
      estimate
    FROM
      ticket_user_estimate
    WHERE
      estimate IS NOT NULL
  ),
  median_calc AS (
    SELECT
      ticket_id,
      estimate,
      ROW_NUMBER() OVER (
        PARTITION BY
          ticket_id
        ORDER BY
          estimate
      ) AS row_num,
      COUNT(*) OVER (
        PARTITION BY
          ticket_id
      ) AS total_rows
    FROM
      estimate_data
  ),
  avg_calc AS (
    SELECT
      ticket_id,
      AVG(estimate) as avg_estimate
    FROM
      estimate_data
    GROUP BY
      ticket_id
  )
SELECT
  e.ticket_id,
  a.avg_estimate,
  (
    SELECT
      AVG(m.estimate)
    FROM
      median_calc m
    WHERE
      m.ticket_id = e.ticket_id
      AND m.row_num BETWEEN (m.total_rows + 1) / 2 AND (m.total_rows + 2)  / 2
  ) as median_estimate,
  CAST(
    SQRT(
      AVG(
        (e.estimate - a.avg_estimate) * (e.estimate - a.avg_estimate)
      )
    ) AS REAL
  ) as std_dev_estimate
FROM
  estimate_data e
  JOIN avg_calc a ON e.ticket_id = a.ticket_id
GROUP BY
  e.ticket_id;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS ticket_estimate_statistics;

-- +goose StatementEnd
