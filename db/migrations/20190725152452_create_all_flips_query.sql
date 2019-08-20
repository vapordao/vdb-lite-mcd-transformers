-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE OR REPLACE FUNCTION api.all_flips(ilk TEXT) RETURNS SETOF api.flip_state AS
-- +goose StatementBegin
$BODY$
BEGIN
    RETURN QUERY (
        WITH ilk_ids AS (SELECT id
                        FROM maker.ilks
                        WHERE identifier = all_flips.ilk),
             address AS (
                 SELECT DISTINCT contract_address
                 FROM maker.flip_ilk
                 WHERE flip_ilk.ilk_id = (SELECT id FROM ilk_ids)
                 LIMIT 1),
             bid_ids AS (
                 SELECT DISTINCT flip_kicks.kicks
                 FROM maker.flip_kicks
                 WHERE contract_address = (SELECT * FROM address)
                 ORDER BY flip_kicks.kicks)
        SELECT f.*
        FROM bid_ids,
             LATERAL api.get_flip(bid_ids.kicks, all_flips.ilk) f
    );
END
$BODY$
    LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION api.all_flips(ilk TEXT);