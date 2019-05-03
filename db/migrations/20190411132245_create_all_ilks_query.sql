-- +goose Up

-- Function returning state for all ilks as of the given block height
CREATE OR REPLACE FUNCTION maker.all_ilks(block_height BIGINT)
  RETURNS SETOF maker.ilk_state
AS $$
WITH rates AS (
  SELECT DISTINCT ON (ilk_id) rate, ilk_id, block_hash
  FROM maker.vat_ilk_rate
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), arts AS (
  SELECT DISTINCT ON (ilk_id) art, ilk_id, block_hash
  FROM maker.vat_ilk_art
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), spots AS (
  SELECT DISTINCT ON (ilk_id) spot, ilk_id, block_hash
  FROM maker.vat_ilk_spot
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), lines AS (
  SELECT DISTINCT ON (ilk_id) line, ilk_id, block_hash
  FROM maker.vat_ilk_line
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), dusts AS (
  SELECT DISTINCT ON (ilk_id) dust, ilk_id, block_hash
  FROM maker.vat_ilk_dust
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), chops AS (
  SELECT DISTINCT ON (ilk_id) chop, ilk_id, block_hash
  FROM maker.cat_ilk_chop
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), lumps AS (
  SELECT DISTINCT ON (ilk_id) lump, ilk_id, block_hash
  FROM maker.cat_ilk_lump
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), flips AS (
  SELECT DISTINCT ON (ilk_id) flip, ilk_id, block_hash
  FROM maker.cat_ilk_flip
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), rhos AS (
  SELECT DISTINCT ON (ilk_id) rho, ilk_id, block_hash
  FROM maker.jug_ilk_rho
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
), duties AS (
  SELECT DISTINCT ON (ilk_id) duty, ilk_id, block_hash
  FROM maker.jug_ilk_duty
  WHERE block_number <= $1
  ORDER BY ilk_id, block_number DESC
)
  SELECT
    ilks.id,
    ilks.name,
    $1 block_height,
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
    (
      SELECT TIMESTAMP 'epoch' + h.block_timestamp * INTERVAL '1 second'
      FROM maker.get_ilk_blocks_before($1, ilks.id) b
      JOIN headers h on h.block_number = b.block_height
      ORDER BY h.block_number DESC
      LIMIT 1
    ),
    (
      SELECT TIMESTAMP 'epoch' + h.block_timestamp * INTERVAL '1 second'
      FROM maker.get_ilk_blocks_before($1, ilks.id) b
      JOIN headers h on h.block_number = b.block_height
      ORDER BY h.block_number ASC
      LIMIT 1
    )
  FROM maker.ilks AS ilks
    LEFT JOIN rates on rates.ilk_id = ilks.id
    LEFT JOIN arts on arts.ilk_id = ilks.id
    LEFT JOIN spots on spots.ilk_id = ilks.id
    LEFT JOIN lines on lines.ilk_id = ilks.id
    LEFT JOIN dusts on dusts.ilk_id = ilks.id
    LEFT JOIN chops on chops.ilk_id = ilks.id
    LEFT JOIN lumps on lumps.ilk_id = ilks.id
    LEFT JOIN flips on flips.ilk_id = ilks.id
    LEFT JOIN rhos on rhos.ilk_id = ilks.id
    LEFT JOIN duties on duties.ilk_id = ilks.id
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
    duties.duty is not null
  )
$$
LANGUAGE SQL
STABLE SECURITY DEFINER;

-- +goose Down
DROP FUNCTION IF EXISTS maker.all_ilks(block_height BIGINT);