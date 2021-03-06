-- +goose Up
CREATE TYPE api.relevant_block AS (
    block_height BIGINT,
    -- TODO: consider removing hash
    block_hash TEXT,
    block_timestamp NUMERIC,
    ilk_id INTEGER
    );

CREATE FUNCTION api.get_ilk_blocks_before(ilk_identifier TEXT, block_height BIGINT)
    RETURNS SETOF api.relevant_block AS
$$
WITH ilk AS (SELECT id FROM maker.ilks WHERE identifier = ilk_identifier)
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.vat_ilk_rate
         LEFT JOIN public.headers ON vat_ilk_rate.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.vat_ilk_art
         LEFT JOIN public.headers ON vat_ilk_art.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.vat_ilk_spot
         LEFT JOIN public.headers ON vat_ilk_spot.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.vat_ilk_line
         LEFT JOIN public.headers ON vat_ilk_line.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.vat_ilk_dust
         LEFT JOIN public.headers ON vat_ilk_dust.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.cat_ilk_chop
         LEFT JOIN public.headers ON cat_ilk_chop.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.cat_ilk_lump
         LEFT JOIN public.headers ON cat_ilk_lump.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.cat_ilk_flip
         LEFT JOIN public.headers ON cat_ilk_flip.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.jug_ilk_rho
         LEFT JOIN public.headers ON jug_ilk_rho.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
UNION
SELECT block_number AS block_height, hash, block_timestamp, ilk_id
FROM maker.jug_ilk_duty
         LEFT JOIN public.headers ON jug_ilk_duty.header_id = headers.id
WHERE block_number <= get_ilk_blocks_before.block_height
  AND ilk_id = (SELECT id FROM ilk)
ORDER BY block_height DESC
$$
    LANGUAGE sql
    STABLE;

COMMENT ON FUNCTION api.get_ilk_blocks_before(TEXT, BIGINT)
    IS E'@omit';


CREATE TYPE api.ilk_state AS (
    ilk_identifier TEXT,
    block_height BIGINT,
    rate NUMERIC,
    art NUMERIC,
    spot NUMERIC,
    line NUMERIC,
    dust NUMERIC,
    chop NUMERIC,
    lump NUMERIC,
    flip TEXT,
    rho NUMERIC,
    duty NUMERIC,
    pip TEXT,
    mat NUMERIC,
    created TIMESTAMP,
    updated TIMESTAMP
    );

-- Function returning the state for a single ilk as of the given block height
CREATE FUNCTION api.get_ilk(ilk_identifier TEXT, block_height BIGINT DEFAULT api.max_block())
    RETURNS api.ilk_state
AS
$$
WITH ilk AS (SELECT id FROM maker.ilks WHERE identifier = ilk_identifier),
     rates AS (SELECT rate, ilk_id, hash
               FROM maker.vat_ilk_rate
                        LEFT JOIN public.headers ON vat_ilk_rate.header_id = headers.id
               WHERE ilk_id = (SELECT id FROM ilk)
                 AND block_number <= get_ilk.block_height
               ORDER BY ilk_id, block_number DESC
               LIMIT 1),
     arts AS (SELECT art, ilk_id, hash
              FROM maker.vat_ilk_art
                       LEFT JOIN public.headers ON vat_ilk_art.header_id = headers.id
              WHERE ilk_id = (SELECT id FROM ilk)
                AND block_number <= get_ilk.block_height
              ORDER BY ilk_id, block_number DESC
              LIMIT 1),
     spots AS (SELECT spot, ilk_id, hash
               FROM maker.vat_ilk_spot
                        LEFT JOIN public.headers ON vat_ilk_spot.header_id = headers.id
               WHERE ilk_id = (SELECT id FROM ilk)
                 AND block_number <= get_ilk.block_height
               ORDER BY ilk_id, block_number DESC
               LIMIT 1),
     lines AS (SELECT line, ilk_id, hash
               FROM maker.vat_ilk_line
                        LEFT JOIN public.headers ON vat_ilk_line.header_id = headers.id
               WHERE ilk_id = (SELECT id FROM ilk)
                 AND block_number <= get_ilk.block_height
               ORDER BY ilk_id, block_number DESC
               LIMIT 1),
     dusts AS (SELECT dust, ilk_id, hash
               FROM maker.vat_ilk_dust
                        LEFT JOIN public.headers ON vat_ilk_dust.header_id = headers.id
               WHERE ilk_id = (SELECT id FROM ilk)
                 AND block_number <= get_ilk.block_height
               ORDER BY ilk_id, block_number DESC
               LIMIT 1),
     chops AS (SELECT chop, ilk_id, hash
               FROM maker.cat_ilk_chop
                        LEFT JOIN public.headers ON cat_ilk_chop.header_id = headers.id
               WHERE ilk_id = (SELECT id FROM ilk)
                 AND block_number <= get_ilk.block_height
               ORDER BY ilk_id, block_number DESC
               LIMIT 1),
     lumps AS (SELECT lump, ilk_id, hash
               FROM maker.cat_ilk_lump
                        LEFT JOIN public.headers ON cat_ilk_lump.header_id = headers.id
               WHERE ilk_id = (SELECT id FROM ilk)
                 AND block_number <= get_ilk.block_height
               ORDER BY ilk_id, block_number DESC
               LIMIT 1),
     flips AS (SELECT flip, ilk_id, hash
               FROM maker.cat_ilk_flip
                        LEFT JOIN public.headers ON cat_ilk_flip.header_id = headers.id
               WHERE ilk_id = (SELECT id FROM ilk)
                 AND block_number <= get_ilk.block_height
               ORDER BY ilk_id, block_number DESC
               LIMIT 1),
     rhos AS (SELECT rho, ilk_id, hash
              FROM maker.jug_ilk_rho
                       LEFT JOIN public.headers ON jug_ilk_rho.header_id = headers.id
              WHERE ilk_id = (SELECT id FROM ilk)
                AND block_number <= get_ilk.block_height
              ORDER BY ilk_id, block_number DESC
              LIMIT 1),
     duties AS (SELECT duty, ilk_id, hash
                FROM maker.jug_ilk_duty
                         LEFT JOIN public.headers ON jug_ilk_duty.header_id = headers.id
                WHERE ilk_id = (SELECT id FROM ilk)
                  AND block_number <= get_ilk.block_height
                ORDER BY ilk_id, block_number DESC
                LIMIT 1),
     pips AS (SELECT pip, ilk_id, hash
              FROM maker.spot_ilk_pip
                       LEFT JOIN public.headers ON spot_ilk_pip.header_id = headers.id
              WHERE ilk_id = (SELECT id FROM ilk)
                AND block_number <= get_ilk.block_height
              ORDER BY ilk_id, block_number DESC
              LIMIT 1),
     mats AS (SELECT mat, ilk_id, hash
              FROM maker.spot_ilk_mat
                       LEFT JOIN public.headers ON spot_ilk_mat.header_id = headers.id
              WHERE ilk_id = (SELECT id FROM ilk)
                AND block_number <= get_ilk.block_height
              ORDER BY ilk_id, block_number DESC
              LIMIT 1),
     relevant_blocks AS (SELECT * FROM api.get_ilk_blocks_before(ilk_identifier, get_ilk.block_height)),
     created AS (SELECT DISTINCT ON (relevant_blocks.ilk_id,
         relevant_blocks.block_height) relevant_blocks.block_height,
                                       relevant_blocks.block_hash,
                                       relevant_blocks.ilk_id,
                                       api.epoch_to_datetime(relevant_blocks.block_timestamp) AS datetime
                 FROM relevant_blocks
                 ORDER BY relevant_blocks.block_height ASC
                 LIMIT 1),
     updated AS (SELECT DISTINCT ON (relevant_blocks.ilk_id,
         relevant_blocks.block_height) relevant_blocks.block_height,
                                       relevant_blocks.block_hash,
                                       relevant_blocks.ilk_id,
                                       api.epoch_to_datetime(relevant_blocks.block_timestamp) AS datetime
                 FROM relevant_blocks
                 ORDER BY relevant_blocks.block_height DESC
                 LIMIT 1)

SELECT ilks.identifier,
       get_ilk.block_height,
       rates.rate,
       arts.art,
       spots.spot,
       lines.line,
       dusts.dust,
       chops.chop,
       lumps.lump,
       flips.flip,
       rhos.rho,
       duties.duty,
       pips.pip,
       mats.mat,
       created.datetime,
       updated.datetime
FROM maker.ilks AS ilks
         LEFT JOIN rates ON rates.ilk_id = ilks.id
         LEFT JOIN arts ON arts.ilk_id = ilks.id
         LEFT JOIN spots ON spots.ilk_id = ilks.id
         LEFT JOIN lines ON lines.ilk_id = ilks.id
         LEFT JOIN dusts ON dusts.ilk_id = ilks.id
         LEFT JOIN chops ON chops.ilk_id = ilks.id
         LEFT JOIN lumps ON lumps.ilk_id = ilks.id
         LEFT JOIN flips ON flips.ilk_id = ilks.id
         LEFT JOIN rhos ON rhos.ilk_id = ilks.id
         LEFT JOIN duties ON duties.ilk_id = ilks.id
         LEFT JOIN pips ON pips.ilk_id = ilks.id
         LEFT JOIN mats ON mats.ilk_id = ilks.id
         LEFT JOIN created ON created.ilk_id = ilks.id
         LEFT JOIN updated ON updated.ilk_id = ilks.id
WHERE (
              rates.rate is not null OR
              arts.art is not null OR
              spots.spot is not null OR
              lines.line is not null OR
              dusts.dust is not null OR
              chops.chop is not null OR
              lumps.lump is not null OR
              flips.flip is not null OR
              rhos.rho is not null OR
              duties.duty is not null OR
              pips.pip is not null OR
              mats.mat is not null
          )
$$
    LANGUAGE SQL
    STABLE
    STRICT;

-- +goose Down
DROP FUNCTION api.get_ilk_blocks_before(TEXT, BIGINT);
DROP TYPE api.relevant_block CASCADE;
DROP FUNCTION api.get_ilk(TEXT, BIGINT);
DROP TYPE api.ilk_state CASCADE;