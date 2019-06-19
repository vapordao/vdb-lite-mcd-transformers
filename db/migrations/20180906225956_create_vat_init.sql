-- +goose Up
CREATE TABLE maker.vat_init
(
    id        SERIAL PRIMARY KEY,
    header_id INTEGER NOT NULL REFERENCES headers (id) ON DELETE CASCADE,
    ilk_id    INTEGER NOT NULL REFERENCES maker.ilks (id) ON DELETE CASCADE,
    log_idx   INTEGER NOT NULL,
    tx_idx    INTEGER NOT NULL,
    raw_log   JSONB,
    UNIQUE (header_id, tx_idx, log_idx)
);

CREATE INDEX vat_init_header_index
    ON maker.vat_init (header_id);

ALTER TABLE public.checked_headers
    ADD COLUMN vat_init_checked BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
DROP INDEX maker.vat_init_header_index;

DROP TABLE maker.vat_init;

ALTER TABLE public.checked_headers
    DROP COLUMN vat_init_checked;