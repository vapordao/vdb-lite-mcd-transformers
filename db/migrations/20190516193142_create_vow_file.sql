-- +goose Up
CREATE TABLE maker.vow_file
(
    id        SERIAL PRIMARY KEY,
    header_id INTEGER NOT NULL REFERENCES headers (id) ON DELETE CASCADE,
    what      TEXT,
    data      NUMERIC,
    log_idx   INTEGER NOT NULL,
    tx_idx    INTEGER NOT NULL,
    raw_log   JSONB,
    UNIQUE (header_id, tx_idx, log_idx)
);

ALTER TABLE public.checked_headers
    ADD COLUMN vow_file_checked INTEGER NOT NULL DEFAULT 0;

-- +goose Down
DROP TABLE maker.vow_file;

ALTER TABLE public.checked_headers
    DROP COLUMN vow_file_checked;
