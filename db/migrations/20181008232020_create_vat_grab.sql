-- +goose Up
CREATE TABLE maker.vat_grab
(
    id        SERIAL PRIMARY KEY,
    header_id INTEGER NOT NULL REFERENCES headers (id) ON DELETE CASCADE,
    urn_id    INTEGER NOT NULL REFERENCES maker.urns (id) ON DELETE CASCADE,
    v         TEXT,
    w         TEXT,
    dink      NUMERIC,
    dart      NUMERIC,
    log_idx   INTEGER NOT NULL,
    tx_idx    INTEGER NOT NULL,
    raw_log   JSONB,
    UNIQUE (header_id, tx_idx, log_idx)
);

CREATE INDEX vat_grab_header_index
    ON maker.vat_grab (header_id);

CREATE INDEX vat_grab_urn_index
    ON maker.vat_grab (urn_id);

ALTER TABLE public.checked_headers
    ADD COLUMN vat_grab_checked INTEGER NOT NULL DEFAULT 0;

-- +goose Down
DROP INDEX maker.vat_grab_header_index;
DROP INDEX maker.vat_grab_urn_index;

DROP TABLE maker.vat_grab;

ALTER TABLE public.checked_headers
    DROP COLUMN vat_grab_checked;
