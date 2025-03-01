-- +goose Up
-- +goose StatementBegin
CREATE VIEW ticket_user_estimate_avg AS
SELECT
    ticket_id,
    AVG(COALESCE(estimate, 0)) as avg_estimate,
    CAST(AVG(estimate) / 40 AS INTEGER) as weeks,
    CAST((AVG(estimate) % 40) / 8 AS INTEGER) as days,
    (AVG(estimate) % 40) % 8 as hours
FROM
  ticket_user_estimate
GROUP BY ticket_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW ticket_user_estimate_avg;
-- +goose StatementEnd
